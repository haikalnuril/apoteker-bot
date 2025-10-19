package usecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
		welcomeMessage := "hello, this is doctor to pharmacy bot.\n[1] Send message to pharmacy\n[2] Get Spreadsheet link\n[3] Cancel\n\nAnswer with number only!"
		uc.SendMessage(phoneNumber, welcomeMessage)
		currentUserState.State = StateAwaitingMenuChoice // <-- State Transition

	case StateAwaitingMenuChoice:
		choice := data.(string)
		if choice == "1" {
			formFormat := "Please send patient details in the format:\nDoctor Name: [Name]\nPatient Name: [Name]\nPatient Birth Date: [Date]\nMedication: [Drug]\nPatient Phone Number: [Number using 62]"
			uc.SendMessage(phoneNumber, formFormat)
			currentUserState.State = StateAwaitingFormSubmission // <-- State Transition
		} else if choice == "2" {
			message := fmt.Sprintf("Here is the spreadsheet link: %s", config.LoadConfig().SheetLink)
			uc.SendMessage(phoneNumber, message)
			uc.SendMessage(phoneNumber, "Session complete.")
			// uc.CloseChat(phoneNumber) // Assuming you have a close function
			utils.ResetUserState(phoneNumber) // <-- Reset State
		} else if choice == "3" {
			uc.SendMessage(phoneNumber, "Session cancelled. To start again, send `/start`.")
			// uc.CloseChat(phoneNumber) // Assuming you have a close function
			utils.ResetUserState(phoneNumber) // <-- Reset State
		}

	case StateAwaitingFormSubmission:
		if cmd, ok := data.(string); ok && cmd == "cancel" {
			welcomeMessage := "[1] Send message to pharmacy\n[2] Get Spreadsheet link\n[3] Cancel\n\nAnswer with number only!"
			uc.SendMessage(phoneNumber, "Request cancelled. Returning to the main menu.\n\n"+welcomeMessage)
			currentUserState.State = StateAwaitingMenuChoice // <-- State Transition
			return nil
		}
		// Store the original message text for confirmation
		currentUserState.PendingMessage = messageText
		confirmationPrompt := fmt.Sprintf("Please confirm the following request is correct:\n\n%s\n\nIs this correct? (Y/N)", messageText)
		uc.SendMessage(phoneNumber, confirmationPrompt)
		currentUserState.State = StateAwaitingConfirmation // <-- State Transition

	case StateAwaitingConfirmation:
		decision := data.(string)
		if decision == "Y" {
			patientDetails, err := utils.ParsePatientDetails(currentUserState.PendingMessage)
			if err != nil {
				uc.SendMessage(phoneNumber, "Error: The submitted data was malformed. Please try again.")

				formFormat := "Please send patient details in the format:\nDoctor Name: [Name]\nPatient Name: [Name]\nPatient Birth Date: [Date]\nMedication: [Drug]\nPatient Phone Number: [Number using 62]"
				uc.SendMessage(phoneNumber, formFormat)
				currentUserState.State = StateAwaitingFormSubmission
				return err
			}

			log.Printf(
				"Parsed Request: Doctor=%s, Patient=%s, Meds=%s, Phone=%s, DOB=%s",
				patientDetails.DoctorName,
				patientDetails.PatientName,
				patientDetails.Medication,
				patientDetails.PatientPhoneNumber,
				patientDetails.PatientBirthDate,
			)

			err = uc.sheetService.AddPrescriptionRow(patientDetails, Queue)
			if err != nil {
				// If it fails, tell the doctor but maybe still send to pharmacy
				uc.SendMessage(phoneNumber, "Note: Failed to save the record to the spreadsheet, but will still attempt to send to pharmacy.")
				// You can decide if you want to stop here or continue
			}

			// **SEND TO PHARMACY LOGIC HERE**
			pharmacyNumber := config.LoadConfig().PharmacyNumber
			// Send the pending message to the pharmacy number
			msgToPharmacy := fmt.Sprintf("New prescription request that need:\n%s \n\nWith Queue Number: %d\n\nThis Medicine for:\n%s\n%s\n%s\n\nFrom:\nDoctor %s", patientDetails.Medication, Queue, patientDetails.PatientName, patientDetails.PatientBirthDate, patientDetails.PatientPhoneNumber, patientDetails.DoctorName)
			err = uc.SendMessage(pharmacyNumber, msgToPharmacy)
			if err != nil {
				uc.SendMessage(phoneNumber, "Failed to send the request to the pharmacy. Please try again later.")
				return err
			}
			msgToPatient := fmt.Sprintf("Your prescription request for %s has been sent to the pharmacy. Your Number Queue is %d. Please wait for further updates from the pharmacy.", patientDetails.Medication, Queue)

			if DateNow != time.Now().Format("2006-01-02") {
				Queue = 1
				DateNow = time.Now().Format("2006-01-02")
			} else {
				Queue += 1
			}

			uc.SendMessage(phoneNumber, "Your request was sent to the pharmacy. Session complete.")
			uc.SendMessage(patientDetails.PatientPhoneNumber, msgToPatient)
			utils.ResetUserState(phoneNumber) // <-- Reset State
		} else if decision == "N" {
			formFormat := "Request cancelled. Please submit the form again with the correct details:\nDoctor Name: [Name]\nPatient Name: [Name]\nPatient Birth Date: [Date]\nMedication: [Drug]\nPatient Phone Number: [Number using 62]"
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
	req.SetBasicAuth("admin", "admin")

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
