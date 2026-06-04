package com.meesho.sms.dto;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor // Generates a constructor with all parameters so we can easily create responses
public class SmsResponse {
    private String status;
    private String message;
}