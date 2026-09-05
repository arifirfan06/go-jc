# Go Movies API Backend

A Go RESTful API backend built with **Gin Web Framework** and **GORM ORM**, connected to a **PostgreSQL** database.

---

## 🛠️ Stack & Technologies

- **Language**: Go 1.20+
- **Framework**: [Gin Web Framework](https://github.com/gin-gonic/gin)
- **ORM**: [GORM](https://gorm.io/)
- **Database**: PostgreSQL (`postgres:latest`)
- **Authentication**: JWT (JSON Web Tokens) with HttpOnly refresh cookies
- **Containerization**: Docker & Docker Compose

---

## 📋 Prerequisites

Before running the application, ensure you have installed:
- [Go](https://go.dev/doc/install) (v1.20 or later)
- [Docker](https://docs.docker.com/get-docker/) & Docker Compose

---

## 🚀 How to Run the Application

### 1. Start the PostgreSQL Database Container

Run Docker Compose from the project root to start PostgreSQL and automatically seed the schema from `sql/create_tables.sql`:

```bash
docker compose up -d
```

To verify the database container is running cleanly:

```bash
docker compose ps
```

---

### 2. Install Go Dependencies

Run `go mod tidy` to download all required modules:

```bash
go mod tidy
```

---

### 3. Run the API Server

> ⚠️ **Important**: Because the `cmd/api` package contains multiple Go source files (`main.go`, `handlers.go`, `routes.go`, `db.go`, `auth.go`, `middleware.go`, `utils.go`), execute the entire package rather than running `main.go` individually.

#### Option A: From the repository root
```bash
go run ./cmd/api
```

#### Option B: From the `cmd/api` directory
```bash
cd cmd/api
go run .
```

The server will start on port `8080`:
```text
Connected to Postgres via GORM!
Starting application on port 8080
[GIN-debug] GET    /                         --> main.(*application).Home-fm (3 handlers)
...
```

---

## ⚙️ Configuration & Flags

You can customize runtime parameters via command line flags:

| Flag | Default Value | Description |
| --- | --- | --- |
| `-dsn` | `"host=localhost port=5432 user=postgres password=postgres dbname=movies sslmode=disable timezone=UTC connect_timeout=5"` | PostgreSQL DSN |
| `-jwt-secret` | `"verysecret"` | Secret used to sign JWT tokens |
| `-jwt-issuer` | `"example.com"` | JWT issuer claim |
| `-jwt-audience` | `"example.com"` | JWT audience claim |
| `-cookie-domain` | `"localhost"` | Cookie domain for refresh tokens |
| `-domain` | `"example.com"` | Application domain |
| `-api-key` | `"b41447e6319d1cd467306735632ba733"` | MovieDB API Key |

Example specifying a custom secret or DSN:
```bash
go run ./cmd/api -jwt-secret="my-custom-jwt-secret" -dsn="host=localhost port=5432 user=postgres password=postgres dbname=movies sslmode=disable"
```

---

## 📌 API Endpoints

### Public Endpoints
- `GET /` - API Status check
- `POST /authenticate` - User authentication (returns JWT & sets refresh cookie)
- `GET /refresh` - Refresh access token using cookie
- `GET /logout` - Logout (clears refresh cookie)
- `GET /movies` - Get list of all movies
- `GET /movies/:id` - Get movie by ID
- `GET /genres` - Get list of all movie genres
- `GET /movies/genres/:id` - Get movies filtered by genre ID

### Admin Endpoints *(Requires `Authorization: Bearer <token>`)*
- `GET /admin/movies` - Catalog view of all movies
- `GET /admin/movies/:id` - Get movie payload for edit
- `PUT /admin/movies/0` - Create new movie
- `PATCH /admin/movies/:id` - Update existing movie
- `DELETE /admin/movies/:id` - Delete movie by ID

---

## 🧪 Testing the API

You can test the health check endpoint using `curl`:

```bash
curl http://localhost:8080/
```

Expected Response:
```json
{
  "status": "active",
  "message": "Go Movies up and running",
  "version": "1.0.0"
}
```

To authenticate:
```bash
curl -X POST http://localhost:8080/authenticate \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "password"}'
```
