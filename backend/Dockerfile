FROM golang:1.25.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /backend main.go


FROM alpine:latest

RUN addgroup -g 1000 -S backend && \
  adduser -u 1000 -S backend -G backend

WORKDIR /home/backend

COPY --from=builder --chown=backend:backend /backend .

EXPOSE 8080
USER backend
CMD ["./backend"]
