package domain

// ChunkSource indicates how text was extracted
type ChunkSource string

const (
	ChunkSourceTextLayer ChunkSource = "text_layer"
	ChunkSourceOCR       ChunkSource = "ocr"
)

// Chunk represents a text segment from a document
type Chunk struct {
	ID         string
	DocumentID string
	PageNumber int
	Text       string
	TokenCount int
	Source     ChunkSource
}

// Vector represents the embedding for a chunk
type Vector struct {
	ChunkID    string
	Values     []float32
	Dimensions int
}
