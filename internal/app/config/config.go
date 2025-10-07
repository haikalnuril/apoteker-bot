package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort         string
	WhatsAppAPIURL  string
	AllowedNumber   string
	ExcelOutputPath string
	GowaAdmin       string
	GowaPassword    string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	c := &Config{}
	return &Config{
		AppPort:         c.Get("APP_PORT", "8080"),
		WhatsAppAPIURL:  c.Get("WHATSAPP_API_URL", "http://localhost:3000"),
		AllowedNumber:   c.Get("ALLOWED_NUMBER", "089123456789"),
		ExcelOutputPath: c.Get("EXCEL_OUTPUT_PATH", "./storage/orders.xlsx"),
		GowaAdmin:       c.Get("GOWA_USERNAME", "admin"),
		GowaPassword:    c.Get("GOWA_PASSWORD", "password"),
	}
}

func (c *Config) Get(key, def string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	log.Printf("Failed to get %s, using default value: %s", key, def)
	return def
}

func (c *Config) GetInt(key string, def int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		log.Printf("Failed to parse %s to int, using default value: %d", key, def)
		return def
	}

	return value
}
