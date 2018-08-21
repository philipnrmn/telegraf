package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc func(chan Container, chan Container) http.HandlerFunc
}

type Routes []Route

func NewRouter(start chan Container, stop chan Container) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method(start, stop)).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		Index,
	},

	Route{
		"ContainerInfo",
		strings.ToUpper("Get"),
		"/container",
		ContainerInfo,
	},

	Route{
		"DescribeContainer",
		strings.ToUpper("Get"),
		"/container/{id}",
		DescribeContainer,
	},

	Route{
		"ListContainers",
		strings.ToUpper("Get"),
		"/containers",
		ListContainers,
	},

	Route{
		"StartServer",
		strings.ToUpper("Post"),
		"/container",
		StartServer,
	},

	Route{
		"StopServer",
		strings.ToUpper("Delete"),
		"/container/{id}",
		StopServer,
	},
}
