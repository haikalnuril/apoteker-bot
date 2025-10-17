package main

import (
	"fmt"
	"log"
	"telegram-doctor-recipe-helper-bot/internal/app/config"
	"telegram-doctor-recipe-helper-bot/internal/app/utils"
	"telegram-doctor-recipe-helper-bot/internal/modules/bot/controller"
	"telegram-doctor-recipe-helper-bot/internal/modules/bot/router"
	"telegram-doctor-recipe-helper-bot/internal/modules/bot/usecase"
)

func main() {
	cfg := config.LoadConfig()

	app := config.NewFiber(cfg)


	sheetService, err := utils.NewSheetService("bot-credentials.json", cfg.SheetID)
    if err != nil {
        log.Fatalf("Failed to create sheet service: %v", err)
    }
	messageUseCase := usecase.NewMessageUseCase(sheetService)
	botController := controller.NewBotController(messageUseCase)

	router.Route(app, botController)

	// Start server
	port := cfg.AppPort
	log.Printf("🚀 Bot server starting on port %s", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}
