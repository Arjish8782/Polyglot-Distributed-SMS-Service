package com.meesho.sms.service;

import com.meesho.sms.dto.SmsEvent;
import com.meesho.sms.dto.SmsRequest;
import com.meesho.sms.dto.SmsResponse;
import com.meesho.sms.exception.UserBlockedException;
import com.meesho.sms.exception.VendorFailureException;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.redis.core.SetOperations;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.kafka.core.KafkaTemplate;

import java.util.concurrent.CompletableFuture;
import java.util.function.BooleanSupplier;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class SmsServiceTest {

    @Mock
    private StringRedisTemplate redisTemplate;

    @Mock
    private SetOperations<String, String> setOperations;

    @Mock
    private KafkaTemplate<String, Object> kafkaTemplate;

    @Mock
    private BooleanSupplier vendorSimulator;

    @InjectMocks
    private SmsService smsService;

    @BeforeEach
    void setUp() {
        when(redisTemplate.opsForSet()).thenReturn(setOperations);
    }

    @Test
    void shouldThrowUserBlockedExceptionWhenUserIsBlocked() {
        SmsRequest request = new SmsRequest("9998887777", "Hello");
        when(setOperations.isMember("blocked_users", "9998887777")).thenReturn(true);

        assertThrows(UserBlockedException.class, () -> smsService.processSms(request));
        verify(kafkaTemplate, never()).send(anyString(), any());
    }

    @Test
    void shouldReturnSuccessAndSendToKafkaWhenVendorSucceeds() {
        SmsRequest request = new SmsRequest("1112223333", "Hello World");
        when(setOperations.isMember("blocked_users", "1112223333")).thenReturn(false);
        when(vendorSimulator.getAsBoolean()).thenReturn(true);
        when(kafkaTemplate.send(anyString(), any())).thenReturn(CompletableFuture.completedFuture(null));

        SmsResponse response = smsService.processSms(request);

        assertEquals("SUCCESS", response.getStatus());
        assertEquals("SMS processed successfully.", response.getMessage());
        verify(kafkaTemplate, times(1)).send(eq("sms_events"), any(SmsEvent.class));
    }

    @Test
    void shouldThrowVendorFailureExceptionAndStillPublishToKafkaWhenVendorFails() {
        SmsRequest request = new SmsRequest("1112223333", "Hello World");
        when(setOperations.isMember("blocked_users", "1112223333")).thenReturn(false);
        when(vendorSimulator.getAsBoolean()).thenReturn(false);
        when(kafkaTemplate.send(anyString(), any())).thenReturn(CompletableFuture.completedFuture(null));

        assertThrows(VendorFailureException.class, () -> smsService.processSms(request));
        // Event is still published to Kafka even when vendor fails (for audit trail)
        verify(kafkaTemplate, times(1)).send(eq("sms_events"), any(SmsEvent.class));
    }
}