package psql

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.psql.New" // Mark for errors

	db, err := sql.Open("postgres", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// TODO: to repair create request to db
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS comments(
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- id комментариев
		article_id INTEGER NOT NULL, -- id статьи комментария
		parent_id  BIGINT, -- для простоты создания и поиска
		path ltree NOT NULL, -- id путь в древовидном представлении
		content TEXT NOT NULL, -- сам комментарий
		author_id INTEGER NOT NULL, -- id автора, чтобы владелец комментария мог удалить и ответы к нему
		created_at TIMESTAMP NOT NULL DEFAULT now()
		);
	create index path_gist_idx on comments using GIST (path);
	CREATE INDEX comments_article_path_idx ON comments (article_id, path);
	`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) CreateComment(comment string, authorId, articleId int, parentComment string) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (s *Storage) DeleteComment(articleId, authorId int) error {
	return nil
}

func (s *Storage) GetComments(articleId int) ([]string, error) {
	return nil, nil
}
