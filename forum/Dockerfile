# Build stage
FROM golang:1.24-alpine AS builder

# Install git, ca-certificates, and build tools for CGO
RUN apk add --no-cache git ca-certificates gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Install SQLite
RUN apk --no-cache add sqlite

# Create app directory
WORKDIR /forum/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy templates folder and forum.db from builder stage
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/forum.db ./forum.db

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./main"] 