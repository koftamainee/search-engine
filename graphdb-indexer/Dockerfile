FROM golang:1.26.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /graphdb-indexer ./cmd/api/


FROM alpine:latest

WORKDIR /home/graphdb-indexer

COPY --from=builder /graphdb-indexer .

EXPOSE 8080
CMD ["./graphdb-indexer"]
