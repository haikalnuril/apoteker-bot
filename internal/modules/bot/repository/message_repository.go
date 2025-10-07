package repository

import (
	"fmt"
	"os"
	"telegram-doctor-recipe-helper-bot/internal/app/model"
	"telegram-doctor-recipe-helper-bot/internal/app/config"

	"github.com/xuri/excelize/v2"
)

type MessageRepository interface {
	SaveToExcel(order *model.OrderData) error
}

type messageRepository struct{}

func NewMessageRepository() MessageRepository {
	return &messageRepository{}
}

func (r *messageRepository) SaveToExcel(order *model.OrderData) error {
	filePath := config.LoadConfig().ExcelOutputPath

	var f *excelize.File
	var err error

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create new file
		f = excelize.NewFile()

		// Create headers
		headers := []string{"Timestamp", "Name", "Order", "Phone Number"}
		for i, header := range headers {
			cell := fmt.Sprintf("%s1", string(rune('A'+i)))
			f.SetCellValue("Sheet1", cell, header)
		}
	} else {
		// Open existing file
		f, err = excelize.OpenFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to open excel file: %w", err)
		}
	}
	defer f.Close()

	// Get last row
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return fmt.Errorf("failed to get rows: %w", err)
	}

	nextRow := len(rows) + 1

	// Add new data
	f.SetCellValue("Sheet1", fmt.Sprintf("A%d", nextRow), order.Timestamp)
	f.SetCellValue("Sheet1", fmt.Sprintf("B%d", nextRow), order.Name)
	f.SetCellValue("Sheet1", fmt.Sprintf("C%d", nextRow), order.Recipe)
	f.SetCellValue("Sheet1", fmt.Sprintf("D%d", nextRow), order.PhoneNumber)

	// Save file
	if err := f.SaveAs(filePath); err != nil {
		return fmt.Errorf("failed to save excel file: %w", err)
	}

	return nil
}
