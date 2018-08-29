package dcos_statsd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
`

type DCOSStatsd struct {
	// Listen is the address on which the command API listens
	Listen string
	// ContainersDir is the directory in which container information is stored
	ContainersDir string
	Timeout       internal.Duration
	apiServer     *http.Server
	containers    []containers.Container
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
		if err := ds.loadContainers(acc); err != nil {
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
	// TODO stop servers
}

// ListContainers returns a list of known containers
func (ds *DCOSStatsd) ListContainers() []containers.Container {
	return ds.containers
}

// GetContainer returns a container from its ID, and whether it was successful
func (ds *DCOSStatsd) GetContainer(cid string) (*containers.Container, bool) {
	return nil, false
}

// AddContainer takes a container definition and adds a container, if one does
// not exist with the same ID. If the statsd_host and statsd_port fields are
// defined, it will attempt to start a server on the defined address. If this
// fails, it will error and the container will not be added. If the fields are
// not defined, it wil attempt to start a server on a random port and the
// default host. If this fails, it will error and the container will not be
// added. If the operation was successful, it will return the container.
func (ds *DCOSStatsd) AddContainer(c containers.Container) (*containers.Container, error) {
	return nil, nil
}

// Remove container will remove a container and stop any associated server. the
// host and port need not be present in the container argument.
func (ds *DCOSStatsd) RemoveContainer(c containers.Container) error {
	return nil
}

// loadContainers loads containers from disk
func (ds *DCOSStatsd) loadContainers(acc telegraf.Accumulator) error {
	files, err := ioutil.ReadDir(ds.ContainersDir)
	if err != nil {
		log.Printf("E! The specified containers dir was not available: %s", err)
		return err
	}
	for _, fInfo := range files {
		fPath := fmt.Sprintf("%s/%s", ds.ContainersDir, fInfo.Name())
		file, err := os.Open(fPath)
		if err != nil {
			log.Printf("E! The specified file %s could not be opened: %s", fPath, err)
			continue
		}
		defer file.Close()
		var ctr containers.Container
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&ctr); err != nil {
			log.Printf("E! The container file %s could not be decoded: %s", fPath, err)
			continue
		}

		ctr.Server = &statsd.Statsd{
			Protocol:               "udp",
			ServiceAddress:         fmt.Sprintf(":%d", ctr.StatsdPort),
			ParseDataDogTags:       true,
			AllowedPendingMessages: 10000,
		}

		err = ctr.Server.Start(acc)
		if err != nil {
			log.Printf("E! Could not start server for container %s", ctr.Id)
			continue
		}
		log.Printf("I! Loaded container %s from disk", ctr.Id)
		ds.containers = append(ds.containers, ctr)
	}
	return nil
}

func init() {
	inputs.Add("dcos_statsd", func() telegraf.Input {
		return &DCOSStatsd{
			ContainersDir: "/run/dcos/telegraf/dcos_statsd/containers",
			Timeout:       internal.Duration{Duration: 10 * time.Second},
			containers:    []containers.Container{},
		}
	})
}
