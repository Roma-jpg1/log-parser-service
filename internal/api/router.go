package api

import (
	"net/http"
)

func NewRouter() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("/health", HealthHandler)
	router.HandleFunc("/api/v1/node/", NodeHandler)
	router.HandleFunc("/api/v1/parse/", ParseHandler)
	router.HandleFunc("/api/v1/topology/", TopologyHandler)
	router.HandleFunc("/api/v1/log/", LogHandler)

	return router

}
