package main

import (
	"fmt"
	"log"
	"telegram-doctor-recipe-helper-bot/internal/app/config"
	"telegram-doctor-recipe-helper-bot/internal/modules/bot/controller"
	"telegram-doctor-recipe-helper-bot/internal/modules/bot/repository"
	"telegram-doctor-recipe-helper-bot/internal/modules/bot/router"
	"telegram-doctor-recipe-helper-bot/internal/modules/bot/usecase"
	"time"
)

func main() {
	cfg := config.LoadConfig()

	app := config.NewFiber(cfg)

	messageRepo := repository.NewMessageRepository()
	messageUseCase := usecase.NewMessageUseCase(messageRepo)
	botController := controller.NewBotController(messageUseCase)

	router.Route(app, botController)

	// Start background message polling (every 10 seconds)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := messageUseCase.ProcessIncomingMessages(); err != nil {
				log.Printf("Error processing messages: %v", err)
			}
		}
	}()

	// Start server
	port := cfg.AppPort
	log.Printf("🚀 Bot server starting on port %s", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}
