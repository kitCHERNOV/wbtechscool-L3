package models

type Config struct {
	LogsFile    string
	StoragePath string
}

type Comments struct {
	CommentText string      `json:"comment_text,omitempty"`
	ParentID    string      `json:"parent_id,omitempty"`
	CommentID   string      `json:"comment_id,omitempty"`
	Path        string      `json:"path,omitempty"`
	SubComments []*Comments `json:"sub_comments,omitempty"`
}

//CREATE TABLE IF NOT EXISTS comments(
// 		id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- id комментариев
// 		article_id INTEGER NOT NULL, -- id статьи комментария
// 		parent_id  BIGINT, -- для простоты создания и поиска
// 		path ltree NOT NULL, -- id путь в древовидном представлении
// 		content TEXT NOT NULL, -- сам комментарий
// 		author_id INTEGER NOT NULL, -- id автора, чтобы владелец комментария мог удалить и ответы к нему
// 		created_at TIMESTAMP NOT NULL DEFAULT now()
//);
