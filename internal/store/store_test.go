//go:build fts5

package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Y4NN777/doc-intel/internal/domain"
	"github.com/google/uuid"
)

func setupTestStore(t *testing.T) (*SQLiteStore, func()) {
	t.Helper()
	
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	
	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}
	
	return store, cleanup
}

// Test workspace CRUD operations
func TestWorkspaceOperations(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	
	// Create workspace
	ws, err := store.CreateWorkspace("test-workspace")
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}
	
	if ws.Name != "test-workspace" {
		t.Errorf("expected workspace name 'test-workspace', got %s", ws.Name)
	}
	
	// List workspaces
	workspaces, err := store.ListWorkspaces()
	if err != nil {
		t.Fatalf("failed to list workspaces: %v", err)
	}
	
	if len(workspaces) != 1 {
		t.Errorf("expected 1 workspace, got %d", len(workspaces))
	}
	
	// Get workspace
	retrieved, err := store.GetWorkspace(ws.ID)
	if err != nil {
		t.Fatalf("failed to get workspace: %v", err)
	}
	
	if retrieved.ID != ws.ID {
		t.Errorf("expected workspace ID %s, got %s", ws.ID, retrieved.ID)
	}
	
	// Delete workspace
	err = store.DeleteWorkspace(ws.ID)
	if err != nil {
		t.Fatalf("failed to delete workspace: %v", err)
	}
	
	// Verify deletion
	workspaces, err = store.ListWorkspaces()
	if err != nil {
		t.Fatalf("failed to list workspaces after deletion: %v", err)
	}
	
	if len(workspaces) != 0 {
		t.Errorf("expected 0 workspaces after deletion, got %d", len(workspaces))
	}
}

// Test workspace name uniqueness
func TestWorkspaceNameUniqueness(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	
	// Create first workspace
	_, err := store.CreateWorkspace("duplicate-name")
	if err != nil {
		t.Fatalf("failed to create first workspace: %v", err)
	}
	
	// Attempt to create second workspace with same name
	_, err = store.CreateWorkspace("duplicate-name")
	if err == nil {
		t.Error("expected error when creating workspace with duplicate name, got nil")
	}
}

// Test document operations
func TestDocumentOperations(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	
	// Create workspace first
	ws, err := store.CreateWorkspace("test-workspace")
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}
	
	// Create document
	doc := &domain.Document{
		ID:          uuid.New().String(),
		WorkspaceID: ws.ID,
		Path:        "/path/to/test.pdf",
		Language:    domain.LanguageEN,
		Status:      domain.DocStatusPending,
		PageCount:   10,
		IsRead:      false,
		CreatedAt:   time.Now(),
	}
	
	err = store.WriteDocument(doc)
	if err != nil {
		t.Fatalf("failed to write document: %v", err)
	}
	
	// List documents
	docs, err := store.ListDocuments(ws.ID)
	if err != nil {
		t.Fatalf("failed to list documents: %v", err)
	}
	
	if len(docs) != 1 {
		t.Errorf("expected 1 document, got %d", len(docs))
	}
	
	// Get document
	retrieved, err := store.GetDocument(doc.ID)
	if err != nil {
		t.Fatalf("failed to get document: %v", err)
	}
	
	if retrieved.ID != doc.ID {
		t.Errorf("expected document ID %s, got %s", doc.ID, retrieved.ID)
	}
	
	// Update document status
	err = store.SetDocumentStatus(doc.ID, domain.DocStatusIndexed)
	if err != nil {
		t.Fatalf("failed to set document status: %v", err)
	}
	
	retrieved, err = store.GetDocument(doc.ID)
	if err != nil {
		t.Fatalf("failed to get document after status update: %v", err)
	}
	
	if retrieved.Status != domain.DocStatusIndexed {
		t.Errorf("expected status %s, got %s", domain.DocStatusIndexed, retrieved.Status)
	}
	
	if retrieved.ProcessedAt == nil {
		t.Error("expected ProcessedAt to be set when status is indexed")
	}
	
	// Mark document as read
	err = store.MarkDocumentRead(doc.ID, true)
	if err != nil {
		t.Fatalf("failed to mark document as read: %v", err)
	}
	
	retrieved, err = store.GetDocument(doc.ID)
	if err != nil {
		t.Fatalf("failed to get document after marking read: %v", err)
	}
	
	if !retrieved.IsRead {
		t.Error("expected document to be marked as read")
	}
	
	// Delete document
	err = store.DeleteDocument(doc.ID)
	if err != nil {
		t.Fatalf("failed to delete document: %v", err)
	}
	
	// Verify deletion
	docs, err = store.ListDocuments(ws.ID)
	if err != nil {
		t.Fatalf("failed to list documents after deletion: %v", err)
	}
	
	if len(docs) != 0 {
		t.Errorf("expected 0 documents after deletion, got %d", len(docs))
	}
}

// Test cascade delete for workspace
func TestWorkspaceCascadeDelete(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	
	// Create workspace
	ws, err := store.CreateWorkspace("test-workspace")
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}
	
	// Create document
	doc := &domain.Document{
		ID:          uuid.New().String(),
		WorkspaceID: ws.ID,
		Path:        "/path/to/test.pdf",
		Language:    domain.LanguageEN,
		Status:      domain.DocStatusPending,
		PageCount:   10,
		IsRead:      false,
		CreatedAt:   time.Now(),
	}
	
	err = store.WriteDocument(doc)
	if err != nil {
		t.Fatalf("failed to write document: %v", err)
	}
	
	// Create chunks
	chunks := []domain.Chunk{
		{
			ID:         uuid.New().String(),
			DocumentID: doc.ID,
			PageNumber: 1,
			Text:       "Test chunk 1",
			TokenCount: 10,
			Source:     domain.ChunkSourceTextLayer,
		},
		{
			ID:         uuid.New().String(),
			DocumentID: doc.ID,
			PageNumber: 2,
			Text:       "Test chunk 2",
			TokenCount: 15,
			Source:     domain.ChunkSourceTextLayer,
		},
	}
	
	err = store.WriteChunks(chunks)
	if err != nil {
		t.Fatalf("failed to write chunks: %v", err)
	}
	
	// Create session
	session := &domain.Session{
		ID:           uuid.New().String(),
		WorkspaceID:  ws.ID,
		StartedAt:    time.Now(),
		LastActiveAt: time.Now(),
		Title:        "Test session",
	}
	
	err = store.CreateSession(session)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	
	// Delete workspace (should cascade delete documents, chunks, and sessions)
	err = store.DeleteWorkspace(ws.ID)
	if err != nil {
		t.Fatalf("failed to delete workspace: %v", err)
	}
	
	// Verify documents are deleted
	docs, err := store.ListDocuments(ws.ID)
	if err != nil {
		t.Fatalf("failed to list documents after workspace deletion: %v", err)
	}
	
	if len(docs) != 0 {
		t.Errorf("expected 0 documents after workspace deletion, got %d", len(docs))
	}
	
	// Verify chunks are deleted
	retrievedChunks, err := store.QueryChunks(ws.ID)
	if err != nil {
		t.Fatalf("failed to query chunks after workspace deletion: %v", err)
	}
	
	if len(retrievedChunks) != 0 {
		t.Errorf("expected 0 chunks after workspace deletion, got %d", len(retrievedChunks))
	}
	
	// Verify sessions are deleted
	sessions, err := store.ListSessions(ws.ID)
	if err != nil {
		t.Fatalf("failed to list sessions after workspace deletion: %v", err)
	}
	
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions after workspace deletion, got %d", len(sessions))
	}
}

// Test cascade delete for document
func TestDocumentCascadeDelete(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	
	// Create workspace
	ws, err := store.CreateWorkspace("test-workspace")
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}
	
	// Create document
	doc := &domain.Document{
		ID:          uuid.New().String(),
		WorkspaceID: ws.ID,
		Path:        "/path/to/test.pdf",
		Language:    domain.LanguageEN,
		Status:      domain.DocStatusPending,
		PageCount:   10,
		IsRead:      false,
		CreatedAt:   time.Now(),
	}
	
	err = store.WriteDocument(doc)
	if err != nil {
		t.Fatalf("failed to write document: %v", err)
	}
	
	// Create chunks
	chunks := []domain.Chunk{
		{
			ID:         uuid.New().String(),
			DocumentID: doc.ID,
			PageNumber: 1,
			Text:       "Test chunk 1",
			TokenCount: 10,
			Source:     domain.ChunkSourceTextLayer,
		},
	}
	
	err = store.WriteChunks(chunks)
	if err != nil {
		t.Fatalf("failed to write chunks: %v", err)
	}
	
	// Delete document (should cascade delete chunks)
	err = store.DeleteDocument(doc.ID)
	if err != nil {
		t.Fatalf("failed to delete document: %v", err)
	}
	
	// Verify chunks are deleted
	retrievedChunks, err := store.QueryChunks(ws.ID)
	if err != nil {
		t.Fatalf("failed to query chunks after document deletion: %v", err)
	}
	
	if len(retrievedChunks) != 0 {
		t.Errorf("expected 0 chunks after document deletion, got %d", len(retrievedChunks))
	}
}

// Test chunk operations
func TestChunkOperations(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	
	// Create workspace and document
	ws, err := store.CreateWorkspace("test-workspace")
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}
	
	doc := &domain.Document{
		ID:          uuid.New().String(),
		WorkspaceID: ws.ID,
		Path:        "/path/to/test.pdf",
		Language:    domain.LanguageEN,
		Status:      domain.DocStatusPending,
		PageCount:   10,
		IsRead:      false,
		CreatedAt:   time.Now(),
	}
	
	err = store.WriteDocument(doc)
	if err != nil {
		t.Fatalf("failed to write document: %v", err)
	}
	
	// Write chunks
	chunks := []domain.Chunk{
		{
			ID:         uuid.New().String(),
			DocumentID: doc.ID,
			PageNumber: 1,
			Text:       "This is a test chunk about machine learning",
			TokenCount: 10,
			Source:     domain.ChunkSourceTextLayer,
		},
		{
			ID:         uuid.New().String(),
			DocumentID: doc.ID,
			PageNumber: 2,
			Text:       "Another chunk discussing neural networks",
			TokenCount: 15,
			Source:     domain.ChunkSourceTextLayer,
		},
	}
	
	err = store.WriteChunks(chunks)
	if err != nil {
		t.Fatalf("failed to write chunks: %v", err)
	}
	
	// Query chunks
	retrievedChunks, err := store.QueryChunks(ws.ID)
	if err != nil {
		t.Fatalf("failed to query chunks: %v", err)
	}
	
	if len(retrievedChunks) != 2 {
		t.Errorf("expected 2 chunks, got %d", len(retrievedChunks))
	}
	
	// Delete chunks
	err = store.DeleteChunks(doc.ID)
	if err != nil {
		t.Fatalf("failed to delete chunks: %v", err)
	}
	
	// Verify deletion
	retrievedChunks, err = store.QueryChunks(ws.ID)
	if err != nil {
		t.Fatalf("failed to query chunks after deletion: %v", err)
	}
	
	if len(retrievedChunks) != 0 {
		t.Errorf("expected 0 chunks after deletion, got %d", len(retrievedChunks))
	}
}

// Test FTS5 keyword search
func TestKeywordSearch(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	
	// Create workspace and document
	ws, err := store.CreateWorkspace("test-workspace")
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}
	
	doc := &domain.Document{
		ID:          uuid.New().String(),
		WorkspaceID: ws.ID,
		Path:        "/path/to/test.pdf",
		Language:    domain.LanguageEN,
		Status:      domain.DocStatusIndexed,
		PageCount:   10,
		IsRead:      false,
		CreatedAt:   time.Now(),
	}
	
	err = store.WriteDocument(doc)
	if err != nil {
		t.Fatalf("failed to write document: %v", err)
	}
	
	// Write chunks with searchable content
	chunks := []domain.Chunk{
		{
			ID:         uuid.New().String(),
			DocumentID: doc.ID,
			PageNumber: 1,
			Text:       "Machine learning is a subset of artificial intelligence",
			TokenCount: 10,
			Source:     domain.ChunkSourceTextLayer,
		},
		{
			ID:         uuid.New().String(),
			DocumentID: doc.ID,
			PageNumber: 2,
			Text:       "Neural networks are used in deep learning applications",
			TokenCount: 15,
			Source:     domain.ChunkSourceTextLayer,
		},
		{
			ID:         uuid.New().String(),
			DocumentID: doc.ID,
			PageNumber: 3,
			Text:       "Natural language processing enables computers to understand text",
			TokenCount: 12,
			Source:     domain.ChunkSourceTextLayer,
		},
	}
	
	err = store.WriteChunks(chunks)
	if err != nil {
		t.Fatalf("failed to write chunks: %v", err)
	}
	
	// Search for "learning"
	results, err := store.SearchKeyword(ws.ID, "learning", 10)
	if err != nil {
		t.Fatalf("failed to search keywords: %v", err)
	}
	
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'learning', got %d", len(results))
	}
	
	// Search for "neural"
	results, err = store.SearchKeyword(ws.ID, "neural", 10)
	if err != nil {
		t.Fatalf("failed to search keywords: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'neural', got %d", len(results))
	}
	
	// Search for non-existent term
	results, err = store.SearchKeyword(ws.ID, "quantum", 10)
	if err != nil {
		t.Fatalf("failed to search keywords: %v", err)
	}
	
	if len(results) != 0 {
		t.Errorf("expected 0 results for 'quantum', got %d", len(results))
	}
}

// Test workspace isolation in keyword search
func TestKeywordSearchWorkspaceIsolation(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	
	// Create two workspaces
	ws1, err := store.CreateWorkspace("workspace-1")
	if err != nil {
		t.Fatalf("failed to create workspace 1: %v", err)
	}
	
	ws2, err := store.CreateWorkspace("workspace-2")
	if err != nil {
		t.Fatalf("failed to create workspace 2: %v", err)
	}
	
	// Create documents in each workspace
	doc1 := &domain.Document{
		ID:          uuid.New().String(),
		WorkspaceID: ws1.ID,
		Path:        "/path/to/doc1.pdf",
		Language:    domain.LanguageEN,
		Status:      domain.DocStatusIndexed,
		PageCount:   5,
		IsRead:      false,
		CreatedAt:   time.Now(),
	}
	
	doc2 := &domain.Document{
		ID:          uuid.New().String(),
		WorkspaceID: ws2.ID,
		Path:        "/path/to/doc2.pdf",
		Language:    domain.LanguageEN,
		Status:      domain.DocStatusIndexed,
		PageCount:   5,
		IsRead:      false,
		CreatedAt:   time.Now(),
	}
	
	err = store.WriteDocument(doc1)
	if err != nil {
		t.Fatalf("failed to write document 1: %v", err)
	}
	
	err = store.WriteDocument(doc2)
	if err != nil {
		t.Fatalf("failed to write document 2: %v", err)
	}
	
	// Write chunks with unique content per workspace
	chunks1 := []domain.Chunk{
		{
			ID:         uuid.New().String(),
			DocumentID: doc1.ID,
			PageNumber: 1,
			Text:       "Workspace one contains information about Python programming",
			TokenCount: 10,
			Source:     domain.ChunkSourceTextLayer,
		},
	}
	
	chunks2 := []domain.Chunk{
		{
			ID:         uuid.New().String(),
			DocumentID: doc2.ID,
			PageNumber: 1,
			Text:       "Workspace two contains information about Go programming",
			TokenCount: 10,
			Source:     domain.ChunkSourceTextLayer,
		},
	}
	
	err = store.WriteChunks(chunks1)
	if err != nil {
		t.Fatalf("failed to write chunks 1: %v", err)
	}
	
	err = store.WriteChunks(chunks2)
	if err != nil {
		t.Fatalf("failed to write chunks 2: %v", err)
	}
	
	// Search in workspace 1 for "Python"
	results, err := store.SearchKeyword(ws1.ID, "Python", 10)
	if err != nil {
		t.Fatalf("failed to search in workspace 1: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("expected 1 result in workspace 1, got %d", len(results))
	}
	
	// Search in workspace 2 for "Python" (should return 0)
	results, err = store.SearchKeyword(ws2.ID, "Python", 10)
	if err != nil {
		t.Fatalf("failed to search in workspace 2: %v", err)
	}
	
	if len(results) != 0 {
		t.Errorf("expected 0 results in workspace 2 for 'Python', got %d", len(results))
	}
	
	// Search in workspace 2 for "Go"
	results, err = store.SearchKeyword(ws2.ID, "Go", 10)
	if err != nil {
		t.Fatalf("failed to search in workspace 2: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("expected 1 result in workspace 2, got %d", len(results))
	}
}

// Test session operations
func TestSessionOperations(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	
	// Create workspace
	ws, err := store.CreateWorkspace("test-workspace")
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}
	
	// Create session
	session := &domain.Session{
		ID:           uuid.New().String(),
		WorkspaceID:  ws.ID,
		StartedAt:    time.Now(),
		LastActiveAt: time.Now(),
		Title:        "Test session",
	}
	
	err = store.CreateSession(session)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	
	// List sessions
	sessions, err := store.ListSessions(ws.ID)
	if err != nil {
		t.Fatalf("failed to list sessions: %v", err)
	}
	
	if len(sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(sessions))
	}
	
	// Get session
	retrieved, err := store.GetSession(session.ID)
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	
	if retrieved.ID != session.ID {
		t.Errorf("expected session ID %s, got %s", session.ID, retrieved.ID)
	}
	
	// Append turn
	turn := &domain.Turn{
		ID:        uuid.New().String(),
		SessionID: session.ID,
		Question:  "What is machine learning?",
		Answer: domain.Answer{
			Text:       "Machine learning is a subset of AI",
			Confidence: domain.ConfidenceHigh,
			Sources: []domain.Citation{
				{
					ChunkID:    uuid.New().String(),
					DocName:    "test.pdf",
					PageNumber: 1,
					Excerpt:    "Machine learning...",
				},
			},
			SessionID: session.ID,
		},
		CreatedAt: time.Now(),
	}
	
	err = store.AppendTurn(turn)
	if err != nil {
		t.Fatalf("failed to append turn: %v", err)
	}
	
	// Get turns
	turns, err := store.GetTurns(session.ID)
	if err != nil {
		t.Fatalf("failed to get turns: %v", err)
	}
	
	if len(turns) != 1 {
		t.Errorf("expected 1 turn, got %d", len(turns))
	}
	
	if turns[0].Question != turn.Question {
		t.Errorf("expected question %s, got %s", turn.Question, turns[0].Question)
	}
	
	if len(turns[0].Answer.Sources) != 1 {
		t.Errorf("expected 1 citation, got %d", len(turns[0].Answer.Sources))
	}
}

// Test transaction control
func TestTransactionControl(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	
	// Begin transaction
	tx, err := store.BeginTx()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	
	// Commit transaction
	err = tx.Commit()
	if err != nil {
		t.Fatalf("failed to commit transaction: %v", err)
	}
	
	// Begin another transaction
	tx, err = store.BeginTx()
	if err != nil {
		t.Fatalf("failed to begin second transaction: %v", err)
	}
	
	// Rollback transaction
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("failed to rollback transaction: %v", err)
	}
}

// Test referential integrity - document with invalid workspace_id
func TestReferentialIntegrityDocument(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	
	// Attempt to create document with non-existent workspace_id
	doc := &domain.Document{
		ID:          uuid.New().String(),
		WorkspaceID: "non-existent-workspace-id",
		Path:        "/path/to/test.pdf",
		Language:    domain.LanguageEN,
		Status:      domain.DocStatusPending,
		PageCount:   10,
		IsRead:      false,
		CreatedAt:   time.Now(),
	}
	
	err := store.WriteDocument(doc)
	if err == nil {
		t.Error("expected error when creating document with invalid workspace_id, got nil")
	}
}

// Test referential integrity - chunk with invalid document_id
func TestReferentialIntegrityChunk(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	
	// Attempt to create chunk with non-existent document_id
	chunks := []domain.Chunk{
		{
			ID:         uuid.New().String(),
			DocumentID: "non-existent-document-id",
			PageNumber: 1,
			Text:       "Test chunk",
			TokenCount: 10,
			Source:     domain.ChunkSourceTextLayer,
		},
	}
	
	err := store.WriteChunks(chunks)
	if err == nil {
		t.Error("expected error when creating chunk with invalid document_id, got nil")
	}
}

// Test referential integrity - session with invalid workspace_id
func TestReferentialIntegritySession(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	
	// Attempt to create session with non-existent workspace_id
	session := &domain.Session{
		ID:           uuid.New().String(),
		WorkspaceID:  "non-existent-workspace-id",
		StartedAt:    time.Now(),
		LastActiveAt: time.Now(),
		Title:        "Test session",
	}
	
	err := store.CreateSession(session)
	if err == nil {
		t.Error("expected error when creating session with invalid workspace_id, got nil")
	}
}
