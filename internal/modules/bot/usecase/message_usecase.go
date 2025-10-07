package usecase

import (
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
	ProcessIncomingMessages() error
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

type WhatsAppMessage struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type WhatsAppResponse struct {
	Success bool              `json:"success"`
	Data    []WhatsAppMessage `json:"data"`
}

// Get latest messages from WhatsApp API
func (uc *messageUseCase) ProcessIncomingMessages() error {
	url := fmt.Sprintf("%s/messages", config.LoadConfig().WhatsAppAPIURL)

	resp, err := http.Get(url)
	if err != nil {
		return &exception.InternalServerError{Message: "Internal Server Error"}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &exception.InternalServerError{Message: "Internal Server Error"}
	}

	var whatsappResp WhatsAppResponse
	if err := json.Unmarshal(body, &whatsappResp); err != nil {
		return &exception.InternalServerError{Message: "Internal Server Error"}
	}

	// Process each message
	for _, msg := range whatsappResp.Data {
		// Extract phone number (remove @s.whatsapp.net)
		phoneNumber := strings.Split(msg.From, "@")[0]

		// Check if from allowed number
		if phoneNumber != config.LoadConfig().AllowedNumber {
			continue
		}

		// Parse message
		orderData, err := uc.parseOrderMessage(msg.Message, phoneNumber)
		if err != nil {
			continue
		}

		// Save to Excel
		if err := uc.repo.SaveToExcel(orderData); err != nil {
			return err
		}

		// Send confirmation
		confirmMsg := fmt.Sprintf("✅ Order received!\nName: %s\nOrder: %s\nPhone: %s",
			orderData.Name, orderData.Recipe, orderData.PhoneNumber)
		uc.SendMessage(phoneNumber, confirmMsg)
	}

	return nil
}

// Parse message with pattern: name:..., order:..., phone number:...
func (uc *messageUseCase) parseOrderMessage(message, from string) (*model.OrderData, error) {
	// Convert to lowercase for easier parsing
	lowerMsg := strings.ToLower(message)

	// Regex patterns
	namePattern := regexp.MustCompile(`name\s*:\s*([^,]+)`)
	orderPattern := regexp.MustCompile(`order\s*:\s*([^,]+)`)
	phonePattern := regexp.MustCompile(`phone\s*number\s*:\s*([^,\s]+)`)

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

// Send message via WhatsApp API
func (uc *messageUseCase) SendMessage(phoneNumber, message string) error {
	url := fmt.Sprintf("%s/send/message", config.LoadConfig().WhatsAppAPIURL)

	payload := map[string]string{
		"phone":   phoneNumber,
		"message": message,
	}

	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return &exception.InternalServerError{Message: "Internal Server Error"}
	}
	req.Header.Set("Content-Type", "application/json")

	// Add basic auth
	req.SetBasicAuth(config.LoadConfig().GowaAdmin, config.LoadConfig().GowaPassword)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &exception.InternalServerError{Message: "Internal Server Error"}
	}
	defer resp.Body.Close()

	return nil
}
