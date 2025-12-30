package main

import (
	"imageProcessor/internal/config"
	"imageProcessor/internal/handlers"
	"imageProcessor/internal/storage/sqlite"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// TODO: add cfg creator
	cfg := config.Config{
		StoragePath: "storage/storage.db",
	}
	//
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// storage
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		panic(err)
	}

	// TODO:
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/upload", handlers.UploadImage(logger, storage))
	router.Group(func(r chi.Router) {
		r.Get("/image/{id}", handlers.DownloadImage(logger, storage))
		r.Delete("/image/{id}", handlers.DeleteImage(logger, storage))
	})

	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}
