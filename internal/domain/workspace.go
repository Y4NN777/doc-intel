package domain

import "time"

// Workspace represents an isolated collection of documents
type Workspace struct {
	ID         string
	Name       string
	CreatedAt  time.Time
	LastUsedAt time.Time
}

// IsActive returns true if this workspace is currently active
func (w *Workspace) IsActive() bool {
	// TODO: implement active workspace tracking
	return false
}
