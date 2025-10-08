package usecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"telegram-doctor-recipe-helper-bot/internal/app/config"
	"telegram-doctor-recipe-helper-bot/internal/app/exception"
	"telegram-doctor-recipe-helper-bot/internal/app/model"
	"telegram-doctor-recipe-helper-bot/internal/modules/bot/repository"
	"time"
)

type MessageUseCase interface {
	ProcessWebhookMessage(payload interface{}) error
	SendMessage(phoneNumber, message string) error
}

type messageUseCase struct {
	repo repository.MessageRepository
}

func NewMessageUseCase(repo repository.MessageRepository) MessageUseCase {
	return &messageUseCase{
		repo: repo,
	}
}

// WebhookMessage represents the structure from the webhook
type WebhookMessage struct {
	MessageID string `json:"message_id"`
	From      string `json:"from"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	PushName  string `json:"pushname,omitempty"`
	Type      string `json:"type,omitempty"`
}

// ProcessWebhookMessage handles incoming webhook messages
func (uc *messageUseCase) ProcessWebhookMessage(payload interface{}) error {
	// Type assertion to get the webhook message
	webhookData, ok := payload.(*WebhookMessage)
	if !ok {
		return &exception.BadRequestError{Message: "Invalid webhook payload type"}
	}

	// Extract phone number (remove @s.whatsapp.net or @c.us)
	phoneNumber := strings.Split(webhookData.From, "@")[0]

	// Check if from allowed number
	if phoneNumber != config.LoadConfig().AllowedNumber {
		// Optionally log or ignore messages from other numbers
		return nil
	}

	// Only process text messages
	if webhookData.Type != "" && webhookData.Type != "text" {
		return nil // Ignore non-text messages
	}

	// Parse message
	orderData, err := uc.parseOrderMessage(webhookData.Message, phoneNumber)
	if err != nil {
		// Send error message back to user
		errorMsg := "❌ Invalid format! Please use:\nname: [name], order: [order], phone number: [number]"
		uc.SendMessage(phoneNumber, errorMsg)
		return nil // Don't return error to avoid webhook retry
	}

	// Save to Excel
	if err := uc.repo.SaveToExcel(orderData); err != nil {
		return err
	}

	// Send confirmation
	confirmMsg := fmt.Sprintf("✅ Order received!\nName: %s\nOrder: %s\nPhone: %s",
		orderData.Name, orderData.Recipe, orderData.PhoneNumber)

	return uc.SendMessage(phoneNumber, confirmMsg)
}

// Parse message with pattern: name:..., order:..., phone number:...
func (uc *messageUseCase) parseOrderMessage(message, from string) (*model.OrderData, error) {
	// Convert to lowercase for easier parsing
	lowerMsg := strings.ToLower(message)

	// Regex patterns (more flexible)
	namePattern := regexp.MustCompile(`name\s*:\s*([^,\n]+)`)
	orderPattern := regexp.MustCompile(`order\s*:\s*([^,\n]+)`)
	phonePattern := regexp.MustCompile(`phone\s*(?:number)?\s*:\s*([^,\s\n]+)`)

	nameMatch := namePattern.FindStringSubmatch(lowerMsg)
	orderMatch := orderPattern.FindStringSubmatch(lowerMsg)
	phoneMatch := phonePattern.FindStringSubmatch(lowerMsg)

	if nameMatch == nil || orderMatch == nil || phoneMatch == nil {
		return nil, &exception.BadRequestError{Message: "Invalid message format"}
	}

	return &model.OrderData{
		Name:        strings.TrimSpace(nameMatch[1]),
		Recipe:      strings.TrimSpace(orderMatch[1]),
		PhoneNumber: strings.TrimSpace(phoneMatch[1]),
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// SendMessage sends a WhatsApp message via the API
func (uc *messageUseCase) SendMessage(phoneNumber, message string) error {
	cfg := config.LoadConfig()
	url := fmt.Sprintf("%s/send/message", cfg.WhatsAppAPIURL)

	// Prepare the request payload
	payload := map[string]interface{}{
		"phone":   phoneNumber,
		"message": message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return &exception.InternalServerError{Message: "Failed to marshal message"}
	}

	// Send POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return &exception.InternalServerError{Message: "Failed to send message"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &exception.InternalServerError{Message: "WhatsApp API returned error"}
	}

	return nil
}
