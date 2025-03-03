package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	CheckIntervalInSec int             `json:"checkIntervalInSec"`
	CheckTimeoutInSec  int             `json:"checkTimeoutInSec"`
	SMTP               *EmailConfig    `json:"smtp,omitempty"`
	Telegram           *TelegramConfig `json:"telegram,omitempty"`
	TargetFile         string          `json:"targetFile,omitempty"`
	Port               int             `json:"port,omitempty"`
}

func LoadConfig(path string) (*Config, error) {
	// Default configuration
	config := &Config{
		CheckIntervalInSec: 30,
		CheckTimeoutInSec:  10,
		Port:               8080,
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil // Return default config if file doesn't exist
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal JSON
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return config, nil
}
