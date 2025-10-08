package controller

import (
	"fmt"
	"telegram-doctor-recipe-helper-bot/internal/app/model"
	"telegram-doctor-recipe-helper-bot/internal/modules/bot/usecase"

	"github.com/gofiber/fiber/v2"
)

type BotController struct {
	useCase usecase.MessageUseCase
}

func NewBotController(useCase usecase.MessageUseCase) *BotController {
	return &BotController{
		useCase: useCase,
	}
}

type WebhookPayload struct {
	MessageID string `json:"message_id"`
	From      string `json:"from"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	// Add other fields based on the actual webhook structure
	PushName string `json:"pushname,omitempty"`
	Type     string `json:"type,omitempty"`
}

// Webhook to receive incoming messages
func (ctrl *BotController) HandleWebhook(c *fiber.Ctx) error {
	var payload WebhookPayload

	// Log raw body for debugging
	fmt.Printf("Received webhook: %s\n", string(c.Body()))

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Code:    400,
			Message: "Invalid webhook payload: " + err.Error(),
		})
	}

	// Log parsed payload
	fmt.Printf("Parsed payload: %+v\n", payload)

	// Process the incoming message through use case
	if err := ctrl.useCase.ProcessWebhookMessage(&payload); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	// Return success response to WhatsApp server
	return c.JSON(model.Response{
		Code:    200,
		Message: "Webhook received successfully",
	})
}

// Manual send message
func (ctrl *BotController) SendMessage(c *fiber.Ctx) error {
	var req struct {
		PhoneNumber string `json:"phone_number"`
		Message     string `json:"message"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.Response{
			Code:    400,
			Message: "Invalid request body",
		})
	}

	if err := ctrl.useCase.SendMessage(req.PhoneNumber, req.Message); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Code:    500,
			Message: err.Error(),
		})
	}

	return c.JSON(model.Response{
		Code:    200,
		Message: "Message sent successfully",
	})
}

// Health check
func (ctrl *BotController) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(model.Response{
		Code:    200,
		Message: "Bot is running",
	})
}
