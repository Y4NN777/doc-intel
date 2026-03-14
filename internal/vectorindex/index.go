package vectorindex

import "github.com/Y4NN777/doc-intel/internal/domain"

// Index provides ANN search over chunk embeddings using FAISS
// Enforces INV-02, INV-09: workspace-scoped search
type Index interface {
	// Insert adds vectors to the index for a workspace
	Insert(workspaceID string, vectors []domain.Vector) error

	// Search performs ANN search scoped to a workspace
	Search(workspaceID string, query []float32, k int) ([]string, []float64, error)

	// Delete removes all vectors for a document
	Delete(documentID string) error

	// DeleteWorkspace removes all vectors for a workspace
	DeleteWorkspace(workspaceID string) error
}
