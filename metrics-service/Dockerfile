FROM rust:1.95-alpine AS builder

RUN apk add --no-cache musl-dev

WORKDIR /app
COPY . .

RUN cargo build --release

FROM alpine:3.23.4

COPY --from=builder /app/target/release/metrics-service /usr/local/bin/metrics-service

EXPOSE 8080
ENTRYPOINT ["metrics-service"]