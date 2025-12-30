package sqlite

import (
	"database/sql"
	"fmt"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "sqlite.New"
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s; error opening sqlite3 storage: %v", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) UploadImage() error {
	const op = "sqlite.UploadImage"

	return nil
}

func (s *Storage) DownloadImage() error {
	return nil
}

func (s *Storage) DeleteImage() error {
	const op = "sqlite.DeleteImage"
	return nil
}
