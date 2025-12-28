package handlers

import (
	"CommentTree/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

// Types of requests and responses

type CommentSaver interface {
	CreateCommentTree(articleId uuid.UUID, authorId int) error
	CreateComment(comment string, authorId int, articleId uuid.UUID, parentCommentID uuid.UUID) (uuid.UUID, error)
	GetComments(articleId uuid.UUID) (*models.Comments, error)
	DeleteComment(commentID uuid.UUID) error
}

type createCommentRequest struct {
	ArticleID   string `json:"article_id"`
	AuthorID    int    `json:"author_id"`
	CommentText string `json:"comment_text"`
	ParentID    string `json:"parent_id"`
}

type createCommentResponse struct {
	CommentID uuid.UUID `json:"comment_id"`
	Response  `json:"response"`
}

type getCommentsResponse struct {
	Comments models.Comments `json:"comments,omitempty"`
	Response `json:"response"`
}

type Response struct {
	Status int    `json:"status"`
	Error  string `json:"error,omitempty"`
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
			render.JSON(w, r, createCommentResponse{
				Response: Response{
					Status: http.StatusBadRequest,
					Error:  "failed to read request body",
				},
			})
			return
		}

		var req createCommentRequest
		err = json.Unmarshal(data, &req)
		if err != nil {
			log.Error("json unmarshal error", "error", err)
			render.JSON(w, r, createCommentResponse{
				Response: Response{
					Status: http.StatusBadRequest,
					Error:  "invalid JSON format",
				},
			})
			return
		}

		articleID, err := uuid.Parse(req.ArticleID)
		if err != nil {
			log.Error("parse article_id error", "error", err)
			render.JSON(w, r, createCommentResponse{
				Response: Response{
					Status: http.StatusBadRequest,
					Error:  "invalid article_id format",
				},
			})
			return
		}

		parentID, err := uuid.Parse(req.ParentID)
		if err != nil {
			log.Error("parse parent_id error", "error", err)
			render.JSON(w, r, createCommentResponse{
				Response: Response{
					Status: http.StatusBadRequest,
					Error:  "invalid parent_id format",
				},
			})
			return
		}

		id, err := commentSaver.CreateComment(req.CommentText, req.AuthorID, articleID, parentID)
		if err != nil {
			log.Error("create comment error", "error", err)
			render.JSON(w, r, createCommentResponse{
				uuid.Nil,
				Response{
					Status: http.StatusInternalServerError,
				},
			})
		} else if id != uuid.Nil {
			log.Info("create comment error; id is nil", "id", id)
			render.JSON(w, r, createCommentResponse{
				CommentID: uuid.Nil,
				Response: Response{
					Status: http.StatusInternalServerError,
					Error:  fmt.Sprintf("create comment error; id is nil"),
				},
			})
		}

		render.JSON(w, r, createCommentResponse{
			CommentID: id,
			Response: Response{
				Status: http.StatusOK,
			},
		})
	}
}

func GetComments(log *slog.Logger, commentSaver CommentSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.GetComments"
		parentID := r.URL.Query().Get("parent")
		if parentID == "" {
			log.Error(fmt.Sprint(op+"parentID is empty"), "url", r.URL)
			render.JSON(w, r, getCommentsResponse{
				Response: Response{
					Status: http.StatusBadRequest,
					Error:  fmt.Sprintf("parent id is required"),
				},
			})
		}

		//intParentID, err := strconv.Atoi(parentID)
		intParentID, err := uuid.Parse(parentID)
		if err != nil {
			log.Error(fmt.Sprint(op+"parentID is invalid"), "url", r.URL)
			render.JSON(w, r, getCommentsResponse{
				Response: Response{
					Status: http.StatusBadRequest,
					Error:  fmt.Sprintf("parent id is required"),
				},
			})
		}
		comments, err := commentSaver.GetComments(intParentID)
		if err != nil {
			log.Error(fmt.Sprint(op+"GetComments error"), "url", r.URL)
			render.JSON(w, r, getCommentsResponse{
				Response: Response{
					Status: http.StatusInternalServerError,
					Error:  fmt.Sprintf("Internal server error"),
				},
			})
		}

		render.JSON(w, r, getCommentsResponse{
			Comments: *comments,
			Response: Response{
				Status: http.StatusOK,
			},
		})
	}
}

func DeleteComment(log *slog.Logger, commentSaver CommentSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.DeleteComment"
		strChiID := chi.URLParam(r, "id")
		if strChiID == "" {
			log.Error(fmt.Sprint(op+"chiID is empty"), "url", r.URL)
			render.JSON(w, r, Response{
				Status: http.StatusBadRequest,
				Error:  fmt.Sprintf("chiID is required"),
			})
		}
		commentID, err := uuid.Parse(strChiID)
		if err != nil {
			log.Error(fmt.Sprint(op+"GetComments id is invalid"), "url", r.URL)
			render.JSON(w, r, Response{
				Status: http.StatusBadRequest,
				Error:  fmt.Sprintf("id is incorrect"),
			})
		}

		err = commentSaver.DeleteComment(commentID)
		if err != nil {
			log.Error(fmt.Sprint(op+"DeleteComment error"), "url", r.URL)
			render.JSON(w, r, Response{
				Status: http.StatusInternalServerError,
			})
		}

	}
}
