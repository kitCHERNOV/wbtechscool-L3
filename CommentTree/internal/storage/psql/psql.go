package psql

import (
	"CommentTree/internal/models"
	"database/sql"
	"errors"
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

func (s *Storage) CreateCommentTree(articleId uuid.UUID, articleName string, authorId int) error {
	const op = "storage.psql.CreateCommentTree"

	stmt, err := s.db.Prepare(`
		INSERT INTO comments(id, article_id, path, content, author_id)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return fmt.Errorf("%s: preparing statement error; %w", op, err)
	}

	_, err = stmt.Exec(articleId, articleId, articleName, authorId)
	if err != nil {
		return fmt.Errorf("%s: executing statement error; %w", op, err)
	}

	return nil
}

func (s *Storage) CreateComment(comment string, authorId, articleId int, parentCommentID int) (uuid.UUID, error) {
	const op = "storage.psql.CreateComment"
	row := s.db.QueryRow(`SELECT path FROM comments WHERE id = $1`, parentCommentID)
	var path string
	err := row.Scan(&path)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, errors.New("parent comment not found error")
	}

	stmt, err := s.db.Prepare(`
		INSERT INTO comments(id, article_id, parent_id, path, content, author_id)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s Prepare database request error: %v", op, err)
	}

	newComID := uuid.New()
	newComPath := fmt.Sprintf("%s.%s", path, newComID.String())

	_, err = stmt.Exec(newComID, articleId, newComPath, comment, authorId)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: save new comment error; %w", op, err)
	}
	return uuid.Nil, nil
}

func (s *Storage) GetComments(articleId int) (models.Comments, error) {
	return models.Comments{}, nil
}

func (s *Storage) DeleteComment(commentID int) error {
	return nil
}
