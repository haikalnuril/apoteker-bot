package utils

import (
	"regexp"
	"strings"
)

// ExtractedFormData holds the parsed data from the user's prescription message.
type ExtractedFormData struct {
	Doctor      string
	Patient     string
	BirthDate   string
	Medication  string
	PhoneNumber string
}

// The CORRECT pre-compiled regex for validating the multi-line format.
// We use the (?s) flag to allow '.' to match newlines.
var formRegex = regexp.MustCompile(`(?is)Doctor Name:\s*(.*?)\s*Patient Name:\s*(.*?)\s*Patient Birth Date:\s*(.*?)\s*Medication:\s*(.*?)\s*Patient Phone Number:\s*(.*)`)

// validateMessageForState checks if a message is valid for the given state.
// It returns: (isValid bool, extractedData interface{}, errorMessage string)
func ValidateMessageForState(state string, message string) (bool, interface{}, string) {
	switch state {
	case StateAwaitingStart:
		if strings.ToLower(message) == "/start" {
			return true, nil, ""
		}
		return false, nil, "To start a new session, please send `/start`."

	case StateAwaitingMenuChoice:
		if message == "1" || message == "2" || message == "3" {
			return true, message, ""
		}
		return false, nil, "Invalid input. Please reply with `1`, `2`, or `3`."

	case StateAwaitingFormSubmission:
        if strings.ToLower(message) == "cancel" {
            return true, "cancel", ""
        }
        // This part will now work correctly with the new regex
        matches := formRegex.FindStringSubmatch(message)
        if len(matches) == 6 { 
            data := ExtractedFormData{
                Doctor:      strings.TrimSpace(matches[1]),
                Patient:     strings.TrimSpace(matches[2]),
                BirthDate:   strings.TrimSpace(matches[3]),
                Medication:  strings.TrimSpace(matches[4]),
                PhoneNumber: strings.TrimSpace(matches[5]),
            }
            return true, data, ""
        }
        return false, nil, "The format appears incorrect. Please follow the required format or send `cancel` to go back to the main menu."

	case StateAwaitingConfirmation:
		cleanMsg := strings.ToUpper(message)
		if cleanMsg == "Y" || cleanMsg == "YES" {
			return true, "Y", ""
		}
		if cleanMsg == "N" || cleanMsg == "NO" {
			return true, "N", ""
		}
		return false, nil, "Invalid response. Please reply with 'Y' to confirm or 'N' to edit."
	}

	// Fallback for any unknown state
	return false, nil, "An unexpected error occurred. Please send /start to begin again."
}
