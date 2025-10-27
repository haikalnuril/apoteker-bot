package utils

import (
	"regexp"
	"strings"
)

// ExtractedFormData holds the parsed data from the user's prescription message.
type ExtractedFormData struct {
	Doctor        string
	Patient       string
	BirthDate     string
	RegistryNum   string
	Medication    string
	PhoneNumber   string
	PaymentMethod string
}

// The CORRECT pre-compiled regex for validating the multi-line format.
// We use the (?s) flag to allow '.' to match newlines.
var formRegex = regexp.MustCompile(`(?is)Nama Dokter:\s*(.*?)\s*Nama Pasien:\s*(.*?)\s*Tanggal Lahir Pasien:\s*(.*?)\s*No Regis:\s*(.*?)\s*Resep Obat:\s*(.*?)\s*Nomor Telpon Pasien:\s*(.*)\s*Pembiayaan:\s*(.*?)`)

// validateMessageForState checks if a message is valid for the given state.
// It returns: (isValid bool, extractedData interface{}, errorMessage string)
func ValidateMessageForState(state string, message string) (bool, interface{}, string) {
	switch state {
	case StateAwaitingStart:
		if strings.ToLower(message) == "/start" {
			return true, nil, ""
		}
		return false, nil, "Untuk memulai sesi baru, kirim pesan `/start`."

	case StateAwaitingMenuChoice:
		if message == "1" || message == "2" || message == "3" {
			return true, message, ""
		}
		return false, nil, "Inputan salah. Mohon reply dengan `1`, `2`, atau `3`."

	case StateAwaitingFormSubmission:
		if strings.ToLower(message) == "cancel" {
			return true, "cancel", ""
		}
		// This part will now work correctly with the new regex
		matches := formRegex.FindStringSubmatch(message)
		if len(matches) == 8 {
			data := ExtractedFormData{
				Doctor:        strings.TrimSpace(matches[1]),
				Patient:       strings.TrimSpace(matches[2]),
				BirthDate:     strings.TrimSpace(matches[3]),
				RegistryNum:   strings.TrimSpace(matches[4]),
				Medication:    strings.TrimSpace(matches[5]),
				PhoneNumber:   strings.TrimSpace(matches[6]),
				PaymentMethod: strings.TrimSpace(matches[7]),
			}
			return true, data, ""
		}
		return false, nil, "Format yang dikirimkan salah. Mohon ikuti syarat format atau kirim pesan `cancel` untuk kembali ke main menu."

	case StateAwaitingConfirmation:
		cleanMsg := strings.ToUpper(message)
		if cleanMsg == "Y" || cleanMsg == "YES" {
			return true, "Y", ""
		}
		if cleanMsg == "N" || cleanMsg == "NO" {
			return true, "N", ""
		}
		return false, nil, "Respon tidak sesuai. Mohon reply dengan 'Y' untuk komfirmasi atau 'N' untuk edit."
	}

	// Fallback for any unknown state
	return false, nil, "Terjadi error yang tidak diinginkan. Mohon kirim pesan `/start` untuk memulai kembali."
}
