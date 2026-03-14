package workspace

import "github.com/Y4NN777/doc-intel/internal/domain"

// Manager owns workspace lifecycle
// Enforces INV-06: total workspace deletion
type Manager interface {
	// Create creates a new workspace
	Create(name string) (*domain.Workspace, error)

	// List returns all workspaces
	List() ([]domain.Workspace, error)

	// Switch sets the active workspace
	Switch(name string) (*domain.Workspace, error)

	// Delete removes a workspace and all associated data atomically
	Delete(name string) error

	// GetActive returns the currently active workspace
	GetActive() (*domain.Workspace, error)
}
