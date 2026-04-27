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

1. (Optional) create `.env` from `.env.example`.
2. Start everything:

```bash
docker compose up --build
```

## Endpoints

- `GET /status` - app liveness check
- `POST /staff/create` - create staff (`username`, `password`, `hospital`)
- `POST /staff/login` - login and return JWT token (`username`, `password`, `hospital`)
- `GET /patient/search` - patients matching criteria in same hospital as logged-in staff

`/patient/search` supports optional filters:

- `id` (can be either `national_id` or `passport_id`)
- `national_id`
- `passport_id`
- `first_name`
- `middle_name`
- `last_name`
- `date_of_birth`
- `phone_number`
- `email`

Because Nginx is in front of the app, access endpoints via:

- `http://localhost/status`
- `http://localhost/patient/search`

## Test

```bash
go test ./...
```

## Run App Without Docker

Make sure PostgreSQL is available and `DATABASE_URL` is set, then:

```bash
go mod tidy
go run ./cmd/api
```
