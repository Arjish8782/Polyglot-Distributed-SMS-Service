package com.meesho.sms.dto;

import lombok.Data;

@Data // This Lombok annotation automatically generates getters and setters
public class SmsRequest {
    private String phoneNumber;
    private String message;

    // Default constructor (required by Spring)
public SmsRequest() {}

// All-arguments constructor (used for our tests)
public SmsRequest(String phoneNumber, String message) {
    this.phoneNumber = phoneNumber;
    this.message = message;
}
}
   