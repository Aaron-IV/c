package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// isTextEmpty checks if text is empty or contains only whitespace
func isTextEmpty(text string) bool {
	return strings.TrimSpace(text) == ""
}

// registerHandler handles user registration
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Debug logging
	log.Printf("Registration attempt - Username: '%s', Email: '%s', Password: '%s'", username, email, password)

	if username == "" || email == "" || password == "" {
		ErrorResponse(w, http.StatusBadRequest, "All fields are required")
		return
	}

	if !isValidEmail(email) {
		ErrorResponse(w, http.StatusBadRequest, "Некорректный формат email")
		return
	}

	// Check if email already exists
	existingUser, _ := getUserByEmail(email)
	if existingUser != nil {
		ErrorResponse(w, http.StatusConflict, "Email already registered")
		return
	}

	// Hash password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Error processing request")
		return
	}

	// Create user
	err = createUser(username, email, hashedPassword)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Error creating user")
		return
	}

	JSONResponse(w, http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

// loginHandler handles user login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		ErrorResponse(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Get user by email
	user, err := getUserByEmail(email)
	if err != nil {
		ErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Check password
	if !checkPassword(password, user.Password) {
		ErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Delete all previous sessions for this user
	_ = deleteAllSessionsForUser(user.ID)

	// Create session
	session, err := createUserSession(user.ID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Error creating session")
		return
	}

	// Set session cookie
	setSessionCookie(w, session.ID)

	JSONResponse(w, http.StatusOK, map[string]string{"message": "Login successful"})
}

// logoutHandler handles user logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err == nil {
		deleteSession(cookie.Value)
	}

	clearSessionCookie(w)
	JSONResponse(w, http.StatusOK, map[string]string{"message": "Logout successful"})
}

// createPostHandler handles post creation
func createPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is logged in
	user, err := getCurrentUser(r)
	if err != nil {
		ErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if err := r.ParseForm(); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	categoriesStr := r.FormValue("categories")

	if isTextEmpty(title) || isTextEmpty(content) {
		ErrorResponse(w, http.StatusBadRequest, "Title and content are required")
		return
	}
	if len(title) < 5 || len(title) > 100 {
		ErrorResponse(w, http.StatusBadRequest, "Оглавление поста должно быть от 5 до 100 символов")
		return
	}
	if len(content) < 10 || len(content) > 2000 {
		ErrorResponse(w, http.StatusBadRequest, "Текст поста должен быть от 10 до 2000 символов")
		return
	}

	// Получить все существующие категории
	allCategories, err := getCategories()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Error processing categories")
		return
	}
	categoryNameToID := make(map[string]int)
	for _, cat := range allCategories {
		categoryNameToID[cat.Name] = cat.ID
	}

	var categoryIDs []int
	var selectedNames []string
	if categoriesStr != "" {
		categoryNames := strings.Split(categoriesStr, ",")
		for _, name := range categoryNames {
			name = strings.TrimSpace(name)
			if name != "" {
				if id, ok := categoryNameToID[name]; ok {
					categoryIDs = append(categoryIDs, id)
					selectedNames = append(selectedNames, name)
				}
			}
		}
		if len(categoryIDs) > 4 {
			ErrorResponse(w, http.StatusBadRequest, "Можно выбрать не более 4 категорий")
			return
		}
	}
	if len(categoryIDs) == 0 {
		// Если нет ни одной категории — добавить 'Другие', создать если нет
		otherID, ok := categoryNameToID["Другие"]
		if !ok {
			// Создать категорию 'Другие'
			_, err := db.Exec("INSERT INTO categories (name) VALUES (?)", "Другие")
			if err != nil {
				ErrorResponse(w, http.StatusInternalServerError, "Не удалось создать категорию 'Другие'")
				return
			}
			// Получить id только что созданной категории
			row := db.QueryRow("SELECT id FROM categories WHERE name = ?", "Другие")
			row.Scan(&otherID)
		}
		categoryIDs = append(categoryIDs, otherID)
	}

	// Create post
	postID, err := createPost(title, content, user.ID, categoryIDs)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Error creating post")
		return
	}

	JSONResponse(w, http.StatusCreated, map[string]interface{}{
		"message": "Post created successfully",
		"post_id": postID,
	})
}

// createCommentHandler handles comment creation
func createCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is logged in
	user, err := getCurrentUser(r)
	if err != nil {
		ErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if err := r.ParseForm(); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	postIDStr := r.FormValue("post_id")
	content := r.FormValue("content")

	if postIDStr == "" || isTextEmpty(content) {
		ErrorResponse(w, http.StatusBadRequest, "Post ID and content are required")
		return
	}
	if len(content) < 2 || len(content) > 500 {
		ErrorResponse(w, http.StatusBadRequest, "Комментарий должен быть от 2 до 500 символов")
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	// Create comment
	err = createComment(postID, content, user.ID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Error creating comment")
		return
	}

	JSONResponse(w, http.StatusCreated, map[string]string{"message": "Comment created successfully"})
}

// likeHandler handles likes and dislikes
func likeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is logged in
	user, err := getCurrentUser(r)
	if err != nil {
		ErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if err := r.ParseForm(); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	postIDStr := r.FormValue("post_id")
	commentIDStr := r.FormValue("comment_id")
	isLikeStr := r.FormValue("is_like")

	if (postIDStr == "" && commentIDStr == "") || isLikeStr == "" {
		ErrorResponse(w, http.StatusBadRequest, "Post ID or comment ID and is_like are required")
		return
	}

	isLike := isLikeStr == "true"

	var postID *int
	var commentID *int

	if postIDStr != "" {
		id, err := strconv.Atoi(postIDStr)
		if err != nil {
			ErrorResponse(w, http.StatusBadRequest, "Invalid post ID")
			return
		}
		postID = &id
	}

	if commentIDStr != "" {
		id, err := strconv.Atoi(commentIDStr)
		if err != nil {
			ErrorResponse(w, http.StatusBadRequest, "Invalid comment ID")
			return
		}
		commentID = &id
	}

	// Toggle like
	err = toggleLike(user.ID, postID, commentID, isLike)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Error processing like")
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"message": "Like updated successfully"})
}

// postsHandler handles getting posts with filtering
func postsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user (optional)
	var userID *int
	user, err := getCurrentUser(r)
	if err == nil {
		userID = &user.ID
		log.Printf("PostsHandler - User authenticated: ID=%d, Username=%s", user.ID, user.Username)
	} else {
		log.Printf("PostsHandler - User not authenticated: %v", err)
	}

	// Get filter parameters
	filter := r.URL.Query().Get("filter")
	filterValue := r.URL.Query().Get("value")
	log.Printf("PostsHandler - Filter: %s, FilterValue: %s", filter, filterValue)

	// Только для текущего пользователя фильтры 'created' и 'liked'
	if (filter == "created" || filter == "liked") && userID == nil {
		log.Printf("PostsHandler - Authentication required for filter: %s", filter)
		ErrorResponse(w, http.StatusUnauthorized, "Authentication required for this filter")
		return
	}

	// Get posts
	posts, err := getPosts(userID, filter, filterValue)
	if err != nil {
		log.Printf("PostsHandler - Error getting posts: %v", err)
		ErrorResponse(w, http.StatusInternalServerError, "Error retrieving posts")
		return
	}
	if posts == nil {
		posts = []Post{}
	}
	log.Printf("PostsHandler - Returning %d posts", len(posts))
	JSONResponse(w, http.StatusOK, posts)
}

// postHandler handles getting a specific post with comments
func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract post ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	log.Printf("PostHandler - URL: %s, PathParts: %v", r.URL.Path, pathParts)
	if len(pathParts) < 3 {
		ErrorResponse(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	postIDStr := pathParts[2] // pathParts = ["api", "post", "5"]
	log.Printf("PostHandler - PostID string: %s", postIDStr)
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		log.Printf("PostHandler - Error parsing post ID: %v", err)
		ErrorResponse(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	// Get current user (optional)
	var userID *int
	user, err := getCurrentUser(r)
	if err == nil {
		userID = &user.ID
	}

	// Get post details (simplified - you'd want to create a specific function for this)
	posts, err := getPosts(userID, "", "")
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Error retrieving post")
		return
	}

	var targetPost *Post
	for _, post := range posts {
		if post.ID == postID {
			targetPost = &post
			break
		}
	}

	if targetPost == nil {
		ErrorResponse(w, http.StatusNotFound, "Post not found")
		return
	}

	// Get comments for this post
	comments, err := getComments(postID, userID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Error retrieving comments")
		return
	}
	if comments == nil {
		comments = []Comment{}
	}
	response := map[string]interface{}{
		"post":     targetPost,
		"comments": comments,
	}

	JSONResponse(w, http.StatusOK, response)
}

// categoriesHandler handles getting all categories
func categoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	categories, err := getCategories()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Error retrieving categories")
		return
	}

	JSONResponse(w, http.StatusOK, categories)
}

// userHandler handles getting current user info
func userHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := getCurrentUser(r)
	if err != nil {
		ErrorResponse(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Don't include password in response
	userResponse := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"created":  user.Created,
	}

	JSONResponse(w, http.StatusOK, userResponse)
}

func isValidEmail(email string) bool {
	var re = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// renderError renders a styled error page
func renderError(w http.ResponseWriter, message string) {
	path := filepath.Join("templates", "error.html")
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Println("Template error:", err)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	tmpl.Execute(w, map[string]string{"Message": message})
}
