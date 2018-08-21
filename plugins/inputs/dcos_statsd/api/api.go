package api

import (
	"encoding/json"
	"log"
	"net/http"
)

func ContainerInfo(_ chan Container, _ chan Container) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		// TODO nice to have
	}
}

func DescribeContainer(_ chan Container, _ chan Container) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		// TODO nice to have
	}
}

func ListContainers(_ chan Container, _ chan Container) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		// TODO nice to have
	}
}

func StartServer(start chan Container, _ chan Container) {
	return func(w http.ResponseWriter, r *http.Request) {
		var c Container
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&c); err != nil {
			// TODO error
			return
		}
		log.Printf("Container: %s", c)
		// TODO write container to start
		// TODO read fulfilled container back from start
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
	}
}

func StopServer(_ chan Container, stop chan Container) {
	return func(w http.ResponseWriter, r *http.Request) {
		// var id string
		// TODO write container to stop
		// TODO read fulfilled container back from stop
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
	}
}
