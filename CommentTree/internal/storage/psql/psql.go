package psql

import (
	"CommentTree/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS comments(
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- id комментариев
		article_id UUID NOT NULL, -- id статьи комментария
		parent_id  UUID, -- для простоты создания и поиска
		path ltree NOT NULL, -- id путь в древовидном представлении
		content TEXT NOT NULL, -- сам комментарий
		author_id INT NOT NULL, -- id автора, чтобы владелец комментария мог удалить и ответы к нему
		created_at TIMESTAMP NOT NULL DEFAULT now()
		);
	create index IF NOT EXISTS path_gist_idx on comments using GIST (path);
	CREATE INDEX IF NOT EXISTS comments_article_path_idx ON comments (article_id, path);
	`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// TODO: to tidy something useless parameters
func (s *Storage) CreateCommentTree(articleId uuid.UUID, authorId int) error {
	const op = "storage.psql.CreateCommentTree"

	stmt, err := s.db.Prepare(`
		INSERT INTO comments(id, article_id, path, content, author_id)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return fmt.Errorf("%s: preparing statement error; %w", op, err)
	}

	_, err = stmt.Exec(articleId, articleId, articleId, "", authorId)
	if err != nil {
		return fmt.Errorf("%s: executing statement error; %w", op, err)
	}

	return nil
}

func (s *Storage) CreateComment(comment string, authorId int, articleId uuid.UUID, parentCommentID uuid.UUID) (uuid.UUID, error) {
	const op = "storage.psql.CreateComment"
	row := s.db.QueryRow(`SELECT path FROM comments WHERE id = $1`, parentCommentID)
	var path string
	err := row.Scan(&path)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, errors.New("parent comment not found error")
	}

	newComID := uuid.New()
	strUUID := newComID.String()
	strUUID = strings.ReplaceAll(strUUID, "-", "_")

	newComPath := fmt.Sprintf("%s.%s", path, strUUID)

	_ = s.db.QueryRow(`
		INSERT INTO comments(id, article_id, parent_id, path, content, author_id)
		VALUES ($1, $2, $3, $4, $5, $6);
	`, newComID, articleId, parentCommentID, newComPath, comment, authorId,
	)

	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: save new comment error; %w", op, err)
	}

	return newComID, nil
}

func (s *Storage) GetComments(articleId uuid.UUID) (*models.Comments, error) {
	const op = "storage.psql.GetComments"
	// TODO: check if columns are correct
	// WARN: ставлю на то что сначала будут считываться родители после дети
	rows, err := s.db.Query(`SELECT id::text, COALESCE(parent_id::text, ''), path::text, content
					FROM comments 
					WHERE path <@ (SELECT path FROM comments WHERE id = $1)
					ORDER BY path;`, articleId)
	if err != nil {
		return nil, fmt.Errorf("%s query error to get all comments; %w", op, err)
	}
	defer rows.Close()

	nodeMap := make(map[string]*models.Comments)
	var rootComment *models.Comments

	for rows.Next() {
		comment := &models.Comments{
			SubComments: []*models.Comments{},
		}

		err := rows.Scan(&comment.CommentID, &comment.ParentID,
			&comment.Path, &comment.CommentText)
		if err != nil {
			return nil, fmt.Errorf("%s scan error: %w", op, err)
		}

		nodeMap[comment.CommentID] = comment

		if comment.CommentID == articleId.String() {
			rootComment = comment
		} else if parent, exists := nodeMap[comment.ParentID]; exists {
			parent.SubComments = append(parent.SubComments, comment)
		}
	}

	if rootComment == nil {
		return nil, fmt.Errorf("%s root comment not found", op)
	}

	return rootComment, nil
}

func (s *Storage) DeleteComment(commentID uuid.UUID) error {
	const op = "storage.psql.DeleteComment"

	_, err := s.db.Exec(`
		DELETE FROM comments
		WHERE path <@ (SELECT path FROM comments WHERE id = $1);
	`, commentID)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
