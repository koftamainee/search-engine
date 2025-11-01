
# **Crawler → Indexer RabbitMQ Contract**

## **1. Queue**

* **Queue Name:** `crawler_queue`
* **Exchange:** default (direct)

**Connection settings:**

* Host: `rabbitmq` (Docker service name)
* Port: `5672`
* User/Password: use env variables `RABBITMQ_USER` / `RABBITMQ_PASSWORD`

**Example RabbitMQ URL:**

```
amqp://${RABBITMQ_USER}:${RABBITMQ_PASSWORD}@rabbitmq:5672/
```

---

## **2. Message Structure**

Each message represents a single crawled webpage with metadata.

```json
{
  "url": "https://example.com/page1",
  "text": "This is the main text of the page...",
  "metadata": {
    "title": "Example Page",
    "description": "Example description",
    "timestamp": "2025-11-01T12:00:00Z",
    "status_code": 200
  }
}
```

### **Fields Description**

| Field                  | Type              | Required | Description                    |
| ---------------------- | ----------------- | -------- | ------------------------------ |
| `url`                  | string            | yes      | Normalized URL of the page     |
| `text`                 | string            | yes      | Normalized lowercase content of the page in plain text  |
| `metadata`             | object            | yes      | Metadata about the page        |
| `metadata.title`       | string            | yes      | Page title from `<title>` tag  |
| `metadata.description` | string            | no       | Page meta description          |
| `metadata.timestamp`   | string (ISO 8601) | yes      | UTC time when page was crawled |
| `metadata.status_code` | integer           | yes      | HTTP response code             |

---

## **3. Message Handling Rules**

* **Crawler**

  * Publishes a message to `crawler_queue` for every successfully crawled page.
  * When the crawler fetches a page, it first checks Redis to see if the URL already exists. **If the URL is present, the crawler skips it; otherwise, it proceeds to fetch the page.**
  * After crawling, the URL is stored in Redis with a time-to-live (TTL) representing the minimum delay before the page can be crawled again.
  * Once the TTL expires and the URL entry is removed from Redis, **the crawler is allowed to fetch that URL again.**
  * If a page fails, do **not publish**; retry logic is handled internally.
  * Ensure `url` is normalized (lowercase, remove fragments, etc.).
  * Ensure `text` is normalized (lowercase, remove html tags, collapse multiple spaces, newlines, etc.)

* **Indexer**

  * Consumes messages from `crawler_queue`.
  * Must **acknowledge messages only after successfully storing in the local search DB**.
  * If processing fails, the message should **remain in the queue for retry**.
  * Indexer **should not check** for uniqueness of given URL, it process it in any case.

---

## **4. JSON Schema (Optional for Validation)**

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "CrawlerMessage",
  "type": "object",
  "properties": {
    "url": { "type": "string", "format": "uri" },
    "text": { "type": "string" },
    "metadata": {
      "type": "object",
      "properties": {
        "title": { "type": "string" },
        "description": { "type": "string" },
        "timestamp": { "type": "string", "format": "date-time" },
        "status_code": { "type": "integer" }
      },
      "required": ["title", "timestamp", "status_code"]
    }
  },
  "required": ["url", "html", "metadata"]
}
```

---

## **5. Example Flow**

1. **Crawler** crawls `https://example.com/page1`.
2. Checks Redis to see if it was already crawled.
3. Prepares message:

```json
{
  "url": "https://example.com/page1",
  "text": "Example Text",
  "metadata": {
    "title": "Example Page",
    "description": "An example",
    "timestamp": "2025-11-01T12:00:00Z",
    "status_code": 200
  }
}
```

4. Publishes message to `crawler_queue`.
5. **Indexer** receives message.
6. Stores data in the local search DB.
7. Acknowledges RabbitMQ message.
