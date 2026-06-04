package com.meesho.sms;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;

import java.util.function.BooleanSupplier;

@SpringBootApplication
public class SmsSenderApplication {
    public static void main(String[] args) {
        SpringApplication.run(SmsSenderApplication.class, args);
    }

    @Bean
    public BooleanSupplier vendorSimulator() {
        return () -> Math.random() > 0.2;
    }
}