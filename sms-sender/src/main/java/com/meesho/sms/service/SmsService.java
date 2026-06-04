package com.meesho.sms.service;

import com.meesho.sms.dto.SmsEvent;
import com.meesho.sms.dto.SmsRequest;
import com.meesho.sms.dto.SmsResponse;
import com.meesho.sms.exception.UserBlockedException;
import com.meesho.sms.exception.VendorFailureException;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.kafka.KafkaException;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.UUID;
import java.util.concurrent.ExecutionException;
import java.util.function.BooleanSupplier;

@Service
public class SmsService {

    private final StringRedisTemplate redisTemplate;
    private final KafkaTemplate<String, Object> kafkaTemplate;
    private final BooleanSupplier vendorSimulator;
    private static final String KAFKA_TOPIC = "sms_events";

    public SmsService(StringRedisTemplate redisTemplate,
                      KafkaTemplate<String, Object> kafkaTemplate,
                      BooleanSupplier vendorSimulator) {
        this.redisTemplate = redisTemplate;
        this.kafkaTemplate = kafkaTemplate;
        this.vendorSimulator = vendorSimulator;
    }

    public SmsResponse processSms(SmsRequest request) {
        // Step 1: Check Redis Blocklist
        Boolean isBlocked = redisTemplate.opsForSet().isMember("blocked_users", request.getPhoneNumber());
        if (Boolean.TRUE.equals(isBlocked)) {
            throw new UserBlockedException(request.getPhoneNumber());
        }

        // Step 2: Mock 3rd Party Vendor Call (80% success, 20% failure)
        boolean is3pSuccess = vendorSimulator.getAsBoolean();
        String status = is3pSuccess ? "SUCCESS" : "FAILED";

        // Step 3: Send SMS Event to Kafka (published regardless of vendor outcome for audit trail)
        String requestId = (request.getRequestId() != null && !request.getRequestId().isBlank())
                ? request.getRequestId()
                : UUID.randomUUID().toString();
        SmsEvent event = new SmsEvent(requestId, request.getPhoneNumber(), request.getMessage(), status, Instant.now().toString());
        try {
            kafkaTemplate.send(KAFKA_TOPIC, event).get();
        } catch (ExecutionException e) {
            throw new KafkaException("Failed to publish SMS event to Kafka", e.getCause());
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new KafkaException("Kafka publish interrupted", e);
        }

        // Step 4: Return success or throw vendor failure
        if (!is3pSuccess) {
            throw new VendorFailureException();
        }
        return new SmsResponse("SUCCESS", "SMS processed successfully.");
    }
}