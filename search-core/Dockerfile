FROM rust:1.90-alpine AS builder

WORKDIR /app

RUN apk add --no-cache musl-dev gcc build-base

COPY . .
RUN cargo build --release


FROM alpine:latest 

WORKDIR /home/search-core
COPY --from=builder --chown=indexer:indexer /app/target/release/search-core ./

CMD ["./search-core"]
