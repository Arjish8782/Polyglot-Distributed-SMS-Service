package com.meesho.sms.controller;

import com.meesho.sms.dto.SmsRequest;
import com.meesho.sms.dto.SmsResponse;
import com.meesho.sms.service.SmsService;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/v1/sms")
public class SmsController {

    // This pulls in the service we created earlier
    private final SmsService smsService;

    public SmsController(SmsService smsService) {
        this.smsService = smsService;
    }

    // This maps the POST request to our business logic
    @PostMapping("/send")
    public SmsResponse sendSms(@RequestBody SmsRequest request) {
        return smsService.processSms(request);
    }
}