FROM golang:1.25.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /backend ./cmd/api/


FROM alpine:latest

WORKDIR /home/backend

COPY --from=builder /backend .

COPY config/ config/

EXPOSE 8080
CMD ["./backend"]
