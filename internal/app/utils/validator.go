package utils

import (
	"regexp"
	"strings"
)

// ExtractedFormData holds the parsed data from the user's prescription message.
type ExtractedFormData struct {
	Patient string
	Recipe  string
	Dosage  string
}

// The pre-compiled regex for validating the form format. Compiling once is more efficient.
var formRegex = regexp.MustCompile(`(?i)Patient Name:\s*(.*),\s*Recipe:\s*(.*),\s*Dosage:\s*(.*)`)

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
		// Use regex to validate and extract data
		matches := formRegex.FindStringSubmatch(message)
		if len(matches) == 4 { // 4 because matches[0] is the full string
			data := ExtractedFormData{
				Patient: strings.TrimSpace(matches[1]),
				Recipe:  strings.TrimSpace(matches[2]),
				Dosage:  strings.TrimSpace(matches[3]),
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
