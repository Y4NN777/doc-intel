package store

import (
	"github.com/Y4NN777/doc-intel/internal/domain"
)

// Store owns all persistent reads and writes
// Enforces INV-01: referential integrity
type Store interface {
	// Workspace operations
	CreateWorkspace(name string) (*domain.Workspace, error)
	ListWorkspaces() ([]domain.Workspace, error)
	GetWorkspace(id string) (*domain.Workspace, error)
	DeleteWorkspace(id string) error

	// Document operations
	WriteDocument(doc *domain.Document) error
	ListDocuments(workspaceID string) ([]domain.Document, error)
	GetDocument(id string) (*domain.Document, error)
	DeleteDocument(id string) error
	SetDocumentStatus(id string, status domain.DocStatus) error
	MarkDocumentRead(id string, isRead bool) error

	// Chunk operations
	WriteChunks(chunks []domain.Chunk) error
	QueryChunks(workspaceID string) ([]domain.Chunk, error)
	DeleteChunks(documentID string) error

	// Transaction control
	BeginTransaction() (Transaction, error)
}

// Transaction represents a database transaction
type Transaction interface {
	Commit() error
	Rollback() error
}
