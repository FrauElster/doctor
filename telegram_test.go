package main

import (
	"net/url"
	"testing"
	"time"
)

func TestTelegram(t *testing.T) {
	t.Run("Test sending a Telegram alert", func(t *testing.T) {
		config, err := LoadConfig("config.json")
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		if config.Telegram == nil {
			t.Skip("Telegram config is not provided, skipping test")
		}

		alerter, err := NewTelegramAlerter(*config.Telegram)
		if err != nil {
			t.Fatalf("Failed to create Telegram alerter: %v", err)
		}

		// Test target and result
		urlString := "https://google.com"
		url, _ := url.Parse(urlString)
		target := HealthTarget{
			URL:       url,
			URLString: urlString,
			ID:        "google",
		}
		result := Result{
			Target:    target,
			Status:    404,
			Healthy:   false,
			Timestamp: time.Now(),
			Duration:  420 * time.Millisecond,
		}

		// Test alert
		err = alerter(target, result)
		if err != nil {
			t.Fatalf("Failed to send Telegram alert: %v", err)
		}
	})

	t.Run("Test sending a Telegram resolution", func(t *testing.T) {
		config, err := LoadConfig("config.json")
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		if config.Telegram == nil {
			t.Skip("Telegram config is not provided, skipping test")
		}

		resolver, err := NewTelegramResolver(*config.Telegram)
		if err != nil {
			t.Fatalf("Failed to create Telegram resolver: %v", err)
		}

		// Test target and result
		urlString := "https://google.com"
		url, _ := url.Parse(urlString)
		target := HealthTarget{
			URL:       url,
			URLString: urlString,
			ID:        "google",
		}
		result := Result{
			Target:    target,
			Status:    200,
			Healthy:   true,
			Timestamp: time.Now(),
			Duration:  420 * time.Millisecond,
		}

		// Test resolution
		err = resolver(target, result)
		if err != nil {
			t.Fatalf("Failed to send Telegram resolution: %v", err)
		}
	})

	t.Run("Test throttling", func(t *testing.T) {
		config, err := LoadConfig("config.json")
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		if config.Telegram == nil {
			t.Skip("Telegram config is not provided, skipping test")
		}

		// Create telegram alerter with short throttle duration for testing
		telegramConfig := TelegramConfig{
			BotToken:          config.Telegram.BotToken,
			ChatID:            config.Telegram.ChatID,
			ThrottleInSeconds: 1,
		}

		alerter, err := NewTelegramAlerter(telegramConfig)
		if err != nil {
			t.Fatalf("Failed to create Telegram alerter: %v", err)
		}

		// Test target
		urlString := "https://google.com"
		url, _ := url.Parse(urlString)
		target := HealthTarget{
			URL:       url,
			URLString: urlString,
			ID:        "google",
		}
		result := Result{
			Target:    target,
			Status:    404,
			Healthy:   false,
			Timestamp: time.Now(),
			Duration:  420 * time.Millisecond,
		}

		// Send first alert
		err = alerter(target, result)
		if err != nil {
			t.Fatalf("Failed to send first Telegram alert: %v", err)
		}

		// Send second alert immediately - should be throttled
		err = alerter(target, result)
		if err != nil {
			t.Fatalf("Throttled alert should not return error: %v", err)
		}

		// Wait for throttle duration
		time.Sleep(2 * time.Second)

		// Send third alert - should work
		err = alerter(target, result)
		if err != nil {
			t.Fatalf("Failed to send third Telegram alert after throttle period: %v", err)
		}
	})
}
