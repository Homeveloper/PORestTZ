package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Homeveloper/PORestTZ/internal/config"
	"github.com/Homeveloper/PORestTZ/internal/httpapi"
	"github.com/Homeveloper/PORestTZ/internal/migrations"
	"github.com/Homeveloper/PORestTZ/internal/repository/postgres"
	"github.com/Homeveloper/PORestTZ/internal/service"
)

func main() {
	cfg, err := config.Load("config.env")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, cfg.DSN)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	schemaBytes, err := migrations.Schema.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("failed to read schema: %v", err)
	}
	if _, err := pool.Exec(ctx, string(schemaBytes)); err != nil {
		log.Fatalf("failed to apply schema: %v", err)
	}
	log.Println("schema applied")

	repo := postgres.NewWalletRepo(pool)
	svc := service.NewWalletService(repo)
	handler := httpapi.NewHandler(svc)
	router := httpapi.NewRouter(handler)

	addr := ":" + cfg.HTTPPort
	log.Println("starting server on", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}