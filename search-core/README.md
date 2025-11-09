# Search core

Rust service that indexes data, sets it into database and listens for incoming request to get data from database.

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
