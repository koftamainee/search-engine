FROM rust:1.90-alpine AS builder

WORKDIR /app

RUN apk add --no-cache musl-dev gcc build-base

COPY . .
RUN cargo build --release


FROM alpine:latest 

RUN addgroup -g 1000 -S indexer && adduser -u 1000 -S indexer -G indexer

WORKDIR /home/indexer
COPY --from=builder --chown=indexer:indexer /app/target/release/indexer ./

USER indexer
CMD ["./indexer"]
