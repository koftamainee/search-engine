# Indexer

Rust service that indexes data collected by the crawler.

---

## Local Development

Install Rust if you haven't already: https://www.rust-lang.org/tools/install

Install cargo-watch for automatic rebuilds:

```bash
cargo install cargo-watch
```

Run the service locally:

```bash
# automatic rebuild on code changes
cargo watch -x run

# or run manually
cargo run

# run tests with
cargo test
```
