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
	RegistryNum        string
	Medication         string
	PatientPhoneNumber string
	PaymentMethod      string
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

	details.DoctorName = parsedFields["Nama Dokter"]
	details.PatientName = parsedFields["Nama Pasien"]
	details.PatientBirthDate = parsedFields["Tanggal Lahir Pasien"]

	rawRegistryNum := parsedFields["No Regis"]
	details.RegistryNum = normalizeRegistryNum(rawRegistryNum)

	details.Medication = parsedFields["Resep Obat"]

	// Normalize phone number before assigning
	rawPhone := parsedFields["Nomor Telpon Pasien"]
	details.PatientPhoneNumber = normalizePhone(rawPhone)

	details.PaymentMethod = parsedFields["Pembiayaan"]

	if details.DoctorName == "" ||
		details.PatientName == "" ||
		details.PatientBirthDate == "" ||
		details.RegistryNum == "" ||
		details.Medication == "" ||
		details.PatientPhoneNumber == "" ||
		details.PaymentMethod == "" {
		return nil, &exception.BadRequestError{Message: "Missing required fields in the message"}
	}

	return details, nil
}

// normalizePhone cleans and converts phone number to Indonesian standard format: 628xxxxxxxxxx
func normalizePhone(input string) string {
	if input == "" || input == "-" {
		return "-"
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

func normalizeRegistryNum(input string) string {
	if input == "" {
		return ""
	}

	// Check if the string starts with "0"
	if strings.HasPrefix(input, "0") {
		// If it does, add a single quote ' in front
		return "'" + input
	}

	// Otherwise, return the original string
	return input
}
