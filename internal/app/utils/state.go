package utils

import "sync"

// --- Define the conversation states as constants for safety ---
const (
	StateAwaitingStart        = "AWAITING_START"
	StateAwaitingMenuChoice   = "AWAITING_MENU_CHOICE"
	StateAwaitingFormSubmission = "AWAITING_FORM_SUBMISSION"
	StateAwaitingConfirmation = "AWAITING_CONFIRMATION"
)

// --- UserState holds the conversation context for a single user ---
type UserState struct {
	State          string
	PendingMessage string // Used to temporarily store the form message for confirmation
}

// --- In-memory store for user states. Replace with a database in production. ---
var (
	userStates = make(map[string]*UserState)
	mu         sync.Mutex // Mutex to prevent race conditions when accessing the map
)

// getOrCreateUserState retrieves the state for a user, creating it if it doesn't exist.
func GetOrCreateUserState(phoneNumber string) *UserState {
	mu.Lock()
	defer mu.Unlock()

	if state, exists := userStates[phoneNumber]; exists {
		return state
	}

	// User doesn't exist, create a new state
	newState := &UserState{State: StateAwaitingStart}
	userStates[phoneNumber] = newState
	return newState
}

// resetUserState resets a user's state to the beginning.
func ResetUserState(phoneNumber string) {
    mu.Lock()
	defer mu.Unlock()
    // We can just create a new one, letting the garbage collector handle the old one.
    userStates[phoneNumber] = &UserState{State: StateAwaitingStart}
}