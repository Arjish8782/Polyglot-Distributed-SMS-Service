#!/bin/bash

# Define the test phone number
PHONE_NUMBER="9998887779"

echo "🚀 Starting End-to-End Flow Demonstration..."
echo "------------------------------------------------"

# Step 1: Call the Java Service
echo "1️⃣ Sending POST request to Java Spring Boot Service (Port 8080)..."
curl -s -X POST http://localhost:8080/v1/sms/send \
  -H "Content-Type: application/json" \
  -d '{"phoneNumber": "'$PHONE_NUMBER'", "message": "Automated E2E Test Message"}' | jq .

echo -e "\n⏳ Waiting 2 seconds for Kafka propagation and MongoDB persistence...\n"
sleep 2

# Step 2: Check the GoLang Service Logs
echo "2️⃣ Checking GoLang Service (go-store) logs for internal Kafka consumption..."
# Fetch the last 5 lines of the Go container logs to prove it caught and processed the message
docker logs --tail 5 go-store

echo -e "\n------------------------------------------------\n"

# Step 3: Retrieve the record using the GoLang History API
echo "3️⃣ Fetching saved record from GoLang API (Port 8081)..."
curl -s -X GET http://localhost:8081/v1/user/$PHONE_NUMBER/messages | jq .

echo -e "\n✅ End-to-End Demonstration Complete!"