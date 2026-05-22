package com.meesho.sms.controller;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.meesho.sms.dto.SmsRequest;
import com.meesho.sms.dto.SmsResponse;
import com.meesho.sms.service.SmsService;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import static org.mockito.ArgumentMatchers.any;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

@WebMvcTest(SmsController.class) // Only loads the web layer, ignoring databases/Kafka
class SmsControllerTest {

    @Autowired
    private MockMvc mockMvc; // Simulates Postman requests

    @Autowired
    private ObjectMapper objectMapper; // Converts Java objects to JSON

    @MockBean
    private SmsService smsService; // A fake version of your service

    @Test
    void shouldAcceptRequestAndReturnSuccess() throws Exception {
        // ARRANGE: Set up the fake request and what the fake service should return
        SmsRequest request = new SmsRequest("9998887777", "Test Web Message");
        SmsResponse mockResponse = new SmsResponse("SUCCESS", "SMS processed successfully.");
        Mockito.when(smsService.processSms(any(SmsRequest.class))).thenReturn(mockResponse);

        // ACT & ASSERT: Simulate a Postman POST request and verify the JSON response
        mockMvc.perform(post("/v1/sms/send")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.status").value("SUCCESS"))
                .andExpect(jsonPath("$.message").value("SMS processed successfully."));
    }
}