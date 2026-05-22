package com.meesho.sms.exception;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.meesho.sms.controller.SmsController;
import com.meesho.sms.dto.SmsRequest;
import com.meesho.sms.service.SmsService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.data.redis.RedisConnectionFailureException;
import org.springframework.http.MediaType;
import org.springframework.kafka.KafkaException;
import org.springframework.test.web.servlet.MockMvc;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

@WebMvcTest({SmsController.class, GlobalExceptionHandler.class})
class GlobalExceptionHandlerTest {

    @Autowired
    private MockMvc mockMvc; // Simulates Postman

    @Autowired
    private ObjectMapper objectMapper;

    @MockBean
    private SmsService smsService; // Dummy service

    @Test
    void shouldReturnServiceUnavailableWhenRedisFails() throws Exception {
        // ARRANGE
        SmsRequest request = new SmsRequest("9998887777", "Hello");
        // We order our dummy service to intentionally crash with a Redis error
        when(smsService.processSms(any(SmsRequest.class)))
                .thenThrow(new RedisConnectionFailureException("Redis container disconnected!"));

        // ACT & ASSERT
        mockMvc.perform(post("/v1/sms/send")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(request)))
                // We expect a 503 Service Unavailable, NOT a 500 Crash
                .andExpect(status().isServiceUnavailable())
                .andExpect(jsonPath("$.status").value("FAILED"))
                .andExpect(jsonPath("$.error").value("Cache Service Unavailable"));
    }

    @Test
    void shouldReturnServiceUnavailableWhenKafkaFails() throws Exception {
        // ARRANGE
        SmsRequest request = new SmsRequest("9998887777", "Hello");
        // We order our dummy service to intentionally crash with a Kafka error
        when(smsService.processSms(any(SmsRequest.class)))
                .thenThrow(new KafkaException("Kafka broker down!"));

        // ACT & ASSERT
        mockMvc.perform(post("/v1/sms/send")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(request)))
                // We expect the custom JSON we designed
                .andExpect(status().isServiceUnavailable())
                .andExpect(jsonPath("$.status").value("FAILED"))
                .andExpect(jsonPath("$.error").value("Message Broker Unavailable"));
    }
}