package domain

// QueryType represents the type of query operation
type QueryType string

const (
	QueryTypeQuestion   QueryType = "question"
	QueryTypeSummary    QueryType = "summary"
	QueryTypeExtraction QueryType = "extraction"
	QueryTypeComparison QueryType = "comparison"
)

// Confidence represents answer confidence level
type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

// QueryRequest represents a user query
type QueryRequest struct {
	WorkspaceID string
	DocumentID  *string // optional - scopes to single document
	Text        string
	SessionID   *string // optional - for conversation context
	Type        QueryType
}

// Answer represents a query response
type Answer struct {
	Text       string
	Confidence Confidence
	Sources    []Citation
	SessionID  string
}

// HasSources returns true if answer includes citations
func (a *Answer) HasSources() bool {
	return len(a.Sources) > 0
}

// Citation links an answer to source content
type Citation struct {
	ChunkID    string
	DocName    string
	PageNumber int
	Excerpt    string
}

// ScoredChunk represents a retrieved chunk with relevance score
type ScoredChunk struct {
	Chunk Chunk
	Score float64
}
