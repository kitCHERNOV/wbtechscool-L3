package handlers

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

type ImageSaver interface {
	UploadImage() error
	DownloadImage() error
	DeleteImage() error
}

func UploadImage(log *slog.Logger, storage ImageSaver) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "sqlite.UploadImage"

		file, handler, err := r.FormFile("image")
		if err != nil {
			log.Error("%s; error getting file: %v", op, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		defer file.Close()

		// check extension
		extension := filepath.Ext(handler.Filename)
		allowedExtensions := map[string]bool{
			".jpg": true,
			".png": true,
			"gif":  true,
		}

		if !allowedExtensions[extension] {
			log.Warn("%s; file extension not allowed", op)
			http.Error(w, fmt.Sprintf("invalid file extension"), http.StatusBadRequest)
			return
		}

		uploadDir := "./uploads"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			log.Error("%s; error creating uploads directory: %v", op, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		dst, err := os.Create(filepath.Join(uploadDir, handler.Filename))
		if err != nil {
			log.Error("%s; error creating file: %v", op, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			log.Error("%s; error uploading file: %v", op, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "File %s downloaded seccessfuly", handler.Filename)
	}
}

func DownloadImage(log *slog.Logger, storage ImageSaver) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "sqlite.DownloadImage"
	}
}

func DeleteImage(log *slog.Logger, storage ImageSaver) func(http.ResponseWriter, *http.Request) {
	return func(http.ResponseWriter, *http.Request) {
		const op = "sqlite.DeleteImage"
	}
}
