package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"link-shortener/internal/api"
	"link-shortener/internal/storage/sqlite"
)

const defaultAddr = ":8080"
const defaultBaseURL = "http://localhost:8080"
const defaultDBPath = "data.db"

func main() {
	store, err := sqlite.New(dbPath())
	if err != nil {
		log.Fatalf("failed to initialize sqlite store: %v", err)
	}
	server := api.NewServer(api.Config{
		Store:   store,
		BaseURL: baseURL(),
	})

	addr := listenAddr()
	log.Printf("server listening on %s", addr)

	if err := http.ListenAndServe(addr, server.Routes()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func listenAddr() string {
	if val := strings.TrimSpace(os.Getenv("LISTEN_ADDR")); val != "" {
		return val
	}
	return defaultAddr
}

func baseURL() string {
	if val := strings.TrimSpace(os.Getenv("BASE_URL")); val != "" {
		return strings.TrimSuffix(val, "/")
	}
	return defaultBaseURL
}

func dbPath() string {
	if val := strings.TrimSpace(os.Getenv("SQLITE_PATH")); val != "" {
		return val
	}
	return defaultDBPath
}
