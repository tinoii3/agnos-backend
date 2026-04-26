FROM golang:1.26.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/api

FROM alpine:3.20

WORKDIR /app
RUN adduser -D appuser

COPY --from=builder /bin/server /app/server

USER appuser
EXPOSE 8080

CMD ["/app/server"]
