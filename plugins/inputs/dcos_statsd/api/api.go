package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/influxdata/telegraf/plugins/inputs/dcos_statsd"
)

func ReportHealth(_ dcos_statsd.DCOSStatsd) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		// TODO report health
	}
}

func ListContainers(_ dcos_statsd.DCOSStatsd) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		// TODO list containers
	}
}

func DescribeContainer(_ dcos_statsd.DCOSStatsd) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		// TODO describe container
	}
}

func AddContainer(_ dcos_statsd.DCOSStatsd) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c Container
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&c); err != nil {
			log.Printf("E! could not decode json: %s", err)
			return
		}
		// TODO start server
		// TODO write container definition to disc

		data, err := json.Marshal(c)
		if err != nil {
			log.Printf("E! could not encode json: %s", err)
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusCreated)
		w.Write(data)

	}
}

func RemoveContainer(_ dcos_statsd.DCOSStatsd) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusAccepted)
		// TODO stop server
		// TODO delete container definition from disc
		// TODO remove container
	}
}
