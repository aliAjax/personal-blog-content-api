package article

import "time"

type Article struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	CategoryID  *int64     `json:"category_id,omitempty"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Excerpt     string     `json:"excerpt"`
	Content     string     `json:"content"`
	Status      string     `json:"status"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Category    *Category  `json:"category,omitempty"`
	Tags        []Tag      `json:"tags"`
}

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type CreateInput struct {
	UserID     int64
	CategoryID *int64
	Title      string
	Slug       string
	Excerpt    string
	Content    string
	Status     string
	TagIDs     []int64
}

type UpdateInput struct {
	CategoryID *int64
	Title      string
	Slug       string
	Excerpt    string
	Content    string
	Status     string
	TagIDs     []int64
}

type ListFilter struct {
	Status        string
	CategoryID    int64
	TagID         int64
	Query         string
	PublishedOnly bool
}

func cloneTagIDs(ids []int64) []int64 {
	if ids == nil {
		return nil
	}
	return append([]int64(nil), ids...)
}
