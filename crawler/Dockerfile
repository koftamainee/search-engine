FROM golang:1.25.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go get golang.org/x/net/html
COPY . .
RUN go build -o /crawler main.go

# PROD stage
#FROM alpine:latest AS production
FROM alpine:latest
WORKDIR /home/crawler
COPY --from=builder /crawler .
CMD ["./crawler"]

EXPOSE 8080

# DEV stage
#FROM golang:1.25.3-alpine AS development
#WORKDIR /app
#RUN go install github.com/air-verse/air@latest
#COPY --from=builder  /crawler .
#COPY . .
#CMD ["air", "-c", ".air.toml"]
