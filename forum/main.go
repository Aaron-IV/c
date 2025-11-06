package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// postsRouteHandler handles both GET and POST requests for /api/posts
func postsRouteHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		postsHandler(w, r)
	case "POST":
		createPostHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func renderHTML(w http.ResponseWriter, filename string, data interface{}) {
	path := filepath.Join("templates", filename)
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Println("Template error:", err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		renderError(w, "Страница не найдена")
		return
	}
	renderHTML(w, "index.html", nil)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	renderHTML(w, "about.html", nil)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok", "message": "Сервер работает"}`))
}

func main() {
	// Initialize database
	initDB()
	defer db.Close()

	// Static files (CSS, JS)
	fs := http.FileServer(http.Dir("templates"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// API routes
	http.HandleFunc("/api/register", registerHandler)
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/logout", logoutHandler)
	http.HandleFunc("/api/user", userHandler)
	http.HandleFunc("/api/posts", postsRouteHandler)
	http.HandleFunc("/api/post/", postHandler)
	http.HandleFunc("/api/comments", createCommentHandler)
	http.HandleFunc("/api/like", likeHandler)
	http.HandleFunc("/api/categories", categoriesHandler)
	http.HandleFunc("/api/health", healthHandler)

	// Page routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/about", aboutHandler)

	// Start server
	port := ":8080"
	log.Printf("Сервер запущен на http://localhost%s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
