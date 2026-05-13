package api

import (
	"awesomeProject/internal/db"
	"net/http"
)

func NewRouter(repo *db.Repository) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(repo)

	router.HandleFunc("/health", HealthHandler)
	router.HandleFunc("/api/v1/node/", handler.NodeHandler)
	router.HandleFunc("/api/v1/port/", handler.PortHandler)
	router.HandleFunc("/api/v1/parse/", handler.ParseHandler)
	router.HandleFunc("/api/v1/topology/", handler.TopologyHandler)
	router.HandleFunc("/api/v1/log/", handler.LogHandler)

	return router
}
