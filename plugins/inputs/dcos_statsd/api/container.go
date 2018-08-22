package api

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/statsd"
)

type Container struct {
	Id string `json:"container_id"`

	StatsdHost string `json:"statsd_host,omitempty"`

	StatsdPort int `json:"statsd_port,omitempty"`

	Server statsd.Statsd `json:"-"`
}

type ServerController interface {
	Start(*Container) error
	Stop(*Container) error
	Gather(telegraf.Accumulator)
}

type StatsdServerController struct {
	containers map[string]Container
}

func NewStatsdServerController() *StatsdServerController {
	return &StatsdServerController{
		containers: map[string]Container{},
	}
}

func (sc *StatsdServerController) Start(c *Container) error {
	if _, ok := sc.containers[c.Id]; ok {
		return fmt.Errorf("container with id %s already exists", c.Id)
	}

	server := statsd.Statsd{
		Protocol:         "udp",
		ServiceAddress:   ":0",
		ParseDataDogTags: true,
	}
	var acc telegraf.Accumulator
	server.Start(acc)

	// TODO wait for this properly
	time.Sleep(1 * time.Second)
	addr := server.UDPlistener.LocalAddr().String()
	// TODO handle error
	url, err := url.Parse("http://" + addr)
	if err != nil {
		return fmt.Errorf("failed to parse statsd server address %s", addr)
	}

	c.StatsdHost = "198.51.100.1"
	c.StatsdPort, _ = strconv.Atoi(url.Port())
	sc.containers[c.Id] = *c
	return nil
}

func (sc *StatsdServerController) Gather(acc telegraf.Accumulator) {
	var wg sync.WaitGroup
	for id, c := range sc.containers {
		log.Printf("StatsdServerController::Gather::%s", id)
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Server.Gather(acc)
		}()
	}
	wg.Wait()

}

func (sc *StatsdServerController) Stop(cc *Container) error {
	c, ok := sc.containers[cc.Id]
	if !ok {
		return fmt.Errorf("container with id %s does not exist", c.Id)
	}

	c.Server.Stop()
	delete(sc.containers, c.Id)

	return nil
}
