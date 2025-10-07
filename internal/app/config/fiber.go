package config

import (
	"encoding/json"
	"os"
	"strconv"
	"telegram-doctor-recipe-helper-bot/internal/app/exception"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func NewFiber(cfg *Config) *fiber.App {
	config := fiber.Config{
		ErrorHandler: exception.Handler,
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
	}

	app := fiber.New(config)
	app.Use(cors.New())
	app.Use(recover.New())

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusNotFound).SendString("Endpoint / is not set or not found.")
	})

	app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe:    func(ctx *fiber.Ctx) bool { return true },
		LivenessEndpoint: "/healthz",
	}))

	if debug, _ := strconv.ParseBool(os.Getenv("APP_DEBUG")); debug {
		log := logger.Config{
			Format:     "[${time}] ${status} - ${method} ${path}\n", // e.g. [2006-01-02 15:04:05] 200 - GET /
			TimeFormat: "2006-01-02 15:04:05",
			TimeZone:   "Asia/Jakarta",
		}
		app.Use(logger.New(log))
	}

	return app
}
