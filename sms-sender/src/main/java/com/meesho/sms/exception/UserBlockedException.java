package com.meesho.sms.exception;

public class UserBlockedException extends RuntimeException {
    public UserBlockedException(String phoneNumber) {
        super("User " + phoneNumber + " is blocked from sending SMS.");
    }
}
