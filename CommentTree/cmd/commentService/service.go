package main

import (
	"CommentTree/internal/http-server/handlers"
	"CommentTree/internal/http-server/middleware/logger"
	"CommentTree/internal/models"
	"CommentTree/internal/storage/psql"
	"log/slog"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/lib/pq"
)

func main() {
	// TODO: MustLoad func to load config parameters

	// WARN: temporally config data
	cfg := models.Config{
		LogsFile:    "logs/comment_service.log",
		StoragePath: "host=localhost port=5455 user=postresUser password=postresPW dbname=postgres sslmode=disable",
	}

	// TODO: create middleware with logger
	log := slog.New(slog.NewTextHandler(logger.MustNewLocalLogger(cfg.LogsFile), &slog.HandlerOptions{Level: slog.LevelDebug}))

	// TODO: open connection with PostgreSQL and use it to keep comments in
	storage, err := psql.New(cfg.StoragePath)
	if err != nil {
		panic(err)
	}

	// TODO: implement routes and use slog package to log any application actions
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.URLFormat)

	router.Post("/comments", handlers.CreateComment(log, storage))
	router.Get("/comments", handlers.GetComments(log, storage))
	router.Delete("/comments/{id}", handlers.DeleteComment(log, storage))

	// TODO: implement server graceful shutdown
}
