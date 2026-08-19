package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/podcctv/detective-chicken/internal/server"
	"github.com/podcctv/detective-chicken/internal/store"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-healthcheck" {
		client := http.Client{Timeout: 2 * time.Second}
		res, err := client.Get("http://127.0.0.1:8080/api/v1/health")
		if err != nil || res.StatusCode != http.StatusOK {
			fmt.Fprintln(os.Stderr, "healthcheck failed")
			os.Exit(1)
		}
		_ = res.Body.Close()
		return
	}
	addr := os.Getenv("DETECTIVE_CHICKEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	api := server.New(store.NewMemory(true), logger)
	logger.Info("Detective Chicken API listening", "addr", addr)
	if err := http.ListenAndServe(addr, api.Handler()); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
