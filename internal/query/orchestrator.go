package query

import "github.com/Y4NN777/doc-intel/internal/domain"

// Orchestrator handles query reasoning and answer generation
// Enforces INV-04: citations from retrieved content only
type Orchestrator interface {
	// Query processes a natural language question
	Query(req domain.QueryRequest) (*domain.Answer, error)

	// Summarize generates a document summary
	Summarize(documentID string, workspaceID string) (*domain.Answer, error)

	// Extract performs targeted data extraction
	Extract(req domain.QueryRequest) (*domain.Answer, error)

	// Compare synthesizes information across documents
	Compare(req domain.QueryRequest) (*domain.Answer, error)
}
