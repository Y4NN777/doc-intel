//go:build fts5

package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Y4NN777/doc-intel/internal/domain"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
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
	SearchKeyword(workspaceID, query string, limit int) ([]domain.ScoredChunk, error)

	// Session operations
	CreateSession(session *domain.Session) error
	ListSessions(workspaceID string) ([]*domain.Session, error)
	GetSession(id string) (*domain.Session, error)
	AppendTurn(turn *domain.Turn) error
	GetTurns(sessionID string) ([]*domain.Turn, error)

	// Transaction control
	BeginTx() (Tx, error)
	Close() error
}

// Tx represents a database transaction
type Tx interface {
	Commit() error
	Rollback() error
}

// SQLiteStore implements Store using SQLite
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite-backed Store
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// initSchema creates all tables, indexes, and constraints
func (s *SQLiteStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS workspaces (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP NOT NULL,
		last_used_at TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS documents (
		id TEXT PRIMARY KEY,
		workspace_id TEXT NOT NULL,
		path TEXT NOT NULL,
		language TEXT NOT NULL,
		status TEXT NOT NULL,
		page_count INTEGER NOT NULL,
		is_read BOOLEAN NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL,
		processed_at TIMESTAMP,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS chunks (
		id TEXT PRIMARY KEY,
		document_id TEXT NOT NULL,
		page_number INTEGER NOT NULL,
		text TEXT NOT NULL,
		token_count INTEGER NOT NULL,
		source TEXT NOT NULL,
		FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		workspace_id TEXT NOT NULL,
		started_at TIMESTAMP NOT NULL,
		last_active_at TIMESTAMP NOT NULL,
		title TEXT NOT NULL,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS turns (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		question TEXT NOT NULL,
		answer_text TEXT NOT NULL,
		answer_confidence TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS citations (
		id TEXT PRIMARY KEY,
		turn_id TEXT NOT NULL,
		chunk_id TEXT NOT NULL,
		doc_name TEXT NOT NULL,
		page_number INTEGER NOT NULL,
		excerpt TEXT NOT NULL,
		FOREIGN KEY (turn_id) REFERENCES turns(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_documents_workspace ON documents(workspace_id);
	CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);
	CREATE INDEX IF NOT EXISTS idx_chunks_document ON chunks(document_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_workspace ON sessions(workspace_id);
	CREATE INDEX IF NOT EXISTS idx_turns_session ON turns(session_id);

	-- Full-text search index for keyword retrieval
	CREATE VIRTUAL TABLE IF NOT EXISTS chunks_fts USING fts5(
		chunk_id UNINDEXED,
		text
	);

	-- Triggers to keep FTS5 index synchronized with chunks table
	CREATE TRIGGER IF NOT EXISTS chunks_fts_insert AFTER INSERT ON chunks BEGIN
		INSERT INTO chunks_fts(chunk_id, text) VALUES (new.id, new.text);
	END;

	CREATE TRIGGER IF NOT EXISTS chunks_fts_delete AFTER DELETE ON chunks BEGIN
		DELETE FROM chunks_fts WHERE chunk_id = old.id;
	END;

	CREATE TRIGGER IF NOT EXISTS chunks_fts_update AFTER UPDATE ON chunks BEGIN
		UPDATE chunks_fts SET text = new.text WHERE chunk_id = new.id;
	END;
	`

	_, err := s.db.Exec(schema)
	return err
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// BeginTx starts a new transaction
func (s *SQLiteStore) BeginTx() (Tx, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	return &sqliteTx{tx: tx}, nil
}

type sqliteTx struct {
	tx *sql.Tx
}

func (t *sqliteTx) Commit() error {
	return t.tx.Commit()
}

func (t *sqliteTx) Rollback() error {
	return t.tx.Rollback()
}

// Workspace operations

func (s *SQLiteStore) CreateWorkspace(name string) (*domain.Workspace, error) {
	ws := &domain.Workspace{
		ID:         uuid.New().String(),
		Name:       name,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}

	_, err := s.db.Exec(
		"INSERT INTO workspaces (id, name, created_at, last_used_at) VALUES (?, ?, ?, ?)",
		ws.ID, ws.Name, ws.CreatedAt, ws.LastUsedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	return ws, nil
}

func (s *SQLiteStore) ListWorkspaces() ([]domain.Workspace, error) {
	rows, err := s.db.Query("SELECT id, name, created_at, last_used_at FROM workspaces ORDER BY last_used_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []domain.Workspace
	for rows.Next() {
		var ws domain.Workspace
		if err := rows.Scan(&ws.ID, &ws.Name, &ws.CreatedAt, &ws.LastUsedAt); err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, ws)
	}

	return workspaces, rows.Err()
}

func (s *SQLiteStore) GetWorkspace(id string) (*domain.Workspace, error) {
	var ws domain.Workspace
	err := s.db.QueryRow(
		"SELECT id, name, created_at, last_used_at FROM workspaces WHERE id = ?",
		id,
	).Scan(&ws.ID, &ws.Name, &ws.CreatedAt, &ws.LastUsedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("workspace not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	return &ws, nil
}

func (s *SQLiteStore) DeleteWorkspace(id string) error {
	result, err := s.db.Exec("DELETE FROM workspaces WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("workspace not found: %s", id)
	}

	return nil
}

// Document operations

func (s *SQLiteStore) WriteDocument(doc *domain.Document) error {
	_, err := s.db.Exec(
		`INSERT INTO documents (id, workspace_id, path, language, status, page_count, is_read, created_at, processed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		doc.ID, doc.WorkspaceID, doc.Path, doc.Language, doc.Status,
		doc.PageCount, doc.IsRead, doc.CreatedAt, doc.ProcessedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to write document: %w", err)
	}
	return nil
}

func (s *SQLiteStore) ListDocuments(workspaceID string) ([]domain.Document, error) {
	rows, err := s.db.Query(
		`SELECT id, workspace_id, path, language, status, page_count, is_read, created_at, processed_at
		 FROM documents WHERE workspace_id = ? ORDER BY created_at DESC`,
		workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var documents []domain.Document
	for rows.Next() {
		var doc domain.Document
		if err := rows.Scan(
			&doc.ID, &doc.WorkspaceID, &doc.Path, &doc.Language, &doc.Status,
			&doc.PageCount, &doc.IsRead, &doc.CreatedAt, &doc.ProcessedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, doc)
	}

	return documents, rows.Err()
}

func (s *SQLiteStore) GetDocument(id string) (*domain.Document, error) {
	var doc domain.Document
	err := s.db.QueryRow(
		`SELECT id, workspace_id, path, language, status, page_count, is_read, created_at, processed_at
		 FROM documents WHERE id = ?`,
		id,
	).Scan(
		&doc.ID, &doc.WorkspaceID, &doc.Path, &doc.Language, &doc.Status,
		&doc.PageCount, &doc.IsRead, &doc.CreatedAt, &doc.ProcessedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("document not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return &doc, nil
}

func (s *SQLiteStore) DeleteDocument(id string) error {
	result, err := s.db.Exec("DELETE FROM documents WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("document not found: %s", id)
	}

	return nil
}

func (s *SQLiteStore) SetDocumentStatus(id string, status domain.DocStatus) error {
	var processedAt *time.Time
	if status == domain.DocStatusIndexed {
		now := time.Now()
		processedAt = &now
	}

	result, err := s.db.Exec(
		"UPDATE documents SET status = ?, processed_at = ? WHERE id = ?",
		status, processedAt, id,
	)
	if err != nil {
		return fmt.Errorf("failed to set document status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("document not found: %s", id)
	}

	return nil
}

func (s *SQLiteStore) MarkDocumentRead(id string, isRead bool) error {
	result, err := s.db.Exec("UPDATE documents SET is_read = ? WHERE id = ?", isRead, id)
	if err != nil {
		return fmt.Errorf("failed to mark document read: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("document not found: %s", id)
	}

	return nil
}

// Chunk operations

func (s *SQLiteStore) WriteChunks(chunks []domain.Chunk) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		"INSERT INTO chunks (id, document_id, page_number, text, token_count, source) VALUES (?, ?, ?, ?, ?, ?)",
	)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, chunk := range chunks {
		if _, err := stmt.Exec(
			chunk.ID, chunk.DocumentID, chunk.PageNumber,
			chunk.Text, chunk.TokenCount, chunk.Source,
		); err != nil {
			return fmt.Errorf("failed to write chunk: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *SQLiteStore) QueryChunks(workspaceID string) ([]domain.Chunk, error) {
	rows, err := s.db.Query(
		`SELECT c.id, c.document_id, c.page_number, c.text, c.token_count, c.source
		 FROM chunks c
		 JOIN documents d ON c.document_id = d.id
		 WHERE d.workspace_id = ?`,
		workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query chunks: %w", err)
	}
	defer rows.Close()

	var chunks []domain.Chunk
	for rows.Next() {
		var chunk domain.Chunk
		if err := rows.Scan(
			&chunk.ID, &chunk.DocumentID, &chunk.PageNumber,
			&chunk.Text, &chunk.TokenCount, &chunk.Source,
		); err != nil {
			return nil, fmt.Errorf("failed to scan chunk: %w", err)
		}
		chunks = append(chunks, chunk)
	}

	return chunks, rows.Err()
}

func (s *SQLiteStore) DeleteChunks(documentID string) error {
	_, err := s.db.Exec("DELETE FROM chunks WHERE document_id = ?", documentID)
	if err != nil {
		return fmt.Errorf("failed to delete chunks: %w", err)
	}
	return nil
}

func (s *SQLiteStore) SearchKeyword(workspaceID, query string, limit int) ([]domain.ScoredChunk, error) {
	// First, get matching chunk IDs and scores from FTS5
	rows, err := s.db.Query(
		`SELECT chunk_id, bm25(chunks_fts) as score
		 FROM chunks_fts
		 WHERE chunks_fts MATCH ?
		 ORDER BY score
		 LIMIT ?`,
		query, limit*10, // Get more results to filter by workspace
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search FTS5: %w", err)
	}
	
	type ftsResult struct {
		chunkID string
		score   float64
	}
	var ftsResults []ftsResult
	for rows.Next() {
		var r ftsResult
		if err := rows.Scan(&r.chunkID, &r.score); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan FTS result: %w", err)
		}
		ftsResults = append(ftsResults, r)
	}
	rows.Close()
	
	if len(ftsResults) == 0 {
		return []domain.ScoredChunk{}, nil
	}
	
	// Build chunk ID list for IN clause
	chunkIDs := make([]string, len(ftsResults))
	scoreMap := make(map[string]float64)
	for i, r := range ftsResults {
		chunkIDs[i] = r.chunkID
		scoreMap[r.chunkID] = r.score
	}
	
	// Build placeholders for IN clause
	placeholders := make([]string, len(chunkIDs))
	args := make([]interface{}, 0, len(chunkIDs)+2)
	args = append(args, workspaceID)
	for i, id := range chunkIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	args = append(args, limit)
	
	// Get full chunk data filtered by workspace
	query2 := fmt.Sprintf(
		`SELECT c.id, c.document_id, c.page_number, c.text, c.token_count, c.source
		 FROM chunks c
		 JOIN documents d ON c.document_id = d.id
		 WHERE d.workspace_id = ? AND c.id IN (%s)
		 LIMIT ?`,
		strings.Join(placeholders, ","),
	)
	
	rows, err = s.db.Query(query2, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query chunks: %w", err)
	}
	defer rows.Close()

	var results []domain.ScoredChunk
	for rows.Next() {
		var chunk domain.Chunk
		if err := rows.Scan(
			&chunk.ID, &chunk.DocumentID, &chunk.PageNumber,
			&chunk.Text, &chunk.TokenCount, &chunk.Source,
		); err != nil {
			return nil, fmt.Errorf("failed to scan chunk: %w", err)
		}
		// BM25 score is negative (lower is better), convert to positive (higher is better)
		bm25Score := scoreMap[chunk.ID]
		normalizedScore := 1.0 / (1.0 - bm25Score)
		results = append(results, domain.ScoredChunk{
			Chunk: chunk,
			Score: normalizedScore,
		})
	}

	return results, rows.Err()
}

// Session operations

func (s *SQLiteStore) CreateSession(session *domain.Session) error {
	_, err := s.db.Exec(
		"INSERT INTO sessions (id, workspace_id, started_at, last_active_at, title) VALUES (?, ?, ?, ?, ?)",
		session.ID, session.WorkspaceID, session.StartedAt, session.LastActiveAt, session.Title,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (s *SQLiteStore) ListSessions(workspaceID string) ([]*domain.Session, error) {
	rows, err := s.db.Query(
		"SELECT id, workspace_id, started_at, last_active_at, title FROM sessions WHERE workspace_id = ? ORDER BY last_active_at DESC",
		workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		var session domain.Session
		if err := rows.Scan(&session.ID, &session.WorkspaceID, &session.StartedAt, &session.LastActiveAt, &session.Title); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, &session)
	}

	return sessions, rows.Err()
}

func (s *SQLiteStore) GetSession(id string) (*domain.Session, error) {
	var session domain.Session
	err := s.db.QueryRow(
		"SELECT id, workspace_id, started_at, last_active_at, title FROM sessions WHERE id = ?",
		id,
	).Scan(&session.ID, &session.WorkspaceID, &session.StartedAt, &session.LastActiveAt, &session.Title)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

func (s *SQLiteStore) AppendTurn(turn *domain.Turn) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert turn
	_, err = tx.Exec(
		"INSERT INTO turns (id, session_id, question, answer_text, answer_confidence, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		turn.ID, turn.SessionID, turn.Question, turn.Answer.Text, turn.Answer.Confidence, turn.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to append turn: %w", err)
	}

	// Insert citations
	for _, citation := range turn.Answer.Sources {
		citationID := uuid.New().String()
		_, err = tx.Exec(
			"INSERT INTO citations (id, turn_id, chunk_id, doc_name, page_number, excerpt) VALUES (?, ?, ?, ?, ?, ?)",
			citationID, turn.ID, citation.ChunkID, citation.DocName, citation.PageNumber, citation.Excerpt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert citation: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *SQLiteStore) GetTurns(sessionID string) ([]*domain.Turn, error) {
	rows, err := s.db.Query(
		"SELECT id, session_id, question, answer_text, answer_confidence, created_at FROM turns WHERE session_id = ? ORDER BY created_at ASC",
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get turns: %w", err)
	}
	defer rows.Close()

	var turns []*domain.Turn
	for rows.Next() {
		var turn domain.Turn
		var answerText string
		var answerConfidence domain.Confidence

		if err := rows.Scan(&turn.ID, &turn.SessionID, &turn.Question, &answerText, &answerConfidence, &turn.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan turn: %w", err)
		}

		// Load citations for this turn
		citations, err := s.getCitations(turn.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get citations: %w", err)
		}

		turn.Answer = domain.Answer{
			Text:       answerText,
			Confidence: answerConfidence,
			Sources:    citations,
			SessionID:  sessionID,
		}

		turns = append(turns, &turn)
	}

	return turns, rows.Err()
}

func (s *SQLiteStore) getCitations(turnID string) ([]domain.Citation, error) {
	rows, err := s.db.Query(
		"SELECT chunk_id, doc_name, page_number, excerpt FROM citations WHERE turn_id = ?",
		turnID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query citations: %w", err)
	}
	defer rows.Close()

	var citations []domain.Citation
	for rows.Next() {
		var citation domain.Citation
		if err := rows.Scan(&citation.ChunkID, &citation.DocName, &citation.PageNumber, &citation.Excerpt); err != nil {
			return nil, fmt.Errorf("failed to scan citation: %w", err)
		}
		citations = append(citations, citation)
	}

	return citations, rows.Err()
}
