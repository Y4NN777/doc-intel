package vectorindex

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Y4NN777/doc-intel/internal/domain"
	"github.com/google/uuid"
)

func setupTestIndex(t *testing.T) (*FAISSIndex, func()) {
	t.Helper()
	
	tmpDir := t.TempDir()
	
	index, err := NewFAISSIndex(tmpDir)
	if err != nil {
		t.Fatalf("failed to create test index: %v", err)
	}
	
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}
	
	return index, cleanup
}

// Test vector insertion
func TestVectorInsertion(t *testing.T) {
	index, cleanup := setupTestIndex(t)
	defer cleanup()
	
	workspaceID := uuid.New().String()
	
	vectors := []domain.Vector{
		{
			ChunkID:    uuid.New().String(),
			Values:     []float32{0.1, 0.2, 0.3, 0.4},
			Dimensions: 4,
		},
		{
			ChunkID:    uuid.New().String(),
			Values:     []float32{0.5, 0.6, 0.7, 0.8},
			Dimensions: 4,
		},
	}
	
	err := index.Insert(workspaceID, vectors)
	if err != nil {
		t.Fatalf("failed to insert vectors: %v", err)
	}
	
	// Verify metadata was saved
	meta, err := index.loadMetadata(workspaceID)
	if err != nil {
		t.Fatalf("failed to load metadata: %v", err)
	}
	
	if len(meta.Chunks) != 2 {
		t.Errorf("expected 2 chunks in metadata, got %d", len(meta.Chunks))
	}
	
	if meta.Dimensions != 4 {
		t.Errorf("expected dimensions 4, got %d", meta.Dimensions)
	}
}

// Test dimension validation
func TestDimensionValidation(t *testing.T) {
	index, cleanup := setupTestIndex(t)
	defer cleanup()
	
	workspaceID := uuid.New().String()
	
	// Insert first vector with 4 dimensions
	vectors1 := []domain.Vector{
		{
			ChunkID:    uuid.New().String(),
			Values:     []float32{0.1, 0.2, 0.3, 0.4},
			Dimensions: 4,
		},
	}
	
	err := index.Insert(workspaceID, vectors1)
	if err != nil {
		t.Fatalf("failed to insert first vector: %v", err)
	}
	
	// Attempt to insert vector with different dimensions
	vectors2 := []domain.Vector{
		{
			ChunkID:    uuid.New().String(),
			Values:     []float32{0.1, 0.2, 0.3},
			Dimensions: 3,
		},
	}
	
	err = index.Insert(workspaceID, vectors2)
	if err == nil {
		t.Error("expected error when inserting vector with different dimensions, got nil")
	}
}

// Test workspace deletion
func TestWorkspaceDeletion(t *testing.T) {
	index, cleanup := setupTestIndex(t)
	defer cleanup()
	
	workspaceID := uuid.New().String()
	
	// Insert vectors
	vectors := []domain.Vector{
		{
			ChunkID:    uuid.New().String(),
			Values:     []float32{0.1, 0.2, 0.3, 0.4},
			Dimensions: 4,
		},
	}
	
	err := index.Insert(workspaceID, vectors)
	if err != nil {
		t.Fatalf("failed to insert vectors: %v", err)
	}
	
	// Verify workspace directory exists
	wsDir := index.workspaceDir(workspaceID)
	if _, err := os.Stat(wsDir); os.IsNotExist(err) {
		t.Fatal("workspace directory should exist after insertion")
	}
	
	// Delete workspace
	err = index.DeleteWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("failed to delete workspace: %v", err)
	}
	
	// Verify workspace directory is deleted
	if _, err := os.Stat(wsDir); !os.IsNotExist(err) {
		t.Error("workspace directory should not exist after deletion")
	}
}

// Test workspace isolation
func TestWorkspaceIsolation(t *testing.T) {
	index, cleanup := setupTestIndex(t)
	defer cleanup()
	
	workspace1 := uuid.New().String()
	workspace2 := uuid.New().String()
	
	// Insert vectors into workspace 1
	vectors1 := []domain.Vector{
		{
			ChunkID:    uuid.New().String(),
			Values:     []float32{0.1, 0.2, 0.3, 0.4},
			Dimensions: 4,
		},
	}
	
	err := index.Insert(workspace1, vectors1)
	if err != nil {
		t.Fatalf("failed to insert vectors into workspace 1: %v", err)
	}
	
	// Insert vectors into workspace 2
	vectors2 := []domain.Vector{
		{
			ChunkID:    uuid.New().String(),
			Values:     []float32{0.5, 0.6, 0.7, 0.8},
			Dimensions: 4,
		},
	}
	
	err = index.Insert(workspace2, vectors2)
	if err != nil {
		t.Fatalf("failed to insert vectors into workspace 2: %v", err)
	}
	
	// Verify metadata is isolated
	meta1, err := index.loadMetadata(workspace1)
	if err != nil {
		t.Fatalf("failed to load metadata for workspace 1: %v", err)
	}
	
	meta2, err := index.loadMetadata(workspace2)
	if err != nil {
		t.Fatalf("failed to load metadata for workspace 2: %v", err)
	}
	
	if len(meta1.Chunks) != 1 {
		t.Errorf("expected 1 chunk in workspace 1, got %d", len(meta1.Chunks))
	}
	
	if len(meta2.Chunks) != 1 {
		t.Errorf("expected 1 chunk in workspace 2, got %d", len(meta2.Chunks))
	}
	
	if meta1.Chunks[0].ChunkID == meta2.Chunks[0].ChunkID {
		t.Error("chunks in different workspaces should have different IDs")
	}
}

// Test empty vector insertion
func TestEmptyVectorInsertion(t *testing.T) {
	index, cleanup := setupTestIndex(t)
	defer cleanup()
	
	workspaceID := uuid.New().String()
	
	// Insert empty vector slice
	err := index.Insert(workspaceID, []domain.Vector{})
	if err != nil {
		t.Errorf("inserting empty vector slice should not error, got: %v", err)
	}
}

// Test search on empty workspace
func TestSearchEmptyWorkspace(t *testing.T) {
	index, cleanup := setupTestIndex(t)
	defer cleanup()
	
	workspaceID := uuid.New().String()
	
	query := []float32{0.1, 0.2, 0.3, 0.4}
	chunkIDs, scores, err := index.Search(workspaceID, query, 5)
	if err != nil {
		t.Fatalf("search on empty workspace should not error, got: %v", err)
	}
	
	if len(chunkIDs) != 0 {
		t.Errorf("expected 0 results from empty workspace, got %d", len(chunkIDs))
	}
	
	if len(scores) != 0 {
		t.Errorf("expected 0 scores from empty workspace, got %d", len(scores))
	}
}

// Test metadata file structure
func TestMetadataFileStructure(t *testing.T) {
	index, cleanup := setupTestIndex(t)
	defer cleanup()
	
	workspaceID := uuid.New().String()
	
	vectors := []domain.Vector{
		{
			ChunkID:    uuid.New().String(),
			Values:     []float32{0.1, 0.2, 0.3, 0.4},
			Dimensions: 4,
		},
	}
	
	err := index.Insert(workspaceID, vectors)
	if err != nil {
		t.Fatalf("failed to insert vectors: %v", err)
	}
	
	// Verify metadata file exists
	metaPath := index.metadataPath(workspaceID)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error("metadata file should exist after insertion")
	}
	
	// Verify index path is correct
	indexPath := index.indexPath(workspaceID)
	expectedPath := filepath.Join(index.workspaceDir(workspaceID), "index.faiss")
	if indexPath != expectedPath {
		t.Errorf("expected index path %s, got %s", expectedPath, indexPath)
	}
}
