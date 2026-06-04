package com.meesho.sms.exception;

import org.springframework.data.redis.RedisConnectionFailureException;
import org.springframework.kafka.KafkaException;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;

import java.util.HashMap;
import java.util.Map;

@ControllerAdvice
public class GlobalExceptionHandler {

    // 🛑 1. Catch Redis Cache Failures
    @ExceptionHandler(RedisConnectionFailureException.class)
    public ResponseEntity<Map<String, String>> handleRedisTimeout(RedisConnectionFailureException ex) {
        Map<String, String> errorResponse = new HashMap<>();
        errorResponse.put("status", "FAILED");
        errorResponse.put("error", "Cache Service Unavailable");
        errorResponse.put("message", "Our system is experiencing high traffic.");
        return new ResponseEntity<>(errorResponse, HttpStatus.SERVICE_UNAVAILABLE);
    }

    // 🛑 2. Catch Kafka Broker Failures
    @ExceptionHandler(KafkaException.class)
    public ResponseEntity<Map<String, String>> handleKafkaError(KafkaException ex) {
        Map<String, String> errorResponse = new HashMap<>();
        errorResponse.put("status", "FAILED");
        errorResponse.put("error", "Message Broker Unavailable");
        errorResponse.put("message", "Unable to route your SMS at this time.");
        return new ResponseEntity<>(errorResponse, HttpStatus.SERVICE_UNAVAILABLE);
    }

    // 🛡️ 3. The Universal Catch-All for unexpected bugs
    @ExceptionHandler(Exception.class)
    public ResponseEntity<Map<String, String>> handleGenericError(Exception ex) {
        Map<String, String> errorResponse = new HashMap<>();
        errorResponse.put("status", "FAILED");
        errorResponse.put("error", "Internal Server Error");
        errorResponse.put("message", "An unexpected error occurred.");
        return new ResponseEntity<>(errorResponse, HttpStatus.INTERNAL_SERVER_ERROR);
    }
}