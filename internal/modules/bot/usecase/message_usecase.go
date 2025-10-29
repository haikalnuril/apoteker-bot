package usecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	// "regexp"
	"strings"
	"telegram-doctor-recipe-helper-bot/internal/app/config"
	"telegram-doctor-recipe-helper-bot/internal/app/exception"

	// "telegram-doctor-recipe-helper-bot/internal/app/model"
	"telegram-doctor-recipe-helper-bot/internal/app/utils"
	"time"
)

type MessageUseCase interface {
	ProcessWebhookMessage(payload *WebhookMessage) error
	SendMessage(phoneNumber, message string) error
}

type messageUseCase struct {
	sheetService *utils.SheetService
}

// --- Define the conversation states as constants for safety ---
const (
	StateAwaitingStart          = "AWAITING_START"
	StateAwaitingMenuChoice     = "AWAITING_MENU_CHOICE"
	StateAwaitingFormSubmission = "AWAITING_FORM_SUBMISSION"
	StateAwaitingConfirmation   = "AWAITING_CONFIRMATION"
)

func NewMessageUseCase(sheetService *utils.SheetService) MessageUseCase {
	return &messageUseCase{
		sheetService: sheetService,
	}
}

// WebhookMessage represents the structure from the webhook
type WebhookMessage struct {
	ChatID    string         `json:"chat_id"`
	From      string         `json:"from"`
	Message   MessageContent `json:"message"`
	PushName  string         `json:"pushname"`
	SenderID  string         `json:"sender_id"`
	Timestamp string         `json:"timestamp"`
}

// MessageContent represents the message object structure
type MessageContent struct {
	Text          string `json:"text"`
	ID            string `json:"id"`
	RepliedID     string `json:"replied_id"`
	QuotedMessage string `json:"quoted_message"`
}

var Queue = 1
var DateNow = time.Now().Format("2006-01-02")
var queueMutex = &sync.Mutex{}

// ProcessWebhookMessage handles incoming webhook messages using a state machine
func (uc *messageUseCase) ProcessWebhookMessage(webhookData *WebhookMessage) error {
	phoneNumber := webhookData.SenderID
	messageText := webhookData.Message.Text

	// Skip empty or unauthorized messages (assuming initial check is still elsewhere)
	if strings.TrimSpace(messageText) == "" {
		return nil
	}
	if phoneNumber != config.LoadConfig().AllowedNumber && phoneNumber != config.LoadConfig().NewDoctor {
		return nil
	}

	// 1. Get the user's current state
	currentUserState := utils.GetOrCreateUserState(phoneNumber)

	// 2. Validate the message against the current state
	isValid, data, errorMessage := utils.ValidateMessageForState(currentUserState.State, messageText)

	// 3. Act on the validation result
	if !isValid {
		// If not valid, just send the specific error message and do nothing else
		return uc.SendMessage(phoneNumber, errorMessage)
	}

	// --- If VALID, process the request based on the current state ---
	switch currentUserState.State {
	case StateAwaitingStart:
		// User sent /start
		welcomeMessage := "halo, ini adalah bot penghubung antara dokter dan apoteker.\n[1] Buat Resep\n[2] Membuka Link Spreadsheet\n[3] Cancel\n\nJawab dengan angka saja!"
		uc.SendMessage(phoneNumber, welcomeMessage)
		currentUserState.State = StateAwaitingMenuChoice // <-- State Transition

	case StateAwaitingMenuChoice:
		choice := data.(string)
		if choice == "1" {
			formFormat := "Mohon kirim data pasien dengan detail format berikut:\nNama Dokter: \nNama Pasien: \nTanggal Lahir Pasien: \nNo Regis: \nResep Obat: \nNomor Telpon Pasien: \nPembiayaan: "
			uc.SendMessage(phoneNumber, formFormat)
			currentUserState.State = StateAwaitingFormSubmission // <-- State Transition
		} else if choice == "2" {
			message := fmt.Sprintf("Berikut adalah link spreadsheet: %s", config.LoadConfig().SheetLink)
			uc.SendMessage(phoneNumber, message)
			uc.SendMessage(phoneNumber, "Sesi selesai.")
			// uc.CloseChat(phoneNumber) // Assuming you have a close function
			utils.ResetUserState(phoneNumber) // <-- Reset State
		} else if choice == "3" {
			uc.SendMessage(phoneNumber, "Sesi dibatalkan. Untuk memulai kembali, kirim `/start`.")
			// uc.CloseChat(phoneNumber) // Assuming you have a close function
			utils.ResetUserState(phoneNumber) // <-- Reset State
		}

	case StateAwaitingFormSubmission:
		if cmd, ok := data.(string); ok && cmd == "cancel" {
			welcomeMessage := "[1] Buat Resep\n[2] Membuka Link Spreadsheet\n[3] Cancel\n\nJawab dengan angka saja!"
			uc.SendMessage(phoneNumber, "Permintaan dibatalkan. Kembali ke halaman utama.\n\n"+welcomeMessage)
			currentUserState.State = StateAwaitingMenuChoice // <-- State Transition
			return nil
		}
		// Store the original message text for confirmation
		currentUserState.PendingMessage = messageText
		confirmationPrompt := fmt.Sprintf("Mohon konfirmasi permintaan anda:\n\n%s\n\nApakah sudah benar? (Y/N)", messageText)
		uc.SendMessage(phoneNumber, confirmationPrompt)
		currentUserState.State = StateAwaitingConfirmation // <-- State Transition

	case StateAwaitingConfirmation:
		decision := data.(string)
		if decision == "Y" {
			patientDetails, err := utils.ParsePatientDetails(currentUserState.PendingMessage)
			if err != nil {
				uc.SendMessage(phoneNumber, "Error: Data yang dikirim terdapat kesalahan format. Mohon untuk mencoba kembali.")

				formFormat := "Mohon kirim data pasien dengan detail format berikut:\nNama Dokter: \nNama Pasien: \nTanggal Lahir Pasien: \nNo Regis: \nResep Obat: \nNomor Telpon Pasien: \nPembiayaan: "
				uc.SendMessage(phoneNumber, formFormat)
				currentUserState.State = StateAwaitingFormSubmission
				return err
			}

			// --- START NEW QUEUE LOGIC ---

			// 1. Lock the mutex to prevent other requests from
			//    reading/writing the queue at the same time.
			queueMutex.Lock()

			// 2. Check if it's a new day
			today := time.Now().Format("2006-01-02")
			if DateNow != today {
				Queue = 1       // Reset queue to 1
				DateNow = today // Update the date to today
			}

			// 3. Get the queue number for *this* request
			currentQueueNumber := Queue

			// 4. Increment the global Queue for the *next* request
			Queue += 1

			// 5. Unlock the mutex so other requests can continue
			queueMutex.Unlock()

			// --- END NEW QUEUE LOGIC ---

			err = uc.sheetService.AddPrescriptionRow(patientDetails, currentQueueNumber)
			if err != nil {
				// If it fails, tell the doctor but maybe still send to pharmacy
				uc.SendMessage(phoneNumber, "Note: Gagal untuk menyimpan ke spreadsheet, tetapi tetap akan dikirim ke apoteker.")
				// You can decide if you want to stop here or continue
			}

			originalMeds := patientDetails.Medication

			medParts := strings.Split(originalMeds, ",")

			numberedMedParts := make([]string, len(medParts))

			for i, part := range medParts {
				trimmedPart := strings.TrimSpace(part)
				// Format as "1. Item", "2. Item", etc.
				numberedMedParts[i] = fmt.Sprintf("%d. %s", i+1, trimmedPart)
			}

			formattedMeds := strings.Join(numberedMedParts, ",\n")

			// **SEND TO PHARMACY LOGIC HERE**
			pharmacyNumber := config.LoadConfig().PharmacyNumber
			// Send the pending message to the pharmacy number
			msgToPharmacy := fmt.Sprintf("Permintaan resep obat baru:\n\n%s \n\nDengan nomor Antrian: %d\n\nObat ini untuk:\n%s\n%s\n%s\n\nDari:\nDokter %s", formattedMeds, currentQueueNumber, patientDetails.PatientName, patientDetails.PatientBirthDate, patientDetails.PatientPhoneNumber, patientDetails.DoctorName)
			err = uc.SendMessage(pharmacyNumber, msgToPharmacy)
			if err != nil {
				uc.SendMessage(phoneNumber, "Gagal mengirim pesan ke apoteker. Mohon coba kembali lagi nanti.")
				return err
			}
			msgToPatient := fmt.Sprintf("Halo %s, permintaan resepmu:\n\n%s \n\nsudah dikirim ke apoteker. Antrian kamu adalah %d. Mohon ditunggu.", patientDetails.PatientName, formattedMeds, currentQueueNumber)

			if patientDetails.PatientPhoneNumber != "-" {
				uc.SendMessage(patientDetails.PatientPhoneNumber, msgToPatient)
			}
			uc.SendMessage(phoneNumber, "Permintaan kamu sudah dikirimkan kebagian apoteker. Sesi Selesai.")
			utils.ResetUserState(phoneNumber) // <-- Reset State
		} else if decision == "N" {
			formFormat := "Permintaan dibatalkan. Mohon kirim ulang dengan detail form yang benar:\nNama Dokter: \nNama Pasien: \nTanggal Lahir Pasien: \nNo Regis: \nResep Obat: \nNomor Telpon Pasien: \nPembiayaan: "
			uc.SendMessage(phoneNumber, formFormat)
			currentUserState.State = StateAwaitingFormSubmission // <-- State Transition
		}
	}

	return nil
}

// SendMessage sends a WhatsApp message via the API
func (uc *messageUseCase) SendMessage(phoneNumber, message string) error {
	cfg := config.LoadConfig()
	url := fmt.Sprintf("%s/send/message", cfg.WhatsAppAPIURL)
	gowaUsername := cfg.GowaAdmin
	gowaPassword := cfg.GowaPassword

	// Prepare the request payload
	payload := map[string]interface{}{
		"phone":   phoneNumber,
		"message": message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return &exception.InternalServerError{Message: "Failed to marshal message"}
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return &exception.InternalServerError{Message: "Failed to create request"}
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")

	// Add Basic Auth (username: admin, password: admin)
	req.SetBasicAuth(gowaUsername, gowaPassword)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("HTTP request failed: %v\n", err)
		return &exception.InternalServerError{Message: "Failed to send message: " + err.Error()}
	}
	defer resp.Body.Close()

	// Read response body for debugging
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return &exception.InternalServerError{Message: fmt.Sprintf("WhatsApp API returned error: %d - %s", resp.StatusCode, string(body))}
	}

	return nil
}
