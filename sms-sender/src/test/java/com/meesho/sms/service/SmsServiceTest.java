package com.meesho.sms.service;

import com.meesho.sms.dto.SmsEvent;
import com.meesho.sms.dto.SmsRequest;
import com.meesho.sms.dto.SmsResponse;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.redis.core.SetOperations;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.kafka.core.KafkaTemplate;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class SmsServiceTest {

    // 1. Create fake ("Mock") versions of our infrastructure
    @Mock
    private StringRedisTemplate redisTemplate;

    @Mock
    private SetOperations<String, String> setOperations;

    @Mock
    private KafkaTemplate<String, Object> kafkaTemplate;

    // 2. Inject those fake tools into the real Service we want to test
    @InjectMocks
    private SmsService smsService;

    @BeforeEach
    void setUp() {
        // Because SmsService calls redisTemplate.opsForSet(), we have to tell Mockito
        // to connect our fake RedisTemplate to our fake SetOperations.
        when(redisTemplate.opsForSet()).thenReturn(setOperations);
    }

    @Test
    void shouldReturnFailedWhenUserIsBlocked() {
        // ARRANGE: Set up the scenario
        SmsRequest request = new SmsRequest("9998887777", "Hello");
        // Tell the fake Redis to return TRUE (user is blocked)
        when(setOperations.isMember("blocked_users", "9998887777")).thenReturn(true);

        // ACT: Run the actual code
        SmsResponse response = smsService.processSms(request);

        // ASSERT: Prove the code made the right decisions
        assertEquals("FAILED", response.getStatus());
        assertEquals("User is blocked from sending SMS.", response.getMessage());
        
        // Prove that Kafka was NEVER triggered (saving money/bandwidth)
        verify(kafkaTemplate, never()).send(anyString(), any());
    }

    @Test
    void shouldSendToKafkaWhenUserIsNotBlocked() {
        // ARRANGE: Set up the scenario
        SmsRequest request = new SmsRequest("1112223333", "Hello World");
        // Tell the fake Redis to return FALSE (user is safe)
        when(setOperations.isMember("blocked_users", "1112223333")).thenReturn(false);

        // ACT: Run the actual code
        SmsResponse response = smsService.processSms(request);

        // ASSERT: Prove the code made the right decisions
        assertNotNull(response);
        
        // Prove that Kafka was triggered exactly 1 time with the "sms_events" topic
        verify(kafkaTemplate, times(1)).send(eq("sms_events"), any(SmsEvent.class));
    }
}