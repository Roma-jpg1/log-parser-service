package main

import (
	"awesomeProject/internal/api"
	"awesomeProject/internal/config"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/health", api.HealthHandler)

	log.Printf("Server started on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
