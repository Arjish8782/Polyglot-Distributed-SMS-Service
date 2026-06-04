package com.meesho.sms.controller;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.meesho.sms.dto.SmsRequest;
import com.meesho.sms.dto.SmsResponse;
import com.meesho.sms.exception.GlobalExceptionHandler;
import com.meesho.sms.exception.UserBlockedException;
import com.meesho.sms.exception.VendorFailureException;
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

@WebMvcTest({SmsController.class, GlobalExceptionHandler.class})
class SmsControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @Autowired
    private ObjectMapper objectMapper;

    @MockBean
    private SmsService smsService;

    @Test
    void shouldAcceptRequestAndReturnSuccess() throws Exception {
        SmsRequest request = new SmsRequest("9998887777", "Test Web Message");
        SmsResponse mockResponse = new SmsResponse("SUCCESS", "SMS processed successfully.");
        Mockito.when(smsService.processSms(any(SmsRequest.class))).thenReturn(mockResponse);

        mockMvc.perform(post("/v1/sms/send")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.status").value("SUCCESS"))
                .andExpect(jsonPath("$.message").value("SMS processed successfully."));
    }

    @Test
    void shouldReturn422WhenUserIsBlocked() throws Exception {
        SmsRequest request = new SmsRequest("9998887777", "Test");
        Mockito.when(smsService.processSms(any(SmsRequest.class)))
                .thenThrow(new UserBlockedException("9998887777"));

        mockMvc.perform(post("/v1/sms/send")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isUnprocessableEntity())
                .andExpect(jsonPath("$.status").value("FAILED"))
                .andExpect(jsonPath("$.error").value("User Blocked"));
    }

    @Test
    void shouldReturn502WhenVendorFails() throws Exception {
        SmsRequest request = new SmsRequest("9998887777", "Test");
        Mockito.when(smsService.processSms(any(SmsRequest.class)))
                .thenThrow(new VendorFailureException());

        mockMvc.perform(post("/v1/sms/send")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isBadGateway())
                .andExpect(jsonPath("$.status").value("FAILED"))
                .andExpect(jsonPath("$.error").value("Vendor Failure"));
    }
}