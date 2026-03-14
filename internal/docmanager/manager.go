package docmanager

import "github.com/Y4NN777/doc-intel/internal/domain"

// Manager handles document lifecycle
type Manager interface {
	// Add registers a new document in a workspace
	Add(workspaceID string, path string) (*domain.Document, error)

	// List returns all documents in a workspace
	List(workspaceID string) ([]domain.Document, error)

	// Delete removes a document and all associated data
	Delete(documentID string) error

	// Reprocess triggers re-ingestion of a document
	Reprocess(documentID string) error

	// MarkRead updates the read status of a document
	MarkRead(documentID string, isRead bool) error
}
