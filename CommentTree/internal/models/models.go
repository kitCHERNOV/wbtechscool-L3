package models

type Config struct {
	LogsFile    string
	StoragePath string
}

type Comments struct {
	CommentText string     `json:"comment_text,omitempty"`
	ParentID    int        `json:"parent_id,omitempty"`
	CommentID   int        `json:"comment_id,omitempty"`
	SubComments []Comments `json:"sub_comments,omitempty"`
}
