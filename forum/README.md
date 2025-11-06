# Web Forum

A comprehensive web forum built with Go, featuring user authentication, post creation, commenting, likes/dislikes, and filtering capabilities.

## Features

### ✅ Core Requirements
- **SQLite Database**: Uses SQLite for data storage with proper table relationships
- **User Authentication**: Registration, login, and session management with cookies
- **Post Management**: Create posts with categories, view all posts
- **Commenting System**: Add comments to posts
- **Like/Dislike System**: Like and dislike posts and comments
- **Filtering**: Filter posts by categories, created posts, and liked posts
- **Docker Support**: Full containerization with Docker and docker-compose

### ✅ Bonus Features
- **Password Encryption**: Passwords are hashed using bcrypt
- **UUID Sessions**: Uses UUID for session management
- **Modern UI**: Responsive design with modern styling
- **Error Handling**: Comprehensive error handling and HTTP status codes
- **API Documentation**: RESTful API with JSON responses

## Technology Stack

- **Backend**: Go 1.24
- **Database**: SQLite3
- **Authentication**: bcrypt for password hashing
- **Sessions**: UUID-based session management
- **Frontend**: HTML, CSS, JavaScript (vanilla)
- **Containerization**: Docker & Docker Compose

## Database Schema

The forum uses the following tables:
- `users` - User accounts and authentication
- `posts` - Forum posts with titles and content
- `comments` - Comments on posts
- `categories` - Post categories
- `post_categories` - Many-to-many relationship between posts and categories
- `likes` - Like/dislike records for posts and comments
- `sessions` - User session management

## API Endpoints

### Authentication
- `POST /api/register` - User registration
- `POST /api/login` - User login
- `POST /api/logout` - User logout
- `GET /api/user` - Get current user info

### Posts
- `GET /api/posts` - Get all posts (with optional filtering)
- `POST /api/posts` - Create a new post
- `GET /api/post/{id}` - Get specific post with comments

### Comments
- `POST /api/comments` - Create a new comment

### Likes
- `POST /api/like` - Toggle like/dislike on post or comment

### Categories
- `GET /api/categories` - Get all categories

### Health
- `GET /api/health` - Health check endpoint

## Setup Instructions

### Prerequisites
- Go 1.24 or later
- Docker and Docker Compose (for containerized deployment)

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd forum
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Run the application**
   ```bash
   go run .
   ```

4. **Access the forum**
   Open your browser and navigate to `http://localhost:8080`

#

## Usage

### Registration and Login
1. Click "Регистрация" to create a new account
2. Fill in username, email, and password
3. Click "Войти" to log in with your credentials

### Creating Posts
1. Log in to your account
2. Click "Создать пост"
3. Fill in title, content, and categories (optional)
4. Submit to create your post

### Interacting with Posts
- **View Posts**: All posts are visible to everyone
- **Like/Dislike**: Logged-in users can like or dislike posts and comments
- **Comment**: Logged-in users can add comments to posts
- **Filter**: Use the sidebar to filter posts by categories or view your own posts/liked posts

### Categories
The forum comes with default categories:
- Общие (General)
- Технологии (Technology)
- Спорт (Sports)
- Кино (Movies)
- Музыка (Music)
- Книги (Books)
- Путешествия (Travel)

## Project Structure

```
forum/
├── main.go           # Application entry point and route setup
├── models.go         # Data structures and types
├── database.go       # Database operations and queries
├── auth.go           # Authentication and session management
├── handlers.go       # HTTP request handlers
├── templates.go      # HTML templates and page rendering
├── go.mod           # Go module dependencies
├── Dockerfile       # Docker container configuration
├── docker-compose.yml # Docker Compose configuration
└── README.md        # This file
```

## Security Features

- **Password Hashing**: All passwords are hashed using bcrypt
- **Session Management**: Secure session handling with UUID
- **Input Validation**: Form validation and sanitization
- **SQL Injection Protection**: Parameterized queries
- **CSRF Protection**: Session-based security

## Error Handling

The application includes comprehensive error handling:
- HTTP status codes for different error types
- User-friendly error messages
- Database error handling
- Authentication error responses

## Testing

To run tests (when implemented):
```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is open source and available under the [MIT License](LICENSE).

## Support

For issues and questions, please create an issue in the repository.

## Docker Deployment

# Running the App with Docker

## 1. Build the Docker Image

To build the Docker image, run:

```bash
docker build -t forum:latest .
```

This will build the Docker image and tag it as `forum:latest`.

To see the built images, run:

```bash
docker images
```

You should see output similar to:

```
REPOSITORY          TAG       IMAGE ID       CREATED          SIZE
forum               latest    85a65d66ca39   7 seconds ago    795MB
```

---

## 2. Run the Docker Container

To start a container using the image you just created, run:

```bash
docker run -d -p 8080:8080 --name forum forum:latest
```

- `-d` runs the container in detached mode
- `-p 8080:8080` maps port 8080 of the container to port 8080 on your host
- `--name forum` names your container "forum"

To see running containers, run:

```bash
docker ps -a
```

You should see output similar to:

```
CONTAINER ID   IMAGE         COMMAND      CREATED          STATUS          PORTS                    NAMES
cc8f5dcf760f   forum:latest  "./main"    6 seconds ago    Up 6 seconds    0.0.0.0:8080->8080/tcp   forum
```

---

## 3. Access the Application

Open your browser and go to [http://localhost:8080](http://localhost:8080) to use the forum application.

---

If you encounter any issues, check the container logs with:

```bash
docker logs forum
```