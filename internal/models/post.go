package models

import "time"

// Post represents a Reddit post
type Post struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	Author      string    `json:"author"`
	URL         string    `json:"url"`
	Timestamp   time.Time `json:"timestamp"`
	Comments    []Comment `json:"comments,omitempty"`
	ProcessedAt time.Time `json:"processed_at"`
}

// Comment represents a Reddit comment
type Comment struct {
	Author string `json:"author"`
	Body   string `json:"body"`
}
