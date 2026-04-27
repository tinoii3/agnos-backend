# Agnos Back End

Starter backend scaffold using:
- Go
- Gin
- Docker
- Nginx
- PostgreSQL

## Project Structure

- `cmd/api`: application entrypoint
- `internal/config`: environment-based config
- `internal/db`: PostgreSQL connection setup
- `internal/http`: Gin router and handlers
- `nginx`: Nginx reverse proxy config
- `scripts`: DB bootstrap SQL

## Quick Start
1. Start everything:

```bash
docker compose up --build
```

## Endpoints

- `GET /healthz` - app liveness check
- `GET /readyz` - readiness check with Postgres ping

Because Nginx is in front of the app, access endpoints via:

- `http://localhost/healthz`
- `http://localhost/readyz`

## Run App Without Docker

Make sure PostgreSQL is available and `DATABASE_URL` is set, then:

```bash
go mod tidy
go run ./cmd/api
```
