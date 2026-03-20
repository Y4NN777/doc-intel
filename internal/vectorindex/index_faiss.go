// +build faiss

package vectorindex

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/DataIntelligenceCrew/go-faiss"
	"github.com/Y4NN777/doc-intel/internal/domain"
)

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

// FAISSIndex implements Index using FAISS library via CGO
type FAISSIndex struct {
	baseDir string
	mu      sync.RWMutex
	// In-memory FAISS indexes per workspace
	indexes map[string]faiss.Index // workspaceID -> FAISS index
}

// WorkspaceMetadata tracks chunk_id to vector_index mappings
type WorkspaceMetadata struct {
	WorkspaceID string        `json:"workspace_id"`
	Dimensions  int           `json:"dimensions"`
	Chunks      []ChunkVector `json:"chunks"`
}

// ChunkVector maps a chunk ID to its position in the FAISS index
type ChunkVector struct {
	ChunkID     string `json:"chunk_id"`
	DocumentID  string `json:"document_id"`
	VectorIndex int    `json:"vector_index"`
}

// NewFAISSIndex creates a new vector index with FAISS CGO bindings
func NewFAISSIndex(baseDir string) (*FAISSIndex, error) {
	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(homeDir, ".docintel", "workspaces")
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &FAISSIndex{
		baseDir: baseDir,
		indexes: make(map[string]faiss.Index),
	}, nil
}

// workspaceDir returns the directory path for a workspace
func (f *FAISSIndex) workspaceDir(workspaceID string) string {
	return filepath.Join(f.baseDir, workspaceID)
}

// metadataPath returns the path to the metadata file for a workspace
func (f *FAISSIndex) metadataPath(workspaceID string) string {
	return filepath.Join(f.workspaceDir(workspaceID), "metadata.json")
}

// indexPath returns the path to the FAISS index file for a workspace
func (f *FAISSIndex) indexPath(workspaceID string) string {
	return filepath.Join(f.workspaceDir(workspaceID), "index.faiss")
}

// loadMetadata loads the metadata for a workspace
func (f *FAISSIndex) loadMetadata(workspaceID string) (*WorkspaceMetadata, error) {
	metaPath := f.metadataPath(workspaceID)
	
	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &WorkspaceMetadata{
				WorkspaceID: workspaceID,
				Chunks:      []ChunkVector{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var meta WorkspaceMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &meta, nil
}

// saveMetadata saves the metadata for a workspace
func (f *FAISSIndex) saveMetadata(meta *WorkspaceMetadata) error {
	wsDir := f.workspaceDir(meta.WorkspaceID)
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		return fmt.Errorf("failed to create workspace directory: %w", err)
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	metaPath := f.metadataPath(meta.WorkspaceID)
	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// Insert adds vectors to the index for a workspace
func (f *FAISSIndex) Insert(workspaceID string, vectors []domain.Vector) error {
	if len(vectors) == 0 {
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// Load existing metadata
	meta, err := f.loadMetadata(workspaceID)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	// Set dimensions from first vector if not set
	if meta.Dimensions == 0 && len(vectors) > 0 {
		meta.Dimensions = vectors[0].Dimensions
	}

	// Validate dimensions
	for _, v := range vectors {
		if v.Dimensions != meta.Dimensions {
			return fmt.Errorf("dimension mismatch: expected %d, got %d", meta.Dimensions, v.Dimensions)
		}
		if len(v.Values) != v.Dimensions {
			return fmt.Errorf("vector values length %d does not match dimensions %d", len(v.Values), v.Dimensions)
		}
	}

	// Load or create FAISS index
	idx, err := f.loadIndex(workspaceID, meta.Dimensions)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	// Prepare vectors for FAISS (flatten to single slice)
	flatVectors := make([]float32, 0, len(vectors)*meta.Dimensions)
	for _, v := range vectors {
		flatVectors = append(flatVectors, v.Values...)
	}

	// Add vectors to FAISS index
	if err := idx.Add(flatVectors); err != nil {
		return fmt.Errorf("failed to add vectors to FAISS index: %w", err)
	}

	// Update metadata with new chunks
	startIndex := len(meta.Chunks)
	for i, v := range vectors {
		meta.Chunks = append(meta.Chunks, ChunkVector{
			ChunkID:     v.ChunkID,
			DocumentID:  "",
			VectorIndex: startIndex + i,
		})
	}

	// Persist index and metadata
	if err := f.saveIndex(workspaceID, idx); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	if err := f.saveMetadata(meta); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// Search performs ANN search scoped to a workspace using FAISS
func (f *FAISSIndex) Search(workspaceID string, query []float32, k int) ([]string, []float64, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Load metadata
	meta, err := f.loadMetadata(workspaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load metadata: %w", err)
	}

	if len(meta.Chunks) == 0 {
		return []string{}, []float64{}, nil
	}

	// Validate query dimensions
	if len(query) != meta.Dimensions {
		return nil, nil, fmt.Errorf("query dimension %d does not match index dimension %d", len(query), meta.Dimensions)
	}

	// Load FAISS index
	idx, exists := f.indexes[workspaceID]
	if !exists {
		f.mu.RUnlock()
		f.mu.Lock()
		var loadErr error
		idx, loadErr = f.loadIndex(workspaceID, meta.Dimensions)
		f.mu.Unlock()
		f.mu.RLock()
		
		if loadErr != nil {
			return nil, nil, fmt.Errorf("failed to load index: %w", loadErr)
		}
	}

	// Perform FAISS search
	distances, labels, err := idx.Search(query, int64(k))
	if err != nil {
		return nil, nil, fmt.Errorf("FAISS search failed: %w", err)
	}

	// Map FAISS labels (vector indices) to chunk IDs
	chunkIDs := make([]string, 0, k)
	scores := make([]float64, 0, k)
	
	for i := 0; i < len(labels) && i < k; i++ {
		vectorIdx := int(labels[i])
		if vectorIdx < 0 || vectorIdx >= len(meta.Chunks) {
			continue
		}
		
		chunkIDs = append(chunkIDs, meta.Chunks[vectorIdx].ChunkID)
		// Convert L2 distance to similarity score (inverse)
		scores = append(scores, 1.0/(1.0+float64(distances[i])))
	}

	return chunkIDs, scores, nil
}

// Delete removes all vectors for a document across all workspaces
func (f *FAISSIndex) Delete(documentID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	entries, err := os.ReadDir(f.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read base directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		workspaceID := entry.Name()
		
		meta, err := f.loadMetadata(workspaceID)
		if err != nil {
			continue
		}

		// Find chunks belonging to this document
		toRemove := make([]int64, 0)
		newChunks := make([]ChunkVector, 0)
		
		for _, chunk := range meta.Chunks {
			if chunk.DocumentID == documentID {
				toRemove = append(toRemove, int64(chunk.VectorIndex))
			} else {
				newChunks = append(newChunks, chunk)
			}
		}

		if len(toRemove) == 0 {
			continue
		}

		// Load FAISS index
		idx, err := f.loadIndex(workspaceID, meta.Dimensions)
		if err != nil {
			return fmt.Errorf("failed to load index for workspace %s: %w", workspaceID, err)
		}

		// Remove vectors from FAISS index
		selector, err := faiss.NewIDSelectorBatch(toRemove)
		if err != nil {
			return fmt.Errorf("failed to create ID selector: %w", err)
		}
		defer selector.Delete()

		if _, err := idx.RemoveIDs(selector); err != nil {
			return fmt.Errorf("failed to remove IDs from FAISS index: %w", err)
		}

		// Update metadata
		meta.Chunks = newChunks
		for i := range meta.Chunks {
			meta.Chunks[i].VectorIndex = i
		}

		if err := f.saveIndex(workspaceID, idx); err != nil {
			return fmt.Errorf("failed to save index for workspace %s: %w", workspaceID, err)
		}

		if err := f.saveMetadata(meta); err != nil {
			return fmt.Errorf("failed to save metadata for workspace %s: %w", workspaceID, err)
		}
	}

	return nil
}

// DeleteWorkspace removes all vectors for a workspace
func (f *FAISSIndex) DeleteWorkspace(workspaceID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if idx, exists := f.indexes[workspaceID]; exists {
		idx.Delete()
		delete(f.indexes, workspaceID)
	}
	
	wsDir := f.workspaceDir(workspaceID)
	if err := os.RemoveAll(wsDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete workspace directory: %w", err)
	}

	return nil
}

// loadIndex loads or creates a FAISS index for a workspace
func (f *FAISSIndex) loadIndex(workspaceID string, dimensions int) (faiss.Index, error) {
	if idx, exists := f.indexes[workspaceID]; exists {
		return idx, nil
	}

	indexPath := f.indexPath(workspaceID)
	
	// Try to load existing index from disk
	if _, err := os.Stat(indexPath); err == nil {
		idx, err := faiss.ReadIndex(indexPath, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to read FAISS index: %w", err)
		}
		f.indexes[workspaceID] = idx
		return idx, nil
	}

	// Create new FAISS index (IndexFlatL2 for L2 distance)
	idx, err := faiss.NewIndexFlatL2(dimensions)
	if err != nil {
		return nil, fmt.Errorf("failed to create FAISS index: %w", err)
	}

	f.indexes[workspaceID] = idx
	return idx, nil
}

// saveIndex persists a FAISS index to disk
func (f *FAISSIndex) saveIndex(workspaceID string, idx faiss.Index) error {
	wsDir := f.workspaceDir(workspaceID)
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		return fmt.Errorf("failed to create workspace directory: %w", err)
	}

	indexPath := f.indexPath(workspaceID)
	if err := faiss.WriteIndex(idx, indexPath); err != nil {
		return fmt.Errorf("failed to write FAISS index: %w", err)
	}

	return nil
}
