package domain

import "time"

// Session represents a conversation thread in a workspace
type Session struct {
	ID           string
	WorkspaceID  string
	StartedAt    time.Time
	LastActiveAt time.Time
	Title        string
}

// IsActive returns true if session is currently active
func (s *Session) IsActive() bool {
	// TODO: implement active session tracking
	return false
}

// Turn represents a single question-answer exchange
type Turn struct {
	ID        string
	SessionID string
	Question  string
	Answer    Answer
	CreatedAt time.Time
}
