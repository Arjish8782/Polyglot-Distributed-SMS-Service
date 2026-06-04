package com.meesho.sms.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;
import jakarta.validation.constraints.Size;
import lombok.Data;

@Data
public class SmsRequest {

    @NotBlank(message = "Phone number is required")
    @Pattern(regexp = "^[6-9]\\d{9}$", message = "Must be a valid 10-digit Indian mobile number")
    private String phoneNumber;

    @NotBlank(message = "Message cannot be blank")
    @Size(min = 1, max = 160, message = "Message must be between 1 and 160 characters")
    private String message;

    // Optional: client may provide their own idempotency key; server generates one if absent
    private String requestId;

    public SmsRequest() {}

    public SmsRequest(String phoneNumber, String message) {
        this.phoneNumber = phoneNumber;
        this.message = message;
    }
}
