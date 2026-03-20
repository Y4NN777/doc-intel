package domain

import (
	"testing"
	"time"
)

// TestDomainEntitiesCompile verifies all domain entities compile correctly
func TestDomainEntitiesCompile(t *testing.T) {
	// Test Workspace
	ws := Workspace{
		ID:         "ws-1",
		Name:       "test-workspace",
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}
	if ws.ID == "" {
		t.Error("Workspace ID should not be empty")
	}

	// Test Document with all statuses
	doc := Document{
		ID:          "doc-1",
		WorkspaceID: "ws-1",
		Path:        "/path/to/doc.pdf",
		Language:    LanguageEN,
		Status:      DocStatusPending,
		PageCount:   10,
		IsRead:      false,
		CreatedAt:   time.Now(),
		ProcessedAt: nil,
	}
	if !doc.IsPending() {
		t.Error("Document should be pending")
	}

	// Test all DocStatus values
	statuses := []DocStatus{DocStatusPending, DocStatusIndexed, DocStatusFailed}
	if len(statuses) != 3 {
		t.Error("Expected 3 document statuses")
	}

	// Test all Language values
	languages := []Language{LanguageEN, LanguageFR, LanguageUnknown}
	if len(languages) != 3 {
		t.Error("Expected 3 languages")
	}

	// Test Chunk with all sources
	chunk := Chunk{
		ID:         "chunk-1",
		DocumentID: "doc-1",
		PageNumber: 1,
		Text:       "Sample text",
		TokenCount: 10,
		Source:     ChunkSourceTextLayer,
	}
	if chunk.TokenCount > 512 {
		t.Error("Chunk token count should be <= 512")
	}

	// Test all ChunkSource values
	sources := []ChunkSource{ChunkSourceTextLayer, ChunkSourceOCR}
	if len(sources) != 2 {
		t.Error("Expected 2 chunk sources")
	}

	// Test Vector
	vector := Vector{
		ChunkID:    "chunk-1",
		Values:     []float32{0.1, 0.2, 0.3},
		Dimensions: 3,
	}
	if len(vector.Values) != vector.Dimensions {
		t.Error("Vector dimensions should match values length")
	}

	// Test Session
	session := Session{
		ID:           "session-1",
		WorkspaceID:  "ws-1",
		StartedAt:    time.Now(),
		LastActiveAt: time.Now(),
		Title:        "Test Session",
	}
	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}

	// Test Turn
	turn := Turn{
		ID:        "turn-1",
		SessionID: "session-1",
		Question:  "What is this about?",
		Answer: Answer{
			Text:       "This is a test answer",
			Confidence: ConfidenceHigh,
			Sources:    []Citation{},
			SessionID:  "session-1",
		},
		CreatedAt: time.Now(),
	}
	if turn.Question == "" {
		t.Error("Turn question should not be empty")
	}

	// Test all QueryType values
	queryTypes := []QueryType{
		QueryTypeQuestion,
		QueryTypeSummary,
		QueryTypeExtraction,
		QueryTypeComparison,
	}
	if len(queryTypes) != 4 {
		t.Error("Expected 4 query types")
	}

	// Test all Confidence values
	confidences := []Confidence{ConfidenceHigh, ConfidenceMedium, ConfidenceLow}
	if len(confidences) != 3 {
		t.Error("Expected 3 confidence levels")
	}

	// Test QueryRequest
	docID := "doc-1"
	sessionID := "session-1"
	queryReq := QueryRequest{
		WorkspaceID: "ws-1",
		DocumentID:  &docID,
		Text:        "What is this about?",
		SessionID:   &sessionID,
		Type:        QueryTypeQuestion,
	}
	if queryReq.WorkspaceID == "" {
		t.Error("QueryRequest workspace ID should not be empty")
	}

	// Test Answer
	answer := Answer{
		Text:       "This is a test answer",
		Confidence: ConfidenceHigh,
		Sources: []Citation{
			{
				ChunkID:    "chunk-1",
				DocName:    "test.pdf",
				PageNumber: 1,
				Excerpt:    "Sample excerpt",
			},
		},
		SessionID: "session-1",
	}
	if !answer.HasSources() {
		t.Error("Answer should have sources")
	}

	// Test Citation
	citation := Citation{
		ChunkID:    "chunk-1",
		DocName:    "test.pdf",
		PageNumber: 1,
		Excerpt:    "Sample excerpt",
	}
	if citation.PageNumber < 1 {
		t.Error("Citation page number should be >= 1")
	}

	// Test ScoredChunk
	scoredChunk := ScoredChunk{
		Chunk: chunk,
		Score: 0.95,
	}
	if scoredChunk.Score < 0.0 || scoredChunk.Score > 1.0 {
		t.Error("ScoredChunk score should be between 0.0 and 1.0")
	}
}

// TestEnumerationValues verifies all enumeration constants are defined
func TestEnumerationValues(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		// DocStatus
		{"DocStatusPending", DocStatusPending, DocStatus("pending")},
		{"DocStatusIndexed", DocStatusIndexed, DocStatus("indexed")},
		{"DocStatusFailed", DocStatusFailed, DocStatus("failed")},

		// Language
		{"LanguageEN", LanguageEN, Language("en")},
		{"LanguageFR", LanguageFR, Language("fr")},
		{"LanguageUnknown", LanguageUnknown, Language("unknown")},

		// ChunkSource
		{"ChunkSourceTextLayer", ChunkSourceTextLayer, ChunkSource("text_layer")},
		{"ChunkSourceOCR", ChunkSourceOCR, ChunkSource("ocr")},

		// QueryType
		{"QueryTypeQuestion", QueryTypeQuestion, QueryType("question")},
		{"QueryTypeSummary", QueryTypeSummary, QueryType("summary")},
		{"QueryTypeExtraction", QueryTypeExtraction, QueryType("extraction")},
		{"QueryTypeComparison", QueryTypeComparison, QueryType("comparison")},

		// Confidence
		{"ConfidenceHigh", ConfidenceHigh, Confidence("high")},
		{"ConfidenceMedium", ConfidenceMedium, Confidence("medium")},
		{"ConfidenceLow", ConfidenceLow, Confidence("low")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.value, tt.expected)
			}
		})
	}
}
