# Crawler Service Documentation

## Environment Variables

```env
REDIS_CRAWLER_PASSWORD=pass
REDIS_CRAWLER_HOST=redis-crawler
REDIS_PORT=6379

MEILI_MASTER_KEY=meilikey
MEILI_HOST=meilisearch
MEILI_PORT=7700
```

## Message Structure Sent to Indexer

Crawler sends documents to Meilisearch index `web_pages` in the following format:

```json
{
  "id": "md5(url)",
  "url": "https://example.com/page",
  "title": "Page title",
  "description": "Meta description",
  "text": "Normalized extracted page text",
  "timestamp": "2026-04-30T12:34:56Z"
}
```

### Field descriptions

* **id** — MD5 hash of page URL, unique document identifier.
* **url** — normalized crawled URL.
* **title** — extracted `<title>`.
* **description** — meta description (`og:description` or `description`).
* **text** — normalized extracted text (max 5000 chars).
* **timestamp** — crawl timestamp in RFC3339 UTC.

## Redis Structures

* `queue` — URLs waiting for processing
* `visited` — already crawled URLs
* `crawled_pages` — archived raw crawler messages
* `robots:<domain>` — cached robots.txt rules (96h TTL)

## Workflow

1. Normalize start URL and push to Redis queue.
2. Dispatcher reads URLs from queue.
3. Worker fetches page and checks robots.txt.
4. HTML is parsed into title, description, text, and links.
5. Links are normalized, filtered, and deduplicated.
6. New links are added to `visited` and `queue`.
7. Parsed page is indexed into Meilisearch (`web_pages`).
8. Graceful shutdown happens on SIGINT/SIGTERM or `/stop`.

---

## HTTP Management API

Crawler exposes management endpoints on port `8081`.

### Health check

```bash
GET /health
```

Response:

```json
{"status":"ok"}
```

### Crawler status

```bash
GET /status
```

Response:

```json
{"running":true}
```

### Start crawling

# url="" - error

```bash
GET /start?url=https://example.com
```

Starts crawler from provided seed URL.

### Stop crawling

```bash
GET /stop
```

Stops all workers gracefully.

---

## Configuration

### Required environment variables

```env
REDIS_CRAWLER_HOST=redis-crawler
REDIS_CRAWLER_PORT=6379
REDIS_CRAWLER_PASSWORD=pass

MEILI_HOST=meilisearch
MEILI_PORT=7700
MEILI_MASTER_KEY=meilikey

CRAWLER_NUM_WORKERS=5
```

If `CRAWLER_NUM_WORKERS` is missing or invalid, default value `5` is used.

---

## Run locally

```bash
go run main.go
```

Run tests:

```bash
go test ./...
```

Example API usage:

```bash
curl http://localhost:8081/health
curl "http://localhost:8081/start?url=https://example.com"
curl http://localhost:8081/status
curl http://localhost:8081/stop
```
