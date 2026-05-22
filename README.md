# Polyglot SMS Gateway Service

An asynchronous, event-driven microservice architecture for processing and storing SMS messages. This system utilizes a Java Spring Boot API for ingestion, Apache Kafka for decoupled message brokering, and a Go microservice for high-throughput database writes to MongoDB.

## Architecture Overview

1. **Ingestion (Java / Spring Boot):** Receives the SMS request, validates against a Redis blocklist, simulates vendor processing, and publishes an event.
2. **Message Bus (Apache Kafka):** Handles reliable, asynchronous message delivery between the Spring Boot application and the Go worker.
3. **Storage Worker (Go / MongoDB):** Consumes events from Kafka and persists SMS records into MongoDB.
4. **Retrieval (Go API):** Serves stored SMS history for specific users.

## Prerequisites

- Docker Desktop installed and running
- Ports **8080**, **8081**, **9092**, **6379**, and **27017** available on the host

## How to Run Locally

1. Open a terminal in the project root directory.
2. Start the entire stack with Docker Compose:

   ```bash
   docker-compose up --build -d
   ```

3. Wait approximately 15–20 seconds for Kafka and Zookeeper to elect a leader and stabilize.

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
  "message": "Hello from the polyglot gateway!"
}
```

### 2. Get SMS History (Go Service)

Retrieves the message history for a specific phone number.

| | |
|---|---|
| **URL** | `http://localhost:8081/v1/user/{phoneNumber}/messages` |
| **Method** | `GET` |

Replace `{phoneNumber}` with the target number (for example, `9998887779`).

**Success response (200 OK):**

```json
[
  {
    "phoneNumber": "9998887779",
    "message": "Hello from the polyglot gateway!",
    "status": "SUCCESS"
  }
]
```

## Author

Arjish Chowdhury
