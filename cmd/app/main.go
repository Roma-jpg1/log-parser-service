package main

import (
	"awesomeProject/internal/api"
	"awesomeProject/internal/config"
	"awesomeProject/internal/db"
	"log"
	"net/http"
)

func main() {
	cfg := config.Load()

	database := db.Connect(cfg.DatabaseURL)
	defer database.Close()

	if err := database.Ping(); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}

	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("run migrations failed: %v", err)
	}

	repo := db.NewRepository(database)
	router := api.NewRouter(repo)

	log.Printf("Server started on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}
