package dcos_statsd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/dcos_statsd/api"
	"github.com/influxdata/telegraf/plugins/inputs/dcos_statsd/containers"
	"github.com/influxdata/telegraf/plugins/inputs/statsd"
)

const sampleConfig = `
## The address on which the command API should listen
listen = ":8888"
## The directory in which container information is stored
containers_dir = "/run/dcos/telegraf/dcos_statsd/containers"
## The period after which requests to the API should time out
timeout = "15s"
## The hostname or IP address on which to host statsd servers
statsd_host = "198.51.100.1"
`

type DCOSStatsd struct {
	// Listen is the address on which the command API listens
	Listen string
	// ContainersDir is the directory in which container information is stored
	ContainersDir string
	Timeout       internal.Duration
	StatsdHost    string
	apiServer     *http.Server
	containers    map[string]containers.Container
}

// SampleConfig returns the default configuration
func (ds *DCOSStatsd) SampleConfig() string {
	return sampleConfig
}

// Description returns a one-sentence description of dcos_statsd
func (ds *DCOSStatsd) Description() string {
	return "Plugin for monitoring statsd metrics from mesos tasks"
}

// Start is called when the service plugin is ready to start working
func (ds *DCOSStatsd) Start(acc telegraf.Accumulator) error {
	// if ds.containers was not properly initiated, we can have issues with
	// assignment to null map
	if ds.containers == nil {
		ds.containers = map[string]containers.Container{}
	}
	router := api.NewRouter(ds)
	ds.apiServer = &http.Server{
		Handler:      router,
		Addr:         ds.Listen,
		WriteTimeout: ds.Timeout.Duration,
		ReadTimeout:  ds.Timeout.Duration,
	}

	if ds.ContainersDir != "" {
		// Check that dir exists
		if _, err := os.Stat(ds.ContainersDir); os.IsNotExist(err) {
			log.Printf("I! %s does not exist and will be created now", ds.ContainersDir)
			os.MkdirAll(ds.ContainersDir, 0666)
		}
		// We fail early if something is up with the containers dir
		// (eg bad permissions)
		if err := ds.loadContainers(); err != nil {
			return err
		}
	} else {
		// We set ContainersDir in init(). If it's not set, it's either been
		// explicitly unset, or we're inside a test
		log.Println("I! No containers_dir was set; state will not persist")
	}

	go func() {
		err := ds.apiServer.ListenAndServe()
		log.Printf("I! dcos_statsd API server closed: %s", err)
	}()
	log.Printf("I! dcos_statsd API server listening on %s", ds.Listen)

	return nil
}

// Gather takes in an accumulator and adds the metrics that the plugin gathers.
// It is invoked on a schedule (default every 10s) by the telegraf runtime.
func (ds *DCOSStatsd) Gather(_ telegraf.Accumulator) error {
	// TODO instantiate a custom accumulator for each plugin
	// TODO wait for all plugins to accumulate their metrics
	// TODO add container_id tags to each metric
	// TODO pass all metrics into the telegraf accumulator
	return nil
}

// Stop is called when the service plugin needs to stop working
func (ds *DCOSStatsd) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), ds.Timeout.Duration)
	defer cancel()
	ds.apiServer.Shutdown(ctx)

	for _, c := range ds.containers {
		c.Server.Stop()
	}
}

// ListContainers returns a list of known containers
func (ds *DCOSStatsd) ListContainers() []containers.Container {
	ctrs := []containers.Container{}
	for _, c := range ds.containers {
		ctrs = append(ctrs, c)
	}
	return ctrs
}

// GetContainer returns a container from its ID, and whether it was successful
func (ds *DCOSStatsd) GetContainer(cid string) (*containers.Container, bool) {
	ctr, ok := ds.containers[cid]
	return &ctr, ok
}

// AddContainer takes a container definition and adds a container, if one does
// not exist with the same ID. If the statsd_host and statsd_port fields are
// defined, it will attempt to start a server on the defined address. If this
// fails, it will error and the container will not be added. If the fields are
// not defined, it wil attempt to start a server on a random port and the
// default host. If this fails, it will error and the container will not be
// added. If the operation was successful, it will return the container.
func (ds *DCOSStatsd) AddContainer(ctr containers.Container) (*containers.Container, error) {
	ctr.Server = &statsd.Statsd{
		Protocol:               "udp",
		ServiceAddress:         fmt.Sprintf(":%d", ctr.StatsdPort),
		ParseDataDogTags:       true,
		AllowedPendingMessages: 10000,
	}

	// statsd will crash the whole Telegraf process if it attempts to listen on
	// an occupied port. We therefore check ports in advance if specified by the
	// user.
	if ctr.StatsdPort != 0 && !checkPort(ctr.StatsdPort) {
		log.Printf("E! Attempted to start a server on an occupied port: %d", ctr.StatsdPort)
		return nil, fmt.Errorf("could not start server on occupied port %d", ctr.StatsdPort)
	}

	// Statsd.Start discards its accumulator
	var acc telegraf.Accumulator
	if err := ctr.Server.Start(acc); err != nil {
		log.Printf("E! Could not start server for container %s", ctr.Id)
		return nil, err
	}
	log.Printf("I! Added container %s", ctr.Id)

	if ctr.StatsdHost == "" {
		ctr.StatsdHost = ds.StatsdHost
	}

	if ctr.StatsdPort == 0 {
		port, err := getStatsdServerPort(ctr.Server)
		if err != nil {
			log.Printf("E! Could not find port for container %s: %s", ctr.Id, err)
			return nil, err
		}
		ctr.StatsdPort = port
	}

	ds.containers[ctr.Id] = ctr
	// TODO persist container to disk
	return &ctr, nil
}

// Remove container will remove a container and stop any associated server. the
// host and port need not be present in the container argument.
func (ds *DCOSStatsd) RemoveContainer(c containers.Container) error {
	return nil
}

// loadContainers loads containers from disk
func (ds *DCOSStatsd) loadContainers() error {
	files, err := ioutil.ReadDir(ds.ContainersDir)
	if err != nil {
		log.Printf("E! The specified containers dir was not available: %s", err)
		return err
	}

	for _, fInfo := range files {
		// No need for filepath.Join - this simple concat works on Windows
		fPath := fmt.Sprintf("%s/%s", ds.ContainersDir, fInfo.Name())

		// Attempt to open file
		file, err := os.Open(fPath)
		if err != nil {
			log.Printf("E! The specified file %s could not be opened: %s", fPath, err)
			continue
		}
		defer file.Close()

		// Consume file as JSON
		var ctr containers.Container
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&ctr); err != nil {
			log.Printf("E! The container file %s could not be decoded: %s", fPath, err)
			continue
		}

		// Finally, add container to cache
		if _, err := ds.AddContainer(ctr); err != nil {
			log.Printf("E! Could not add container %s: %s", ctr.Id, err)
			continue
		}
		log.Printf("I! Loaded container %s from disk", ctr.Id)
	}
	return nil
}

// getStatsdServerPort waits for the statsd server to start up, then returns
// the port on which it is running, or times out.
func getStatsdServerPort(s *statsd.Statsd) (int, error) {
	done := make(chan bool)
	go func() {
		for {
			if s.UDPlistener != nil {
				done <- true
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	select {
	case <-time.After(time.Second):
		return 0, errors.New("timed out waiting for statsd server to start")
	case <-done:
		break
	}

	addr := s.UDPlistener.LocalAddr().String()
	su, err := url.Parse("http://" + addr)

	if err != nil {
		return 0, err
	}

	return strconv.Atoi(su.Port())
}

// checkPort checks that a port is free.
// statsd.listenUDP will throw Fatal if it attempts to listen on a port which
// was already bound. As we cannot guarantee that a port is always free, since
// other processes are running on our machines, we need to check ahead of time.
func checkPort(port int) bool {
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	ln, err := net.ListenUDP("udp", addr)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func init() {
	inputs.Add("dcos_statsd", func() telegraf.Input {
		return &DCOSStatsd{
			ContainersDir: "/run/dcos/telegraf/dcos_statsd/containers",
			Timeout:       internal.Duration{Duration: 10 * time.Second},
			StatsdHost:    "198.51.100.1",
			containers:    map[string]containers.Container{},
		}
	})
}
