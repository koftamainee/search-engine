FROM rust:1.90-alpine AS builder

WORKDIR /app

RUN apk add --no-cache musl-dev gcc build-base

COPY . .
RUN cargo build --release


FROM alpine:latest 

RUN addgroup -g 1000 -S search-core && \
  adduser -u 1000 -S search-core -G search-core

WORKDIR /home/search-core
COPY --from=builder --chown=indexer:indexer /app/target/release/search-core ./

USER search-core
CMD ["./search-core"]
