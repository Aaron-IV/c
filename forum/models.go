package main

import (
	"time"
)

// User represents a forum user
type User struct {
	ID       int       `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Password string    `json:"-"` // Don't expose password in JSON
	Created  time.Time `json:"created"`
}

// Post represents a forum post
type Post struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	AuthorID     int       `json:"author_id"`
	AuthorName   string    `json:"author_name"`
	Created      time.Time `json:"created"`
	Updated      time.Time `json:"updated"`
	Likes        int       `json:"likes"`
	Dislikes     int       `json:"dislikes"`
	Categories   []string  `json:"categories"`
	UserLiked    *bool     `json:"user_liked,omitempty"`    // For logged in users
	UserDisliked *bool     `json:"user_disliked,omitempty"` // For logged in users
}

// Comment represents a comment on a post
type Comment struct {
	ID           int       `json:"id"`
	PostID       int       `json:"post_id"`
	Content      string    `json:"content"`
	AuthorID     int       `json:"author_id"`
	AuthorName   string    `json:"author_name"`
	Created      time.Time `json:"created"`
	Likes        int       `json:"likes"`
	Dislikes     int       `json:"dislikes"`
	UserLiked    *bool     `json:"user_liked,omitempty"`
	UserDisliked *bool     `json:"user_disliked,omitempty"`
}

// Category represents a post category
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Like represents a like/dislike on a post or comment
type Like struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	PostID    *int      `json:"post_id,omitempty"`
	CommentID *int      `json:"comment_id,omitempty"`
	IsLike    bool      `json:"is_like"` // true for like, false for dislike
	Created   time.Time `json:"created"`
}

// PostCategory represents the many-to-many relationship between posts and categories
type PostCategory struct {
	PostID     int `json:"post_id"`
	CategoryID int `json:"category_id"`
}
