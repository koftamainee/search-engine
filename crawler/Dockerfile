FROM golang:1.25.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /crawler main.go


FROM alpine:latest

WORKDIR /home/crawler

COPY --from=builder --chown=crawler:crawler /crawler .

CMD ["./crawler"]
