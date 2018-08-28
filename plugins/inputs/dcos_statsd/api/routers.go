package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf/plugins/inputs/dcos_statsd"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc func(ds dcos_statsd.DCOSStatsd) http.HandlerFunc
}

type Routes []Route

func NewRouter(ds dcos_statsd.DCOSStatsd) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc(ds)
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

func Index(ds dcos_statsd.DCOSStatsd) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Not Found")
	}
}

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		Index,
	},

	Route{
		"ListContainers",
		strings.ToUpper("Get"),
		"/containers",
		ListContainers,
	},

	Route{
		"DescribeContainer",
		strings.ToUpper("Get"),
		"/container/{id}",
		DescribeContainer,
	},

	Route{
		"AddContainer",
		strings.ToUpper("Post"),
		"/container",
		AddContainer,
	},

	Route{
		"RemoveCOntainer",
		strings.ToUpper("Delete"),
		"/container/{id}",
		RemoveContainer,
	},
}
