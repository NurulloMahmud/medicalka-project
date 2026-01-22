# Medicalka Social Network API

A mini social network backend with users, posts, comments, likes, JWT authentication, and email verification.

## Tech Stack

- Go 1.23
- PostgreSQL 16
- Chi Router
- JWT Authentication
- Docker

## Quick Start

```bash
docker compose up --build
```

The API will be available at `http://localhost:8080`

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| ENV | development | Environment mode |
| SERVER_ADDRESS | :8080 | Server port |
| DATABASE_URL | postgres://postgres:postgres@localhost:5432/medicalka?sslmode=disable | PostgreSQL connection string |
| JWT_SECRET | (random) | Secret key for JWT tokens |
| LIMITER_RPS | 2 | Rate limit requests per second |
| LIMITER_BURST | 4 | Rate limit burst size |
| LIMITER_ENABLED | true | Enable rate limiting |
| SMTP_HOST | smtp.gmail.com | SMTP server host |
| SMTP_PORT | 587 | SMTP server port |
| SMTP_USERNAME | | SMTP username |
| SMTP_PASSWORD | | SMTP password |
| SMTP_FROM | | Sender email address |

## Project Structure

```
.
├── cmd/
│   ├── api/            # Main application entry point
│   └── cleanup/        # Cleanup unverified users command
├── config/             # Configuration loading
├── internal/
│   ├── auth/           # JWT token handling
│   ├── comment/        # Comment domain (handler, service, repository)
│   ├── like/           # Like domain (handler, service, repository)
│   ├── middleware/     # Authentication and rate limiting
│   ├── platform/       # Database connection
│   ├── post/           # Post domain (handler, service, repository)
│   ├── server/         # Application setup and routes
│   ├── tasks/          # Background tasks (email sending)
│   └── user/           # User domain (handler, service, repository)
├── migrations/         # Database migrations
├── utils/              # Shared utilities
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## API Endpoints

### Authentication

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | /api/auth/register | Register new user | No |
| POST | /api/auth/login | Login and get token | No |
| GET | /api/auth/me | Get current user | Yes |
| PATCH | /api/auth/me | Update current user | Yes |
| GET | /api/verify-email?token= | Verify email address | Yes |

### Posts

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | /api/posts | List posts with pagination | No |
| GET | /api/posts/{id} | Get post with comments | No |
| POST | /api/posts | Create post | Yes (verified) |
| PATCH | /api/posts/{id} | Update post | Yes (author) |
| DELETE | /api/posts/{id} | Delete post | Yes (author) |

### Comments

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | /api/posts/{id}/comments | Add comment | Yes (verified) |
| DELETE | /api/posts/{postID}/comments/{commentID} | Delete comment | Yes (author) |

### Likes

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | /api/posts/{id}/like | Like post | Yes |
| DELETE | /api/posts/{id}/like | Remove like | Yes |

### Feed

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | /api/feed | Get users with their posts and likes | No |

## Request Examples

### Register

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "johndoe",
    "full_name": "John Doe",
    "password": "secret123"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "secret123"
  }'
```

### Create Post

```bash
curl -X POST http://localhost:8080/api/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "title": "My First Post",
    "content": "Hello world!"
  }'
```

### Get Posts with Filtering

```bash
curl "http://localhost:8080/api/posts?page=1&page_size=10&search=hello&date_from=2024-01-01T00:00:00Z&date_to=2024-12-31T23:59:59Z"
```

### Like Post

```bash
curl -X POST http://localhost:8080/api/posts/{id}/like \
  -H "Authorization: Bearer <token>"
```

### Get Feed

```bash
curl "http://localhost:8080/api/feed?page=1&page_size=10"
```

## Access Control

| User State | Permissions |
|------------|-------------|
| Not logged in | View posts, view feed |
| Logged in (unverified) | View posts, like posts |
| Logged in (verified) | Create/edit/delete own posts, create/delete own comments |

## Cleanup Command

Remove unverified users older than 24 hours:

```bash
go run cmd/cleanup/main.go
```

With Docker:

```bash
docker compose exec app ./cleanup
```

## Running Tests

```bash
createdb medicalka_test

TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/medicalka_test?sslmode=disable" go test ./internal/user/... -v
```

## Local Development

```bash
createdb medicalka

go run cmd/api/main.go
```