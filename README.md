# Polyglot SMS Gateway Service

An asynchronous, event-driven microservice architecture for processing and storing SMS messages. This system uses a Java Spring Boot API for ingestion, Apache Kafka for decoupled message brokering, and a Go microservice for high-throughput database writes to MongoDB.

## Architecture Overview

1. **Ingestion (Java / Spring Boot):** Validates the request (phone number format, message length), checks against a Redis blocklist, simulates vendor processing (80% success / 20% failure), and publishes an event to Kafka with an idempotency `requestId` and `timestamp`.
2. **Message Bus (Apache Kafka):** Handles reliable, asynchronous message delivery between the Spring Boot application and the Go worker.
3. **Storage Worker (Go / MongoDB):** Consumes events from Kafka. Only SUCCESS events are persisted — FAILED vendor events are acknowledged but not stored. Implements exponential backoff retry on database failures.
4. **Retrieval (Go API):** Serves paginated SMS history for specific users.

## Prerequisites

- Docker Desktop installed and running
- Ports **8080**, **8081**, and **9092** available on the host

> **Note:** MongoDB (27017) and Redis (6379) are internal to the Docker network and are not exposed to the host.

## How to Run Locally

Open a terminal in the project root directory and start the entire stack:

```bash
docker-compose up --build -d
```

Services start in dependency order via Docker healthchecks — no manual wait needed. The stack is ready when `docker-compose ps` shows all services as `Up`.

## API Documentation

### 1. Send SMS (Java Service)

Ingests a new SMS request and queues it for processing.

| | |
|---|---|
| **URL** | `http://localhost:8080/v1/sms/send` |
| **Method** | `POST` |
| **Content-Type** | `application/json` |

**Request body:**

```json
{
  "phoneNumber": "9998887779",
  "message": "Hello from the polyglot gateway!",
  "requestId": "optional-client-idempotency-key"
}
```

> `requestId` is optional. If omitted, the server generates a UUID automatically.

**Validation rules:**
- `phoneNumber` — required, must be a valid 10-digit Indian mobile number (starts with 6–9)
- `message` — required, 1–160 characters

**Responses:**

| Status | Meaning |
|--------|---------|
| `200 OK` | SMS accepted and successfully published to Kafka |
| `400 Bad Request` | Validation failed (missing or invalid fields) |
| `422 Unprocessable Entity` | User is blocked from sending SMS |
| `502 Bad Gateway` | Third-party vendor call failed |
| `503 Service Unavailable` | Redis or Kafka is unreachable |

**Success response (200):**
```json
{
  "status": "SUCCESS",
  "message": "SMS processed successfully."
}
```

**Validation error response (400):**
```json
{
  "status": "FAILED",
  "error": "Validation Failed",
  "message": "phoneNumber: Must be a valid 10-digit Indian mobile number"
}
```

---

### 2. Get SMS History (Go Service)

Retrieves the paginated message history for a specific phone number. Only successfully delivered messages are returned.

| | |
|---|---|
| **URL** | `http://localhost:8081/v1/user/{phoneNumber}/messages` |
| **Method** | `GET` |

**Query parameters:**

| Parameter | Default | Max | Description |
|-----------|---------|-----|-------------|
| `page` | `1` | — | Page number |
| `limit` | `20` | `100` | Records per page |

**Example:**
```
GET http://localhost:8081/v1/user/9998887779/messages?page=1&limit=5
```

**Success response (200 OK):**

```json
[
  {
    "phoneNumber": "9998887779",
    "message": "Hello from the polyglot gateway!",
    "status": "SUCCESS",
    "requestId": "de3dc6a5-b8d0-4933-8b74-71391391e299",
    "timestamp": "2026-06-04T08:11:58.803732590Z",
    "createdAt": "2026-06-04T08:11:59.669Z"
  }
]
```

Results are sorted newest-first by `createdAt`.

## Author

Arjish Chowdhury
