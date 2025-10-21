package utils

import (
	"regexp"
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
	lines := strings.Split(message, "\n")

	parsedFields := make(map[string]string)

	for _, line := range lines {
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

	// Normalize phone number before assigning
	rawPhone := parsedFields["Patient Phone Number"]
	details.PatientPhoneNumber = normalizePhone(rawPhone)

	if details.DoctorName == "" ||
		details.PatientName == "" ||
		details.PatientBirthDate == "" ||
		details.Medication == "" ||
		details.PatientPhoneNumber == "" {
		return nil, &exception.BadRequestError{Message: "Missing required fields in the message"}
	}

	return details, nil
}

// normalizePhone cleans and converts phone number to Indonesian standard format: 628xxxxxxxxxx
func normalizePhone(input string) string {
	if input == "" {
		return ""
	}

	// Remove all non-digit characters
	re := regexp.MustCompile(`\D`)
	phone := re.ReplaceAllString(input, "")

	// Normalize prefix
	switch {
	case strings.HasPrefix(phone, "0"):
		phone = "62" + phone[1:]
	case strings.HasPrefix(phone, "62"):
		// already correct
	case strings.HasPrefix(phone, "8"):
		phone = "62" + phone
	}

	return phone
}