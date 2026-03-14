package session

import "github.com/Y4NN777/doc-intel/internal/domain"

// Manager handles conversation history
// Enforces INV-08: append-only history
type Manager interface {
	// GetContext retrieves prior turns for a session
	GetContext(sessionID string) ([]domain.Turn, error)

	// List returns all sessions in a workspace
	List(workspaceID string) ([]domain.Session, error)

	// Load retrieves all turns from a past session
	Load(sessionID string) ([]domain.Turn, error)

	// AppendTurn adds a new turn to session history
	AppendTurn(sessionID string, turn domain.Turn) error

	// CreateSession starts a new conversation session
	CreateSession(workspaceID string, title string) (*domain.Session, error)
}
