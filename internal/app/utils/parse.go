package utils

import (
	"strings"
	"telegram-doctor-recipe-helper-bot/internal/app/exception"
)

type PatientDetails struct {
	DoctorName         string
	PatientName        string
	PatientBirthDate   string
	Medication         string
	PatientPhoneNumber string
}

func ParsePatientDetails(message string) (*PatientDetails, error) {
	details := &PatientDetails{}
	line := strings.Split(message, "\n")

	parsedFields := make(map[string]string)

	for _, line := range line {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		parsedFields[key] = value
	}

	details.DoctorName = parsedFields["Doctor Name"]
	details.PatientName = parsedFields["Patient Name"]
	details.PatientBirthDate = parsedFields["Patient Birth Date"]
	details.Medication = parsedFields["Medication"]
	details.PatientPhoneNumber = parsedFields["Patient Phone Number"]

	if details.DoctorName == "" || details.PatientName == "" || details.PatientBirthDate == "" || details.Medication == "" || details.PatientPhoneNumber == "" {
		return nil, &exception.BadRequestError{Message: "Missing required fields in the message"}
	}
	return details, nil
}
