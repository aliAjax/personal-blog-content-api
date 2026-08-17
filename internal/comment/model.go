package comment

import "time"

type Comment struct {
	ID          int64     `json:"id"`
	ArticleID   int64     `json:"article_id"`
	ParentID    *int64    `json:"parent_id,omitempty"`
	AuthorName  string    `json:"author_name"`
	AuthorEmail string    `json:"author_email"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateInput struct {
	ArticleID   int64
	ParentID    *int64
	AuthorName  string
	AuthorEmail string
	Content     string
}

type ListFilter struct {
	ArticleID int64
	Status    string
}

func cloneParentID(id *int64) *int64 {
	if id == nil {
		return nil
	}
	value := *id
	return &value
}
