package com.meesho.sms.service;

import com.meesho.sms.dto.SmsEvent;
import com.meesho.sms.dto.SmsRequest;
import com.meesho.sms.dto.SmsResponse;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;


@Service // Tells Spring this is a business logic class
public class SmsService {

    private final StringRedisTemplate redisTemplate;
    private final KafkaTemplate<String, Object> kafkaTemplate;
    private static final String KAFKA_TOPIC = "sms_events";

    // Constructor Injection
    public SmsService(StringRedisTemplate redisTemplate, KafkaTemplate<String, Object> kafkaTemplate) {
        this.redisTemplate = redisTemplate;
        this.kafkaTemplate = kafkaTemplate;
    }

    public SmsResponse processSms(SmsRequest request) {
        // Step 1: Check Redis Blocklist
        // We look for a set called "blocked_users" in Redis
        Boolean isBlocked = redisTemplate.opsForSet().isMember("blocked_users", request.getPhoneNumber());
        if (Boolean.TRUE.equals(isBlocked)) {
            return new SmsResponse("FAILED", "User is blocked from sending SMS.");
        }

        // Step 2: Mock 3rd Party Call
        // We use Math.random() to simulate an 80% success rate from the vendor
        boolean is3pSuccess = Math.random() > 0.2;
        String status = is3pSuccess ? "SUCCESS" : "FAILED";

        // Step 3: Send SMS Event to Kafka
        SmsEvent event = new SmsEvent(request.getPhoneNumber(), request.getMessage(), status);
        kafkaTemplate.send(KAFKA_TOPIC, event);

        // Step 4: Return Final Response to Client
        String responseMsg = is3pSuccess ? "SMS processed successfully." : "SMS failed at vendor.";
        return new SmsResponse(status, responseMsg);
    }
}