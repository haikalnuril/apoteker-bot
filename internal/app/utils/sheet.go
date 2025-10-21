package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SheetService struct {
	client        *sheets.Service
	spreadsheetID string
}

func NewSheetService(credentialsFile string, spreadsheetID string) (*SheetService, error) {
	ctx := context.Background()

	// This creates the authenticated client using your JSON file
	client, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	return &SheetService{
		client:        client,
		spreadsheetID: spreadsheetID,
	}, nil
}

func (s *SheetService) AddPrescriptionRow(details *PatientDetails, Queue int) error {
	writeRange := "Prescriptions"

	var row []interface{}

	wibTime := time.Now().UTC().Add(7 * time.Hour)

	// Format: "02 January 2006 15:04 WIB"
	formattedTime := wibTime.Format("02 January 2006 15:04") + " WIB"

	row = append(row,
		Queue,
		details.DoctorName,
		details.PatientName,
		details.PatientBirthDate,
		details.Medication,
		details.PatientPhoneNumber,
		formattedTime,
	)

	// 3. Create the data structure the API needs
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{row},
	}

	// 4. Make the API call to append the data
	_, err := s.client.Spreadsheets.Values.Append(s.spreadsheetID, writeRange, valueRange).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("Unable to write data to sheet: %v", err)
		return err
	}

	log.Println("Successfully added a row to the spreadsheet.")
	return nil
}
