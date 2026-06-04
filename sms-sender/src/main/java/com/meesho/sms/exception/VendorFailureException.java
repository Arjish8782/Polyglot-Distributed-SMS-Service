package com.meesho.sms.exception;

public class VendorFailureException extends RuntimeException {
    public VendorFailureException() {
        super("SMS failed at vendor.");
    }
}
