package com.meesho.sms.dto;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class SmsEvent {
    private String requestId;
    private String phoneNumber;
    private String message;
    private String status;
    private String timestamp;
}