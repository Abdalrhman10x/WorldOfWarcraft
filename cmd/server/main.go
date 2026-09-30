package main

import (
	"log/slog"
	"net/http"
	"os"

	"WorldOfWarcraft/internal/api"
	"WorldOfWarcraft/internal/store"
)

const version = "0.1.0"

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	srv := api.New(store.NewMemory())

	addr := ":8080"
	slog.Info("starting", "version", version, "addr", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
