package main

import (
	"net/url"
	"testing"
	"time"
)

func TestMail(t *testing.T) {
	t.Run("Test sending a mail", func(t *testing.T) {
		config, err := LoadConfig("config.json")
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		if config.SMTP == nil {
			t.Skip("SMTP config is not provided, skipping test")
		}

		sendErrorMail := NewEmailAlert(*config.SMTP)

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

		err = sendErrorMail(target, result)
		if err != nil {
			t.Fatalf("Failed to send error mail: %v", err)
		}
	})
}
