package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type CommentSaver interface {
	CreateComment(authorId, articleId int, parentCommentID int, comment string) (uuid.UUID, error)
	GetComments(articleId int) ([]string, error)
	DeleteComment(articleId, authorId int) error
}

type createCommentRequest struct {
	ArticleID   int    `json:"article_id"`
	AuthorID    int    `json:"author_id"`
	CommentText string `json:"comment_text"`
	ParentID    int    `json:"parent_id"`
}

type createCommentResponse struct {
	CommentID uuid.UUID `json:"comment_id"`
	Response  `json:"response"`
}

type Response struct {
	// TODO: добавить поле отвечающее интерфейсу, типа все что иммеет метод Responser сами сделаем
	Status int   `json:"status"`
	Error  error `json:"error,omitempty"`
}

// CreateComment godoc
// @Summary      Create comment
// @Description  Creates a comment; parent_id is optional for root comments
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        article_id  path
// @Param        request     body
// @Success      201 {object} CommentDTO
// @Failure      400 {object} HTTPError
// @Failure      500 {object} HTTPError
// @Router       /comments [post]
func CreateComment(log *slog.Logger, commentSaver CommentSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("reading body error", "error", err)
		}

		var req createCommentRequest
		err = json.Unmarshal(data, &req)
		if err != nil {
			log.Error("json unmarshal error", "error", err)
		}

		id, err := commentSaver.CreateComment(req.AuthorID, req.ArticleID, req.ParentID, req.CommentText)
		if err != nil {
			log.Error("create comment error", "error", err)
			render.JSON(w, r, createCommentResponse{
				uuid.Nil,
				Response{
					http.StatusInternalServerError,
					err,
				},
			})
		} else if id != uuid.Nil {
			log.Info("create comment error; id is nil", "id", id)
			render.JSON(w, r, Response{
				Status: http.StatusInternalServerError,
				Error:  fmt.Errorf("create comment error; id is nil"),
			})
		}

		render.JSON(w, r, Response{
			Status: http.StatusOK,
		})

	}
}
