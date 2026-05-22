# Polyglot SMS Gateway Service

An asynchronous, event-driven microservice architecture for processing and storing SMS messages. This system utilizes a Java Spring Boot API for ingestion, Apache Kafka for decoupled message brokering, and a GoLang microservice for high-throughput database writes to MongoDB.

## Architecture Overview
1. **Ingestion (Java/Spring Boot):** Receives the SMS request, validates against a Redis blocklist, simulates vendor processing, and publishes an event.
2. **Message Bus (Apache Kafka):** Handles reliable, asynchronous message delivery between the Spring Boot application and the Go worker.
3. **Storage Worker (Go/MongoDB):** Constantly consumes events from Kafka and persists the SMS records into a MongoDB database.
4. **Retrieval (Go API):** Serves stored SMS history for specific users.

## Prerequisites
* Docker Desktop installed and running
* Ports 8080, 8081, 9092, 6379, and 27017 available

## How to Run Locally

1. Open your terminal in the root project directory.
2. Start the entire infrastructure using Docker Compose:
   ```bash
   docker-compose up --build -d