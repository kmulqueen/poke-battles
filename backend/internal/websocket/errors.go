package websocket

import "encoding/json"

// ErrorCode represents a protocol error code
type ErrorCode string

// Error codes
const (
	ErrCodeAuthRequired      ErrorCode = "AUTH_REQUIRED"
	ErrCodeAuthFailed        ErrorCode = "AUTH_FAILED"
	ErrCodeSessionExpired    ErrorCode = "SESSION_EXPIRED"
	ErrCodeLobbyNotFound     ErrorCode = "LOBBY_NOT_FOUND"
	ErrCodeLobbyFull         ErrorCode = "LOBBY_FULL"
	ErrCodeInvalidState      ErrorCode = "INVALID_STATE"
	ErrCodeInvalidAction     ErrorCode = "INVALID_ACTION"
	ErrCodeNotYourTurn       ErrorCode = "NOT_YOUR_TURN"
	ErrCodeTurnMismatch      ErrorCode = "TURN_MISMATCH"
	ErrCodeActionTimeout     ErrorCode = "ACTION_TIMEOUT"
	ErrCodeMalformedMessage  ErrorCode = "MALFORMED_MESSAGE"
	ErrCodeVersionMismatch   ErrorCode = "VERSION_MISMATCH"
	ErrCodeInternalError     ErrorCode = "INTERNAL_ERROR"
	ErrCodePlayerNotInLobby  ErrorCode = "PLAYER_NOT_IN_LOBBY"
)

// ErrorPayload is the payload for error messages
type ErrorPayload struct {
	Code        ErrorCode       `json:"code"`
	Message     string          `json:"message"`
	Details     json.RawMessage `json:"details,omitempty"`
	Recoverable bool            `json:"recoverable"`
}

// IsRecoverable returns whether an error code is recoverable
func IsRecoverable(code ErrorCode) bool {
	switch code {
	case ErrCodeInvalidState, ErrCodeInvalidAction, ErrCodeNotYourTurn,
		ErrCodeTurnMismatch, ErrCodeMalformedMessage:
		return true
	default:
		return false
	}
}

// NewErrorPayload creates a new error payload
func NewErrorPayload(code ErrorCode, message string) ErrorPayload {
	return ErrorPayload{
		Code:        code,
		Message:     message,
		Recoverable: IsRecoverable(code),
	}
}

// NewErrorPayloadWithDetails creates a new error payload with details
func NewErrorPayloadWithDetails(code ErrorCode, message string, details interface{}) (ErrorPayload, error) {
	payload := NewErrorPayload(code, message)
	if details != nil {
		detailsBytes, err := json.Marshal(details)
		if err != nil {
			return payload, err
		}
		payload.Details = detailsBytes
	}
	return payload, nil
}

// Common error payloads
var (
	ErrPayloadAuthRequired = NewErrorPayload(
		ErrCodeAuthRequired,
		"Authentication required before sending messages",
	)

	ErrPayloadMalformedMessage = NewErrorPayload(
		ErrCodeMalformedMessage,
		"Could not parse message",
	)

	ErrPayloadVersionMismatch = NewErrorPayload(
		ErrCodeVersionMismatch,
		"Protocol version not supported",
	)
)
