package com.meesho.sms.dto;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class SmsEvent {
    private String phoneNumber;
    private String message;
    private String status; // Will hold "SUCCESS" or "FAILED"
}