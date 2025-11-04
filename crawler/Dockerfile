# FROM golang:1.25.3-alpine AS builder

# WORKDIR /app

# RUN go install github.com/air-verse/air@latest

# COPY go.mod go.sum ./
# RUN go mod download

# COPY . .
# RUN go build -o /crawler main.go

# RUN cp /go/bin/air /air

# #FROM alpine:latest
# FROM golang:1.25.3-alpine

# WORKDIR /home/crawler

# COPY --from=builder --chown=crawler:crawler /crawler .
# COPY --from=builder --chown=crawler:crawler /air /usr/local/bin/air

# COPY --from=builder --chown=crawler:crawler /app/.air.toml ./

# #CMD ["./crawler"]
# CMD ["air", "-c", ".air.toml"]

# crawler/Dockerfile
FROM golang:1.25.3-alpine

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["air", "-c", ".air.toml"]
