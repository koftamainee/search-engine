FROM golang:1.25.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /backend main.go


FROM alpine:latest

WORKDIR /home/backend

COPY --from=builder --chown=backend:backend /backend .

EXPOSE 8080
CMD ["./backend"]
