package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = LogApi(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

var routes = Routes{
	Route{
		"GetStatus",
		"GET",
		"/status",
		GetStatus,
	},
	Route{
		"Quiesce",
		"POST",
		"/quiesce/{backupId}",
		Quiesce,
	},
	Route{
		"UnQuiesce",
		"POST",
		"/unquiesce/{backupId}",
		UnQuiesce,
	},
	Route{
		"Backup",
		"POST",
		"/backup/{backupId}",
		Backup,
	},
}
