package retriever

import "github.com/Y4NN777/doc-intel/internal/domain"

// Retriever performs scoped search operations
// Enforces INV-02, INV-09: workspace scope is always explicit
type Retriever interface {
	// Search performs hybrid search (semantic + keyword) in a workspace
	Search(workspaceID string, query string, k int) ([]domain.ScoredChunk, error)

	// SearchDoc performs search scoped to a single document
	SearchDoc(documentID string, query string, k int) ([]domain.ScoredChunk, error)

	// SearchMultiDoc performs search optimized for cross-document queries
	SearchMultiDoc(workspaceID string, query string, k int) ([]domain.ScoredChunk, error)
}
