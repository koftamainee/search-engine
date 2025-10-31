FROM golang:1.25.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /crawler main.go


FROM alpine:latest

RUN addgroup -g 1000 -S crawler && \
  adduser -u 1000 -S crawler -G crawler

WORKDIR /home/crawler

COPY --from=builder --chown=crawler:crawler /crawler .

USER crawler
CMD ["./crawler"]
