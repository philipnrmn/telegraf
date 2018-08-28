package api

import (
	"encoding/json"
	"log"
	"net/http"
)

func ContainerInfo(_ ServerController) http.HandlerFunc {
	// TODO health endpoint

	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("dcos_statsd::ContainerInfo::anon")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		// TODO nice to have
	}
}

func DescribeContainer(_ ServerController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("dcos_statsd::DescribeContainers::anon")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		// TODO nice to have
	}
}

func ListContainers(_ ServerController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("dcos_statsd::ListContainers::anon")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		// TODO nice to have
	}
}

func StartServer(sc ServerController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("dcos_statsd::StartServer::anon")
		// TODO detect whether the server actually needs to be started
		var c Container
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&c); err != nil {
			// TODO error
			return
		}
		if err := sc.Start(&c); err != nil {
			// TODO log error
			log.Printf("error while creating server: %s", err)
		}
		// TODO read fulfilled container back from start
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusCreated)

		// TODO catch error
		data, _ := json.Marshal(c)
		w.Write(data)

	}
}

func StopServer(sc ServerController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("dcos_statsd::StopServer::anon")
		// var id string
		// TODO write container to stop
		// TODO read fulfilled container back from stop
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusAccepted)
	}
}
