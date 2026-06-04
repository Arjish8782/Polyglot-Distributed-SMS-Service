package com.meesho.sms.exception;

import org.springframework.data.redis.RedisConnectionFailureException;
import org.springframework.kafka.KafkaException;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;

import java.util.HashMap;
import java.util.Map;

@ControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ResponseEntity<Map<String, String>> handleValidationError(MethodArgumentNotValidException ex) {
        String firstError = ex.getBindingResult().getFieldErrors().stream()
                .map(fe -> fe.getField() + ": " + fe.getDefaultMessage())
                .findFirst()
                .orElse("Invalid request");
        Map<String, String> errorResponse = new HashMap<>();
        errorResponse.put("status", "FAILED");
        errorResponse.put("error", "Validation Failed");
        errorResponse.put("message", firstError);
        return new ResponseEntity<>(errorResponse, HttpStatus.BAD_REQUEST);
    }

    @ExceptionHandler(UserBlockedException.class)
    public ResponseEntity<Map<String, String>> handleUserBlocked(UserBlockedException ex) {
        Map<String, String> errorResponse = new HashMap<>();
        errorResponse.put("status", "FAILED");
        errorResponse.put("error", "User Blocked");
        errorResponse.put("message", ex.getMessage());
        return new ResponseEntity<>(errorResponse, HttpStatus.UNPROCESSABLE_ENTITY);
    }

    @ExceptionHandler(VendorFailureException.class)
    public ResponseEntity<Map<String, String>> handleVendorFailure(VendorFailureException ex) {
        Map<String, String> errorResponse = new HashMap<>();
        errorResponse.put("status", "FAILED");
        errorResponse.put("error", "Vendor Failure");
        errorResponse.put("message", ex.getMessage());
        return new ResponseEntity<>(errorResponse, HttpStatus.BAD_GATEWAY);
    }

    @ExceptionHandler(RedisConnectionFailureException.class)
    public ResponseEntity<Map<String, String>> handleRedisTimeout(RedisConnectionFailureException ex) {
        Map<String, String> errorResponse = new HashMap<>();
        errorResponse.put("status", "FAILED");
        errorResponse.put("error", "Cache Service Unavailable");
        errorResponse.put("message", "Our system is experiencing high traffic.");
        return new ResponseEntity<>(errorResponse, HttpStatus.SERVICE_UNAVAILABLE);
    }

    @ExceptionHandler(KafkaException.class)
    public ResponseEntity<Map<String, String>> handleKafkaError(KafkaException ex) {
        Map<String, String> errorResponse = new HashMap<>();
        errorResponse.put("status", "FAILED");
        errorResponse.put("error", "Message Broker Unavailable");
        errorResponse.put("message", "Unable to route your SMS at this time.");
        return new ResponseEntity<>(errorResponse, HttpStatus.SERVICE_UNAVAILABLE);
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<Map<String, String>> handleGenericError(Exception ex) {
        Map<String, String> errorResponse = new HashMap<>();
        errorResponse.put("status", "FAILED");
        errorResponse.put("error", "Internal Server Error");
        errorResponse.put("message", "An unexpected error occurred.");
        return new ResponseEntity<>(errorResponse, HttpStatus.INTERNAL_SERVER_ERROR);
    }
}