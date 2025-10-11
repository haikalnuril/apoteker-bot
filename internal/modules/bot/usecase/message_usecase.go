package usecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	ProcessWebhookMessage(payload *WebhookMessage) error
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
	ChatID    string         `json:"chat_id"`
	From      string         `json:"from"`
	Message   MessageContent `json:"message"`
	PushName  string         `json:"pushname"`
	SenderID  string         `json:"sender_id"`
	Timestamp string         `json:"timestamp"`
}

// MessageContent represents the message object structure
type MessageContent struct {
	Text          string `json:"text"`
	ID            string `json:"id"`
	RepliedID     string `json:"replied_id"`
	QuotedMessage string `json:"quoted_message"`
}

// ProcessWebhookMessage handles incoming webhook messages
func (uc *messageUseCase) ProcessWebhookMessage(webhookData *WebhookMessage) error {
	// Use sender_id directly (no need to parse)
	phoneNumber := webhookData.SenderID

	// The validation is now done in the controller, so we can skip it here
	// But we can add a safety check
	if phoneNumber != config.LoadConfig().AllowedNumber {
		return nil // This shouldn't happen as controller already checked
	}

	// Get the actual message text from the message object
	messageText := webhookData.Message.Text

	// Skip empty messages
	if strings.TrimSpace(messageText) == "" {
		return nil
	}
	
	return uc.SendMessage(phoneNumber, "Thank you for your message. We have received it and will process your order shortly.")
}

// Parse message with pattern: name:..., order:..., phone number:...
func (uc *messageUseCase) parseOrderMessage(message string) (*model.OrderData, error) {
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

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return &exception.InternalServerError{Message: "Failed to create request"}
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")

	// Add Basic Auth (username: admin, password: admin)
	req.SetBasicAuth("admin", "admin")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("HTTP request failed: %v\n", err)
		return &exception.InternalServerError{Message: "Failed to send message: " + err.Error()}
	}
	defer resp.Body.Close()

	// Read response body for debugging
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return &exception.InternalServerError{Message: fmt.Sprintf("WhatsApp API returned error: %d - %s", resp.StatusCode, string(body))}
	}

	return nil
}