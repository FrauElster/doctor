package main

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramConfig struct {
	BotToken          string `json:"botToken"`
	ChatID            int64  `json:"chatId"`
	ThrottleInSeconds int    `json:"throttleInSecs,omitempty"` // Optional throttle duration in minutes
}

type telegramAlerter struct {
	bot      *tgbotapi.BotAPI
	chatID   int64
	cache    map[string]time.Time // Cache to store last alert time for each target
	throttle time.Duration
}

// NewTelegramAlerter creates a new AlertFunc that sends alerts to Telegram
func NewTelegramAlerter(config TelegramConfig) (AlertFunc, error) {
	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	alerter := &telegramAlerter{
		bot:      bot,
		chatID:   config.ChatID,
		cache:    make(map[string]time.Time),
		throttle: time.Duration(config.ThrottleInSeconds) * time.Second,
	}

	return alerter.alert, nil
}

func (t *telegramAlerter) alert(target HealthTarget, result Result) error {
	// Check if we should throttle this alert
	if lastAlert, ok := t.cache[target.ID]; ok {
		if time.Since(lastAlert) < t.throttle {
			return nil
		}
	}

	t.cache[target.ID] = time.Now()

	// Create alert message
	msg := fmt.Sprintf("⚠️ Alert for %s (%s)\n", target.ID, target.URLString)
	msg += fmt.Sprintf("Status: %d\n", result.Status)
	msg += fmt.Sprintf("Duration: %v\n", result.Duration)
	msg += fmt.Sprintf("Timestamp: %v\n", result.Timestamp.Format(time.RFC3339))

	if result.Error != nil {
		msg += fmt.Sprintf("Error: %v\n", result.Error)
	}

	// Send message to Telegram
	tgMsg := tgbotapi.NewMessage(t.chatID, msg)
	_, err := t.bot.Send(tgMsg)
	if err != nil {
		return fmt.Errorf("failed to send Telegram alert: %w", err)
	}

	return nil
}

type telegramResolver struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

// NewTelegramResolver creates a new AlertFunc that sends resolution notices to Telegram
func NewTelegramResolver(config TelegramConfig) (AlertFunc, error) {
	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	resolver := &telegramResolver{
		bot:    bot,
		chatID: config.ChatID,
	}

	return resolver.resolve, nil
}

func (t *telegramResolver) resolve(target HealthTarget, result Result) error {
	// Create resolution message
	msg := fmt.Sprintf("✅ Resolved for %s (%s)\n", target.ID, target.URLString)
	msg += fmt.Sprintf("Status: %d\n", result.Status)
	msg += fmt.Sprintf("Duration: %v\n", result.Duration)
	msg += fmt.Sprintf("Timestamp: %v\n", result.Timestamp.Format(time.RFC3339))

	// Send message to Telegram
	tgMsg := tgbotapi.NewMessage(t.chatID, msg)
	_, err := t.bot.Send(tgMsg)
	if err != nil {
		return fmt.Errorf("failed to send Telegram resolution notice: %w", err)
	}

	return nil
}
