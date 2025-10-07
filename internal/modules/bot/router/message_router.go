package router

import (
	"telegram-doctor-recipe-helper-bot/internal/modules/bot/controller"

	"github.com/gofiber/fiber/v2"
)

func Route(app *fiber.App, ctrl *controller.BotController) {
	message := app.Group("/v1/messages")

	message.Post("/webhook", ctrl.HandleWebhook)
	message.Post("/send", ctrl.SendMessage)
	message.Get("/health", ctrl.HealthCheck)
}
