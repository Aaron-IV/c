package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// initDB initializes the database and creates all necessary tables
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create users table
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Create categories table
	createCategoriesTable := `
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL
	);`

	// Create posts table
	createPostsTable := `
	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		author_id INTEGER NOT NULL,
		created DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (author_id) REFERENCES users (id)
	);`

	// Create comments table
	createCommentsTable := `
	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		author_id INTEGER NOT NULL,
		created DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (post_id) REFERENCES posts (id),
		FOREIGN KEY (author_id) REFERENCES users (id)
	);`

	// Create sessions table
	createSessionsTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users (id)
	);`

	// Create likes table
	createLikesTable := `
	CREATE TABLE IF NOT EXISTS likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		post_id INTEGER,
		comment_id INTEGER,
		is_like BOOLEAN NOT NULL,
		created DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id),
		FOREIGN KEY (post_id) REFERENCES posts (id),
		FOREIGN KEY (comment_id) REFERENCES comments (id),
		UNIQUE(user_id, post_id, comment_id)
	);`

	// Create post_categories table (many-to-many relationship)
	createPostCategoriesTable := `
	CREATE TABLE IF NOT EXISTS post_categories (
		post_id INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		PRIMARY KEY (post_id, category_id),
		FOREIGN KEY (post_id) REFERENCES posts (id),
		FOREIGN KEY (category_id) REFERENCES categories (id)
	);`

	// Execute all table creation statements
	statements := []string{
		createUsersTable,
		createCategoriesTable,
		createPostsTable,
		createCommentsTable,
		createSessionsTable,
		createLikesTable,
		createPostCategoriesTable,
	}

	for _, stmt := range statements {
		_, err = db.Exec(stmt)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Insert default categories if they don't exist
	insertDefaultCategories()
}

// insertDefaultCategories adds some default categories to the forum
func insertDefaultCategories() {
	categories := []string{"Общие", "Технологии", "Спорт", "Кино", "Музыка", "Книги", "Путешествия", "Другие"}

	for _, category := range categories {
		_, err := db.Exec("INSERT OR IGNORE INTO categories (name) VALUES (?)", category)
		if err != nil {
			log.Printf("Error inserting category %s: %v", category, err)
		}
	}
}

// getUserByEmail retrieves a user by email
func getUserByEmail(email string) (*User, error) {
	user := &User{}
	err := db.QueryRow("SELECT id, username, email, password, created FROM users WHERE email = ?", email).
		Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Created)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// getUserByID retrieves a user by ID
func getUserByID(id int) (*User, error) {
	user := &User{}
	err := db.QueryRow("SELECT id, username, email, password, created FROM users WHERE id = ?", id).
		Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Created)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// createUser creates a new user
func createUser(username, email, password string) error {
	_, err := db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", username, email, password)
	return err
}

// createSession creates a new session for a user
func createSession(userID int, sessionID string, expiresAt time.Time) error {
	_, err := db.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, userID, expiresAt)
	return err
}

// getSession retrieves a session by ID
func getSession(sessionID string) (*Session, error) {
	session := &Session{}
	err := db.QueryRow("SELECT id, user_id, expires_at FROM sessions WHERE id = ?", sessionID).
		Scan(&session.ID, &session.UserID, &session.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// deleteSession deletes a session
func deleteSession(sessionID string) error {
	_, err := db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	return err
}

// deleteAllSessionsForUser deletes all sessions for a given user ID
func deleteAllSessionsForUser(userID int) error {
	_, err := db.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	return err
}

// createPost creates a new post
func createPost(title, content string, authorID int, categoryIDs []int) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec("INSERT INTO posts (title, content, author_id) VALUES (?, ?, ?)", title, content, authorID)
	if err != nil {
		return 0, err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Add categories to the post
	for _, categoryID := range categoryIDs {
		_, err = tx.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, categoryID)
		if err != nil {
			return 0, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return postID, nil
}

// getPosts retrieves all posts with optional filtering
func getPosts(userID *int, filter string, filterValue string) ([]Post, error) {
	var query string
	var args []interface{}

	log.Printf("getPosts - Filter: %s, FilterValue: %s, UserID: %v", filter, filterValue, userID)

	switch filter {
	case "category":
		query = `
			SELECT p.id, p.title, p.content, p.author_id, u.username, p.created, p.updated,
				   (SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 1) as likes,
				   (SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 0) as dislikes
			FROM posts p
			JOIN users u ON p.author_id = u.id
			JOIN post_categories pc ON p.id = pc.post_id
			JOIN categories c ON pc.category_id = c.id
			WHERE c.name = ?
			ORDER BY p.created DESC`
		args = append(args, filterValue)
		log.Printf("getPosts - Using category filter with value: %s", filterValue)
	case "created":
		if userID == nil {
			log.Printf("getPosts - UserID is nil for created filter, returning empty")
			return nil, nil
		}
		query = `
			SELECT p.id, p.title, p.content, p.author_id, u.username, p.created, p.updated,
				   (SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 1) as likes,
				   (SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 0) as dislikes
			FROM posts p
			JOIN users u ON p.author_id = u.id
			WHERE p.author_id = ?
			ORDER BY p.created DESC`
		args = append(args, *userID)
		log.Printf("getPosts - Using created filter for user ID: %d", *userID)
	case "liked":
		if userID == nil {
			log.Printf("getPosts - UserID is nil for liked filter, returning empty")
			return nil, nil
		}
		query = `
			SELECT p.id, p.title, p.content, p.author_id, u.username, p.created, p.updated,
				   (SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 1) as likes,
				   (SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 0) as dislikes
			FROM posts p
			JOIN users u ON p.author_id = u.id
			JOIN likes l ON p.id = l.post_id
			WHERE l.user_id = ? AND l.is_like = 1
			ORDER BY p.created DESC`
		args = append(args, *userID)
		log.Printf("getPosts - Using liked filter for user ID: %d", *userID)
	default:
		query = `
			SELECT p.id, p.title, p.content, p.author_id, u.username, p.created, p.updated,
				   (SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 1) as likes,
				   (SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 0) as dislikes
			FROM posts p
			JOIN users u ON p.author_id = u.id
			ORDER BY p.created DESC`
		log.Printf("getPosts - Using default filter (all posts)")
	}

	log.Printf("getPosts - Executing query: %s with args: %v", query, args)
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("getPosts - Database error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.AuthorName, &post.Created, &post.Updated, &post.Likes, &post.Dislikes)
		if err != nil {
			return nil, err
		}

		// Get categories for this post
		categories, err := getPostCategories(post.ID)
		if err != nil {
			return nil, err
		}
		post.Categories = categories

		// Get user's like/dislike status if logged in
		if userID != nil {
			userLike, userDislike, err := getUserPostLikeStatus(*userID, post.ID)
			if err == nil {
				post.UserLiked = userLike
				post.UserDisliked = userDislike
			}
		}

		posts = append(posts, post)
	}

	log.Printf("getPosts - Found %d posts", len(posts))
	return posts, nil
}

// getPostCategories retrieves categories for a specific post
func getPostCategories(postID int) ([]string, error) {
	rows, err := db.Query("SELECT c.name FROM categories c JOIN post_categories pc ON c.id = pc.category_id WHERE pc.post_id = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		err := rows.Scan(&category)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// getUserPostLikeStatus gets the like/dislike status for a user on a specific post
func getUserPostLikeStatus(userID, postID int) (*bool, *bool, error) {
	var isLike bool
	err := db.QueryRow("SELECT is_like FROM likes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&isLike)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	userLiked := isLike
	userDisliked := !isLike
	return &userLiked, &userDisliked, nil
}

// getCategories retrieves all categories
func getCategories() ([]Category, error) {
	rows, err := db.Query("SELECT id, name FROM categories ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// createComment creates a new comment
func createComment(postID int, content string, authorID int) error {
	_, err := db.Exec("INSERT INTO comments (post_id, content, author_id) VALUES (?, ?, ?)", postID, content, authorID)
	return err
}

// getComments retrieves comments for a specific post
func getComments(postID int, userID *int) ([]Comment, error) {
	query := `
		SELECT c.id, c.post_id, c.content, c.author_id, u.username, c.created,
			   (SELECT COUNT(*) FROM likes WHERE comment_id = c.id AND is_like = 1) as likes,
			   (SELECT COUNT(*) FROM likes WHERE comment_id = c.id AND is_like = 0) as dislikes
		FROM comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created ASC`

	rows, err := db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.Content, &comment.AuthorID, &comment.AuthorName, &comment.Created, &comment.Likes, &comment.Dislikes)
		if err != nil {
			return nil, err
		}

		// Get user's like/dislike status if logged in
		if userID != nil {
			userLike, userDislike, err := getUserCommentLikeStatus(*userID, comment.ID)
			if err == nil {
				comment.UserLiked = userLike
				comment.UserDisliked = userDislike
			}
		}

		comments = append(comments, comment)
	}

	return comments, nil
}

// getUserCommentLikeStatus gets the like/dislike status for a user on a specific comment
func getUserCommentLikeStatus(userID, commentID int) (*bool, *bool, error) {
	var isLike bool
	err := db.QueryRow("SELECT is_like FROM likes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&isLike)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	userLiked := isLike
	userDisliked := !isLike
	return &userLiked, &userDisliked, nil
}

// toggleLike toggles a like/dislike on a post or comment
func toggleLike(userID int, postID *int, commentID *int, isLike bool) error {
	// First, check if there's already a like/dislike
	var existingID int
	var existingIsLike bool
	var err error

	if postID != nil {
		err = db.QueryRow("SELECT id, is_like FROM likes WHERE user_id = ? AND post_id = ?", userID, *postID).Scan(&existingID, &existingIsLike)
	} else if commentID != nil {
		err = db.QueryRow("SELECT id, is_like FROM likes WHERE user_id = ? AND comment_id = ?", userID, *commentID).Scan(&existingID, &existingIsLike)
	}

	if err == sql.ErrNoRows {
		// No existing like/dislike, create new one
		if postID != nil {
			_, err = db.Exec("INSERT INTO likes (user_id, post_id, is_like) VALUES (?, ?, ?)", userID, *postID, isLike)
		} else if commentID != nil {
			_, err = db.Exec("INSERT INTO likes (user_id, comment_id, is_like) VALUES (?, ?, ?)", userID, *commentID, isLike)
		}
		return err
	} else if err != nil {
		return err
	}

	// Existing like/dislike found
	if existingIsLike == isLike {
		// Same type, remove it
		_, err = db.Exec("DELETE FROM likes WHERE id = ?", existingID)
	} else {
		// Different type, update it
		_, err = db.Exec("UPDATE likes SET is_like = ? WHERE id = ?", isLike, existingID)
	}

	return err
}
