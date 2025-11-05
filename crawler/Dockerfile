FROM golang:1.25.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /crawler main.go

# PROD stage
FROM alpine:latest AS production
WORKDIR /home/crawler
COPY --from=builder /crawler .
CMD ["./crawler"]

# DEV stage
FROM golang:1.25.3-alpine AS development
WORKDIR /app
RUN go install github.com/air-verse/air@latest
COPY --from=builder  /crawler .
COPY . .
CMD ["air", "-c", ".air.toml"]
