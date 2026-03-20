# Design Document: Doc-Intel Implementation

## Overview

Doc-Intel is a terminal-native document intelligence system that enables developers to organize private PDF documents into workspace-scoped collections, process them locally, and ask natural language questions with grounded answers. The system enforces zero network calls, workspace isolation, and atomic operations while supporting English and French documents and queries.

This design document details the implementation approach for Doc-Intel based on the comprehensive engineering documentation. The system follows a Layered Pipeline Architecture with an embedded Agent Loop in the query layer, compiled into a single Go binary with no runtime dependencies.

### Design Goals

1. **Privacy-First**: Absolute guarantee that no document content leaves the local machine
2. **Workspace Isolation**: Complete separation of document collections with no cross-contamination
3. **Atomic Operations**: All state transitions are transactional - no partial states at rest
4. **Grounded Answers**: Every response is derived from and cited to actual document content
5. **Developer Experience**: Fast, terminal-native interface with streaming responses
6. **Bilingual Support**: Seamless handling of English and French documents and queries

### Key Constraints

- Single compiled binary (no runtime dependencies)
- Target hardware: Ryzen 3, 16GB RAM, Ubuntu 22.04+
- Zero outbound network calls (enforced at process start)
- Manual ingestion only (no background processes)
- TUI-only interface in v1

## Architecture

### Architectural Style

Doc-Intel follows a **Layered Pipeline Architecture** with an embedded **Agent Loop** in the query layer. This choice is justified by three observations:

1. **Data flows in one direction**: Documents enter the system, get processed through sequential stages (extract → chunk → embed → store), and never flow backwards
2. **Concerns are strictly separated by depth**: The user interface knows nothing about storage; storage knows nothing about reasoning; reasoning knows nothing about rendering
3. **The query path is non-linear**: Unlike ingestion, a query does not follow a fixed sequence - the system must decide how many retrieval passes to make

### System Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    USER INTERFACE LAYER                     │
│                          (TUI)                              │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────┴────────────────────────────────────┐
│                   ORCHESTRATION LAYER                       │
│         WorkspaceManager | DocumentManager                  │
└────────────┬────────────────────────┬───────────────────────┘
             │                        │
    ┌────────┴────────┐      ┌───────┴────────┐
    │  INGESTION      │      │  QUERY         │
    │  Pipeline       │      │  Orchestrator  │
    │  Parser         │      │  Agent Loop    │
    │  Chunker        │      │  Retriever     │
    │  Embedder       │      │  SessionMgr    │
    └────────┬────────┘      └───────┬────────┘
             │                        │
┌────────────┴────────────────────────┴───────────────────────┐
│                   PERSISTENCE LAYER                         │
│         Store (SQLite) | VectorIndex (FAISS)                │
└────────────┬────────────────────────┬───────────────────────┘
             │                        │
┌────────────┴────────────────────────┴───────────────────────┐
│                 EXTERNAL RUNTIME LAYER                      │
│      Local LLM | Embedding Model | File System             │
└─────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

**Layer 1 - User Interface (TUI)**
- Only entry point for user interaction
- Renders output, accepts input, streams answers
- Dispatches commands to orchestration layer
- Handles progress display and error presentation

**Layer 2 - Orchestration**
- WorkspaceManager: workspace lifecycle (create, list, switch, delete)
- DocumentManager: document lifecycle (add, list, delete, mark read)
- Routes commands to correct pipeline or query path
- Enforces cascade correctness on deletions

**Layer 3 - Ingestion (left path)**
- Pipeline: coordinates Parser → Chunker → Embedder sequence
- Parser: extracts text from PDFs (text layer or OCR fallback)
- Chunker: splits text into semantic units (≤512 tokens, 64 overlap)
- Embedder: generates vector embeddings using local model
- Transaction boundary: all-or-nothing commit

**Layer 3 - Query (right path)**
- QueryOrchestrator: agent loop (plan → retrieve → generate → cite → score)
- Retriever: hybrid search (semantic + keyword, workspace-scoped)
- SessionManager: conversation history (append-only)
- Bounded to 5 retrieval passes maximum

**Layer 4 - Persistence**
- Store (SQLite): structured data with transaction control
- VectorIndex (FAISS): ANN search over embeddings
- Enforces referential integrity and atomic operations

**Layer 5 - External Runtime**
- Local LLM binary (text generation)
- Embedding Model binary (vector generation)
- File System (read-only PDF access)
- All communication via local IPC only

### Data Flows

**Flow A - Ingestion**
```
User command → TUI → DocumentManager → Pipeline
  → Parser → Chunker → Embedder
  → Store + VectorIndex (atomic commit)
  → Status report streamed to TUI
```

**Flow B - Query**
```
User question → TUI → QueryOrchestrator
  → Agent Loop [Retriever ↔ Store/VectorIndex] (up to 5x)
  → LLM (generate)
  → CitationExtractor + ConfidenceScorer
  → Answer streamed to TUI
  → Turn persisted to SessionManager → Store
```

## Components and Interfaces

### Core Components

#### WorkspaceManager

**Responsibility**: Manages workspace lifecycle operations

**Interface**:
```go
type WorkspaceManager interface {
    Create(name string) (*domain.Workspace, error)
    List() ([]*domain.Workspace, error)
    Switch(name string) error
    Delete(name string) error
    GetActive() (*domain.Workspace, error)
}
```

**Dependencies**: Store

**Key Behaviors**:
- Create validates name uniqueness before persisting
- Delete delegates to Store for atomic cascade removal
- Switch updates active workspace in memory
- All operations are synchronous and return errors immediately

#### DocumentManager

**Responsibility**: Manages document lifecycle operations

**Interface**:
```go
type DocumentManager interface {
    Add(workspaceID, path string) (*domain.Document, error)
    List(workspaceID string) ([]*domain.Document, error)
    Delete(documentID string) error
    MarkRead(documentID string, isRead bool) error
    GetByID(documentID string) (*domain.Document, error)
}
```

**Dependencies**: Store

**Key Behaviors**:
- Add creates document record with status "pending"
- Delete delegates to Store for atomic removal of document + chunks + embeddings
- MarkRead updates read flag only
- All operations scoped by workspace

#### Pipeline

**Responsibility**: Orchestrates the full ingestion sequence with transaction boundary

**Interface**:
```go
type Pipeline interface {
    Ingest(workspaceID, documentID string) error
    Reprocess(documentID string) error
}
```

**Dependencies**: Parser, Chunker, Embedder, Store, VectorIndex

**Key Behaviors**:
- Ingest coordinates Parser → Chunker → Embedder in sequence
- Wraps entire sequence in Store transaction
- On success: commits all chunks + embeddings, sets status "indexed"
- On failure: rolls back all changes, sets status "failed"
- Reprocess purges existing data before re-ingesting
- Reports progress to TUI during processing

#### Parser

**Responsibility**: Extracts text from PDF files

**Interface**:
```go
type Parser interface {
    Extract(path string) ([]PageText, Language, error)
}

type PageText struct {
    PageNumber int
    Text       string
    Source     domain.ChunkSource
}
```

**Dependencies**: File System, OCR library

**Key Behaviors**:
- Attempts text layer extraction first
- Falls back to OCR on pages with no text layer
- Detects primary language (English, French, or unknown)
- Returns page-by-page text with source annotation
- Handles corrupt/password-protected PDFs with descriptive errors

#### Chunker

**Responsibility**: Splits text into semantic units with overlap

**Interface**:
```go
type Chunker interface {
    Chunk(pages []PageText, documentID string) ([]domain.Chunk, error)
}
```

**Dependencies**: Tokenizer

**Key Behaviors**:
- Splits text into chunks where no chunk exceeds 512 tokens
- Ensures minimum 64 token overlap between consecutive chunks
- Preserves page number and source metadata
- Respects sentence boundaries when possible
- Generates unique chunk IDs

#### Embedder

**Responsibility**: Generates vector embeddings from text

**Interface**:
```go
type Embedder interface {
    Embed(chunks []domain.Chunk) ([]domain.Vector, error)
}
```

**Dependencies**: Embedding Model binary

**Key Behaviors**:
- Calls local embedding model via IPC
- Generates one vector per chunk
- Supports English and French text
- Returns vectors with dimensions metadata
- Handles embedding failures gracefully

#### Retriever

**Responsibility**: Hybrid search combining semantic and keyword retrieval

**Interface**:
```go
type Retriever interface {
    Search(req SearchRequest) ([]domain.ScoredChunk, error)
}

type SearchRequest struct {
    WorkspaceID string
    DocumentID  *string // optional
    Query       string
    TopK        int
}
```

**Dependencies**: Store, VectorIndex

**Key Behaviors**:
- Requires explicit workspace_id (no default)
- Performs semantic search via VectorIndex
- Performs keyword search via Store
- Merges and ranks results by relevance score
- Optional document_id further scopes search
- Returns chunks with text, metadata, page number, and score

#### QueryOrchestrator

**Responsibility**: Agent loop for query planning, retrieval, generation, citation, and scoring

**Interface**:
```go
type QueryOrchestrator interface {
    Query(req domain.QueryRequest) (*domain.Answer, error)
    Summarize(workspaceID, documentID string) (*domain.Answer, error)
    Extract(workspaceID, query, extractType string) (*domain.Answer, error)
    Compare(workspaceID, query string) (*domain.Answer, error)
}
```

**Dependencies**: Retriever, SessionManager, Store, LLM

**Key Behaviors**:
- Implements agent loop: plan → retrieve → evaluate → generate → cite → score
- Bounded to maximum 5 retrieval passes
- Loads session context for multi-turn conversations
- Constructs prompts containing only retrieved chunks
- Calls LLM for text generation
- Extracts citations from retrieved chunks only (never generates)
- Scores confidence based on retrieval quality
- Returns Answer with text, citations, and confidence
- Streams tokens progressively to TUI

#### SessionManager

**Responsibility**: Manages conversation history (append-only)

**Interface**:
```go
type SessionManager interface {
    Create(workspaceID string) (*domain.Session, error)
    List(workspaceID string) ([]*domain.Session, error)
    Load(sessionID string) ([]*domain.Turn, error)
    AppendTurn(sessionID string, question string, answer domain.Answer) error
    GetActive() (*domain.Session, error)
}
```

**Dependencies**: Store

**Key Behaviors**:
- Create generates new session scoped to workspace
- AppendTurn writes question + answer as immutable Turn
- No update path exists (append-only)
- Load retrieves all turns for session
- List returns sessions scoped to workspace
- All turns deleted atomically when workspace deleted

#### Store

**Responsibility**: SQLite-based persistence for structured data with transaction control

**Interface**:
```go
type Store interface {
    // Workspace operations
    CreateWorkspace(ws *domain.Workspace) error
    ListWorkspaces() ([]*domain.Workspace, error)
    DeleteWorkspace(id string) error
    
    // Document operations
    CreateDocument(doc *domain.Document) error
    ListDocuments(workspaceID string) ([]*domain.Document, error)
    GetDocument(id string) (*domain.Document, error)
    UpdateDocumentStatus(id string, status domain.DocStatus) error
    DeleteDocument(id string) error
    
    // Chunk operations
    CreateChunks(chunks []domain.Chunk) error
    GetChunks(documentID string) ([]domain.Chunk, error)
    GetChunksByWorkspace(workspaceID string) ([]domain.Chunk, error)
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
}

type Tx interface {
    Commit() error
    Rollback() error
}
```

**Dependencies**: SQLite library

**Key Behaviors**:
- Enforces referential integrity via foreign keys
- Provides transaction control for atomic operations
- Implements cascade delete for workspace and document removal
- Keyword search uses full-text search on chunk text
- All queries scoped by workspace_id
- Single writer pattern (no concurrent writes)

#### VectorIndex

**Responsibility**: FAISS-based ANN search over chunk embeddings

**Interface**:
```go
type VectorIndex interface {
    Insert(workspaceID string, vectors []domain.Vector) error
    Search(workspaceID string, queryVector []float32, topK int) ([]string, []float64, error)
    Delete(workspaceID string, chunkIDs []string) error
    DeleteWorkspace(workspaceID string) error
}
```

**Dependencies**: FAISS library

**Key Behaviors**:
- Namespaces embeddings by workspace_id
- Performs approximate nearest neighbor search
- Returns chunk IDs and similarity scores
- Insert and Delete are eventually consistent
- No cross-workspace search path exists
- Workspace deletion removes all associated embeddings

## Data Models

### Domain Entities

#### Workspace
```go
type Workspace struct {
    ID         string    // UUID
    Name       string    // unique, user-provided
    CreatedAt  time.Time
    LastUsedAt time.Time
}
```

**Invariants**:
- Name must be unique across all workspaces
- ID is immutable after creation
- LastUsedAt updated on switch or query

#### Document
```go
type Document struct {
    ID          string    // UUID
    WorkspaceID string    // foreign key to Workspace
    Path        string    // absolute file path
    Language    Language  // en, fr, unknown
    Status      DocStatus // pending, indexed, failed
    PageCount   int
    IsRead      bool
    CreatedAt   time.Time
    ProcessedAt *time.Time // nil until indexed
}

type DocStatus string
const (
    DocStatusPending DocStatus = "pending"
    DocStatusIndexed DocStatus = "indexed"
    DocStatusFailed  DocStatus = "failed"
)

type Language string
const (
    LanguageEN      Language = "en"
    LanguageFR      Language = "fr"
    LanguageUnknown Language = "unknown"
)
```

**Invariants**:
- WorkspaceID must reference existing workspace
- Status transitions: pending → indexed OR pending → failed
- ProcessedAt set only when status becomes indexed
- Path must be valid and accessible at Add time

#### Chunk
```go
type Chunk struct {
    ID         string      // UUID
    DocumentID string      // foreign key to Document
    PageNumber int         // 1-indexed
    Text       string      // extracted text
    TokenCount int         // ≤512
    Source     ChunkSource // text_layer or ocr
}

type ChunkSource string
const (
    ChunkSourceTextLayer ChunkSource = "text_layer"
    ChunkSourceOCR       ChunkSource = "ocr"
)
```

**Invariants**:
- DocumentID must reference existing document
- TokenCount ≤ 512
- PageNumber ≥ 1
- Text non-empty (except for OCR failures)
- Consecutive chunks overlap by ≥64 tokens

#### Vector
```go
type Vector struct {
    ChunkID    string    // foreign key to Chunk
    Values     []float32 // embedding vector
    Dimensions int       // vector dimensionality
}
```

**Invariants**:
- ChunkID must reference existing chunk
- len(Values) == Dimensions
- Dimensions determined by embedding model

#### Session
```go
type Session struct {
    ID           string    // UUID
    WorkspaceID  string    // foreign key to Workspace
    StartedAt    time.Time
    LastActiveAt time.Time
    Title        string    // auto-generated from first question
}
```

**Invariants**:
- WorkspaceID must reference existing workspace
- LastActiveAt ≥ StartedAt
- Title generated from first turn's question

#### Turn
```go
type Turn struct {
    ID        string    // UUID
    SessionID string    // foreign key to Session
    Question  string    // user query
    Answer    Answer    // system response
    CreatedAt time.Time
}
```

**Invariants**:
- SessionID must reference existing session
- Question non-empty
- Answer must have at least one citation (except for errors)
- Turns are immutable after creation

#### QueryRequest
```go
type QueryRequest struct {
    WorkspaceID string
    DocumentID  *string    // optional - scopes to single document
    Text        string     // user query
    SessionID   *string    // optional - for conversation context
    Type        QueryType  // question, summary, extraction, comparison
}

type QueryType string
const (
    QueryTypeQuestion   QueryType = "question"
    QueryTypeSummary    QueryType = "summary"
    QueryTypeExtraction QueryType = "extraction"
    QueryTypeComparison QueryType = "comparison"
)
```

#### Answer
```go
type Answer struct {
    Text       string
    Confidence Confidence
    Sources    []Citation
    SessionID  string
}

type Confidence string
const (
    ConfidenceHigh   Confidence = "high"
    ConfidenceMedium Confidence = "medium"
    ConfidenceLow    Confidence = "low"
)
```

**Invariants**:
- Sources must contain only chunks retrieved for this query
- Confidence based on retrieval scores and answer grounding
- Text non-empty

#### Citation
```go
type Citation struct {
    ChunkID    string // references retrieved chunk
    DocName    string // document filename
    PageNumber int    // 1-indexed
    Excerpt    string // text snippet from chunk
}
```

**Invariants**:
- ChunkID must be from retrieval result for this query
- PageNumber ≥ 1
- Excerpt is substring of chunk text

#### ScoredChunk
```go
type ScoredChunk struct {
    Chunk Chunk
    Score float64 // relevance score [0.0, 1.0]
}
```

### Database Schema

**SQLite Tables**:

```sql
CREATE TABLE workspaces (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL,
    last_used_at TIMESTAMP NOT NULL
);

CREATE TABLE documents (
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

CREATE TABLE chunks (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL,
    page_number INTEGER NOT NULL,
    text TEXT NOT NULL,
    token_count INTEGER NOT NULL,
    source TEXT NOT NULL,
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
);

CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    workspace_id TEXT NOT NULL,
    started_at TIMESTAMP NOT NULL,
    last_active_at TIMESTAMP NOT NULL,
    title TEXT NOT NULL,
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);

CREATE TABLE turns (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    question TEXT NOT NULL,
    answer_text TEXT NOT NULL,
    answer_confidence TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE TABLE citations (
    id TEXT PRIMARY KEY,
    turn_id TEXT NOT NULL,
    chunk_id TEXT NOT NULL,
    doc_name TEXT NOT NULL,
    page_number INTEGER NOT NULL,
    excerpt TEXT NOT NULL,
    FOREIGN KEY (turn_id) REFERENCES turns(id) ON DELETE CASCADE
);

-- Full-text search index for keyword retrieval
CREATE VIRTUAL TABLE chunks_fts USING fts5(
    chunk_id UNINDEXED,
    text,
    content=chunks,
    content_rowid=rowid
);
```

**Indexes**:
```sql
CREATE INDEX idx_documents_workspace ON documents(workspace_id);
CREATE INDEX idx_documents_status ON documents(status);
CREATE INDEX idx_chunks_document ON chunks(document_id);
CREATE INDEX idx_sessions_workspace ON sessions(workspace_id);
CREATE INDEX idx_turns_session ON turns(session_id);
```

### FAISS Index Structure

The VectorIndex maintains a separate FAISS index per workspace:

```
~/.docintel/
  workspaces/
    <workspace_id>/
      index.faiss       # FAISS index file
      metadata.json     # chunk_id → vector mapping
```

**Metadata JSON**:
```json
{
  "workspace_id": "uuid",
  "dimensions": 384,
  "chunks": [
    {
      "chunk_id": "uuid",
      "vector_index": 0
    }
  ]
}
```


## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property Reflection

After analyzing all acceptance criteria, several redundancies were identified:

- Properties 7.7 and 7.8 both test citation grounding (citations must be from retrieved chunks)
- Properties 5.2, 5.3, and 5.4 all test workspace isolation in retrieval
- Properties 1.5, 9.7, and 22.5 all test cascade deletion
- Properties 3.8 and 13.1 both test atomic ingestion commit
- Properties 3.9 and 13.4 both test rollback and isolation

These have been consolidated into comprehensive properties that capture the full requirement.

### Property 1: Workspace Creation Persistence

*For any* unique workspace name, creating a workspace then querying the Store should return a workspace with that name.

**Validates: Requirements 1.1**

### Property 2: Workspace Name Uniqueness

*For any* workspace name that already exists in the Store, attempting to create another workspace with the same name should fail with an error.

**Validates: Requirements 1.2**

### Property 3: Workspace Listing Completeness

*For any* set of created workspaces, listing workspaces should return exactly the set of workspaces that were created.

**Validates: Requirements 1.3**

### Property 4: Workspace Cascade Deletion

*For any* workspace with associated documents, chunks, embeddings, sessions, and turns, deleting the workspace should atomically remove all associated data such that no queries can retrieve any of it.

**Validates: Requirements 1.5, 9.7, 22.5**

### Property 5: Document Creation with Pending Status

*For any* valid PDF file path and workspace, adding the document should create a document record with status "pending" in the Store.

**Validates: Requirements 2.1**

### Property 6: Document Listing Workspace Scoping

*For any* two workspaces with documents, listing documents in workspace A should return only documents belonging to workspace A and none from workspace B.

**Validates: Requirements 2.2**

### Property 7: Document Cascade Deletion

*For any* document with associated chunks and embeddings, deleting the document should atomically remove all associated data such that no queries can retrieve any of it.

**Validates: Requirements 2.3**

### Property 8: Document Read Flag Update

*For any* document, marking it as read then querying the Store should return the document with isRead=true, and marking it as unread should return isRead=false.

**Validates: Requirements 2.4**

### Property 9: Chunk Token Count Invariant

*For any* text processed by the Chunker, all resulting chunks should have TokenCount ≤ 512.

**Validates: Requirements 3.5**

### Property 10: Chunk Overlap Invariant

*For any* sequence of consecutive chunks from the same document, each pair of adjacent chunks should overlap by at least 64 tokens.

**Validates: Requirements 3.6**

### Property 11: Embedding Generation Completeness

*For any* set of chunks processed by the Embedder, the number of generated vectors should equal the number of input chunks.

**Validates: Requirements 3.7**

### Property 12: Atomic Ingestion Commit

*For any* document that completes ingestion successfully, either all chunks and embeddings are persisted and the status is "indexed", or none are persisted and the status is "pending" or "failed".

**Validates: Requirements 3.8, 3.10, 13.1**

### Property 13: Ingestion Rollback on Failure

*For any* document where ingestion fails at any stage, the Store should contain no chunks or embeddings for that document and the status should be "failed".

**Validates: Requirements 3.9, 13.4**

### Property 14: Reprocess Purge Completeness

*For any* indexed document that is re-processed, after purge completes and before new ingestion starts, the Store should contain no chunks or embeddings for that document.

**Validates: Requirements 4.1, 4.4**

### Property 15: Reprocess Isolation

*For any* document being re-processed, queries should not return any chunks from that document until re-processing completes successfully.

**Validates: Requirements 4.3**

### Property 16: Retrieval Workspace Isolation (Semantic)

*For any* two workspaces with embeddings, semantic search in workspace A should return only chunks belonging to workspace A and none from workspace B.

**Validates: Requirements 5.2**

### Property 17: Retrieval Workspace Isolation (Keyword)

*For any* two workspaces with chunks, keyword search in workspace A should return only chunks belonging to workspace A and none from workspace B.

**Validates: Requirements 5.3**

### Property 18: Retrieval Document Scoping

*For any* workspace with multiple documents, when a document_id is provided, search should return only chunks belonging to that specific document.

**Validates: Requirements 5.5**

### Property 19: Hybrid Search Result Ordering

*For any* search query, the returned results should be ordered by relevance score in descending order.

**Validates: Requirements 6.2**

### Property 20: Search Result Completeness

*For any* search result, each ScoredChunk should include chunk text, document metadata, page number, and relevance score.

**Validates: Requirements 6.3**

### Property 21: Agent Loop Bounded Iteration

*For any* query, the QueryOrchestrator should perform at most 5 retrieval passes before generating an answer.

**Validates: Requirements 7.4**

### Property 22: Prompt Context Grounding

*For any* query, the prompt constructed for the LLM should contain only text from chunks that were retrieved for that query.

**Validates: Requirements 7.5**

### Property 23: Citation Grounding

*For any* answer, all citations should reference chunks that were part of the retrieval result for that specific query.

**Validates: Requirements 7.7, 7.8**

### Property 24: Answer Completeness

*For any* successful query, the returned Answer should contain non-empty text, at least one citation, and a confidence level.

**Validates: Requirements 7.10**

### Property 25: Citation Structure Completeness

*For any* citation, it should include document filename, page number ≥ 1, and a non-empty excerpt.

**Validates: Requirements 8.2**

### Property 26: Multi-Document Citation Coverage

*For any* query where retrieved chunks span multiple documents, the answer's citations should reference chunks from at least two different documents.

**Validates: Requirements 8.4**

### Property 27: Session Creation Workspace Scoping

*For any* workspace, creating a session should result in a session with workspace_id matching that workspace.

**Validates: Requirements 9.1**

### Property 28: Turn Append Persistence

*For any* session, appending a turn then loading the session should return all appended turns in order.

**Validates: Requirements 9.2**

### Property 29: Session Context Loading

*For any* session with existing turns, submitting a new query in that session should include previous turns in the context passed to the QueryOrchestrator.

**Validates: Requirements 9.3**

### Property 30: Session Listing Workspace Scoping

*For any* two workspaces with sessions, listing sessions in workspace A should return only sessions belonging to workspace A and none from workspace B.

**Validates: Requirements 9.4**

### Property 31: Turn Immutability

*For any* turn, after it is written to the Store, its question, answer text, and citations should never change.

**Validates: Requirements 9.6**

### Property 32: Bilingual Embedding Support

*For any* text in English or French, the Embedder should successfully generate a vector embedding without error.

**Validates: Requirements 10.3**

### Property 33: Bilingual LLM Support

*For any* query in English or French, the LLM should successfully generate an answer without error.

**Validates: Requirements 10.4**

### Property 34: Network Isolation

*For any* system operation (ingestion, query, session management), the system should not initiate any outbound network connection.

**Validates: Requirements 12.1**

### Property 35: Network Address Rejection

*For any* configuration that specifies a network address for the LLM or EmbeddingModel, the system should reject the configuration with an error.

**Validates: Requirements 12.4**

### Property 36: Inconsistent State Recovery

*For any* document in an inconsistent state at system startup (e.g., has chunks but status is "pending"), the Store should reset it to "pending" status and remove orphaned chunks.

**Validates: Requirements 13.2**

### Property 37: Atomic Deletion

*For any* workspace or document deletion, either all associated data is removed or none is removed (no partial deletion).

**Validates: Requirements 13.3**

### Property 38: OCR Failure Resilience

*For any* PDF where OCR fails on some pages, the Parser should continue processing remaining pages and return text for all successfully processed pages.

**Validates: Requirements 15.3**

### Property 39: Timeout Session Preservation

*For any* query that times out, the session context should remain unchanged and available for retry.

**Validates: Requirements 16.2**

### Property 40: Referential Integrity - Chunks

*For any* chunk, attempting to create it with a document_id that does not exist should fail with an error.

**Validates: Requirements 22.1**

### Property 41: Referential Integrity - Documents

*For any* document, attempting to create it with a workspace_id that does not exist should fail with an error.

**Validates: Requirements 22.2**

### Property 42: Referential Integrity - Turns

*For any* turn, attempting to create it with a session_id that does not exist should fail with an error.

**Validates: Requirements 22.3**

### Property 43: Referential Integrity - Sessions

*For any* session, attempting to create it with a workspace_id that does not exist should fail with an error.

**Validates: Requirements 22.4**

### Property 44: Configuration Round-Trip

*For any* valid Configuration object, parsing then printing then parsing should produce an equivalent Configuration object.

**Validates: Requirements 23.3**

## Error Handling

### Error Categories

The system handles errors in four categories:

1. **User Input Errors**: Invalid commands, non-existent resources, malformed queries
2. **File Processing Errors**: Corrupt PDFs, password-protected files, OCR failures
3. **System Errors**: Database failures, LLM timeouts, out-of-memory conditions
4. **Invariant Violations**: Referential integrity failures, inconsistent state detection

### Error Handling Strategies

#### User Input Errors

**Strategy**: Validate early, fail fast, provide actionable feedback

**Examples**:
- Workspace name already exists → return error with suggestion to use different name
- Document not found → return error with list of available documents
- Invalid PDF file → return error identifying file type issue

**Implementation**:
- Validation at API boundaries (WorkspaceManager, DocumentManager)
- Descriptive error messages with context
- No state changes on validation failure

#### File Processing Errors

**Strategy**: Isolate failures, preserve partial progress where safe, enable retry

**Examples**:
- Corrupt PDF → set document status to "failed", preserve document record for retry
- OCR failure on single page → log warning, continue with remaining pages
- Password-protected PDF → set status to "failed", return descriptive error

**Implementation**:
- Parser wraps PDF library calls with error handling
- Document status tracks processing state
- Failed documents can be deleted or retried
- OCR failures don't abort entire document processing

#### System Errors

**Strategy**: Rollback transactions, preserve session state, enable recovery

**Examples**:
- Database write failure during ingestion → rollback transaction, set status to "failed"
- LLM timeout during query → preserve session, return timeout error, allow retry
- Out of memory during embedding → rollback transaction, return error

**Implementation**:
- Pipeline wraps ingestion in Store transaction
- SessionManager preserves context on query failure
- TUI displays error and allows retry without losing context
- No partial state persisted on system errors

#### Invariant Violations

**Strategy**: Detect at startup, repair automatically, log for investigation

**Examples**:
- Document has chunks but status is "pending" → reset to "pending", purge chunks
- Orphaned chunks with no parent document → delete chunks
- Session references deleted workspace → delete session

**Implementation**:
- Store runs consistency check at startup
- Automatic repair for known inconsistencies
- Logging for unexpected violations
- Fail-safe: reset to known-good state

### Error Propagation

Errors propagate up the layer stack:

```
External Runtime → Persistence → Ingestion/Query → Orchestration → TUI → User
```

Each layer:
1. Catches errors from layer below
2. Adds context (e.g., "failed to ingest document X")
3. Decides: handle locally or propagate up
4. Wraps in layer-appropriate error type

### Error Recovery Patterns

**Retry with Backoff**:
- LLM timeouts: immediate retry allowed
- Embedding failures: exponential backoff
- Database contention: retry with jitter

**Graceful Degradation**:
- OCR failure on pages: continue with available text
- Partial retrieval results: generate answer with lower confidence
- Missing session context: treat as new session

**Fail-Safe Defaults**:
- Unknown language → treat as "unknown", attempt processing
- Missing confidence score → default to "low"
- Empty retrieval results → return "no documents indexed" error

### Error Logging

All errors logged with:
- Timestamp
- Layer/component
- Operation context (workspace_id, document_id, session_id)
- Error message and stack trace
- User-facing message (if different)

Log levels:
- ERROR: User-visible failures, invariant violations
- WARN: Recoverable issues (OCR failures, partial results)
- INFO: Normal operations (ingestion complete, query answered)
- DEBUG: Detailed flow (retrieval passes, chunk counts)

## Testing Strategy

### Dual Testing Approach

The system requires both unit testing and property-based testing for comprehensive coverage:

**Unit Tests**: Verify specific examples, edge cases, and error conditions
- Specific document ingestion scenarios
- Integration points between components
- Error handling for known failure modes
- Edge cases (empty workspaces, single-page documents, etc.)

**Property Tests**: Verify universal properties across all inputs
- Invariants that must hold for all data (token counts, workspace isolation)
- Round-trip properties (configuration serialization)
- Referential integrity constraints
- Atomic operations and rollback behavior

Both approaches are complementary and necessary. Unit tests catch concrete bugs in specific scenarios, while property tests verify general correctness across the input space.

### Property-Based Testing Configuration

**Library**: Use `gopter` (Go property testing library)

**Configuration**:
- Minimum 100 iterations per property test (due to randomization)
- Each property test must reference its design document property
- Tag format: `Feature: doc-intel-implementation, Property {number}: {property_text}`

**Example**:
```go
// Feature: doc-intel-implementation, Property 9: Chunk Token Count Invariant
func TestChunkTokenCountInvariant(t *testing.T) {
    properties := gopter.NewProperties(nil)
    properties.Property("all chunks have token count <= 512", 
        prop.ForAll(
            func(text string) bool {
                chunks := chunker.Chunk(text)
                for _, chunk := range chunks {
                    if chunk.TokenCount > 512 {
                        return false
                    }
                }
                return true
            },
            gen.AnyString(),
        ))
    properties.TestingRun(t, gopter.ConsoleReporter(t))
}
```

### Test Organization

```
docintel/
  internal/
    workspace/
      manager.go
      manager_test.go          # unit tests
      manager_properties_test.go  # property tests
    docmanager/
      manager.go
      manager_test.go
      manager_properties_test.go
    pipeline/
      pipeline.go
      pipeline_test.go
      pipeline_properties_test.go
    query/
      orchestrator.go
      orchestrator_test.go
      orchestrator_properties_test.go
    retriever/
      retriever.go
      retriever_test.go
      retriever_properties_test.go
    store/
      store.go
      store_test.go
      store_properties_test.go
    vectorindex/
      index.go
      index_test.go
      index_properties_test.go
```

### Unit Test Coverage

**WorkspaceManager**:
- Create workspace with unique name
- Create workspace with duplicate name (error)
- List workspaces
- Switch to existing workspace
- Switch to non-existent workspace (error)
- Delete workspace with no documents
- Delete workspace with documents (cascade)

**DocumentManager**:
- Add valid PDF
- Add non-PDF file (error)
- List documents in workspace
- Delete document with no chunks
- Delete document with chunks (cascade)
- Mark document as read/unread

**Pipeline**:
- Ingest text-native PDF
- Ingest image-only PDF (OCR)
- Ingest with parser failure (rollback)
- Ingest with chunker failure (rollback)
- Ingest with embedder failure (rollback)
- Reprocess indexed document

**Parser**:
- Extract from text-native PDF
- Extract from image-only PDF (OCR)
- Handle corrupt PDF (error)
- Handle password-protected PDF (error)
- Detect English language
- Detect French language
- OCR failure on single page (continue)

**Chunker**:
- Chunk short text (single chunk)
- Chunk long text (multiple chunks with overlap)
- Chunk at sentence boundaries
- Chunk with no sentence boundaries

**Embedder**:
- Embed English text
- Embed French text
- Embed empty text (error)
- Handle embedding model failure (error)

**Retriever**:
- Search with workspace_id only
- Search with workspace_id and document_id
- Search without workspace_id (error)
- Hybrid search merges semantic and keyword results
- Results ordered by score

**QueryOrchestrator**:
- Query with sufficient context (1 pass)
- Query with insufficient context (multiple passes)
- Query with maximum passes (5)
- Query with no indexed documents (error)
- Query with session context
- Query without session context (create new)
- Extract citations from retrieved chunks
- Score confidence based on retrieval quality

**SessionManager**:
- Create session
- Append turn
- Load session turns
- List sessions in workspace
- Resume session
- Verify turn immutability

**Store**:
- Create workspace
- Create document with valid workspace_id
- Create document with invalid workspace_id (error)
- Create chunk with valid document_id
- Create chunk with invalid document_id (error)
- Cascade delete workspace
- Cascade delete document
- Transaction commit
- Transaction rollback
- Keyword search scoped by workspace
- Detect inconsistent state at startup

**VectorIndex**:
- Insert vectors
- Search vectors
- Delete vectors by chunk_id
- Delete workspace
- Workspace isolation in search

### Property Test Coverage

Each of the 44 correctness properties should have a corresponding property-based test:

**Property 1-4**: Workspace operations
**Property 5-8**: Document operations
**Property 9-11**: Chunking and embedding
**Property 12-15**: Ingestion atomicity and reprocessing
**Property 16-20**: Retrieval and search
**Property 21-26**: Query orchestration and citations
**Property 27-31**: Session management
**Property 32-33**: Bilingual support
**Property 34-35**: Network isolation
**Property 36-39**: Error handling and recovery
**Property 40-43**: Referential integrity
**Property 44**: Configuration round-trip

### Integration Tests

**End-to-End Flows**:
- Create workspace → add document → ingest → query → verify answer
- Create workspace → add multiple documents → query across documents
- Create session → multi-turn conversation → resume session
- Reprocess document → verify old chunks removed → query returns new content
- Delete workspace → verify all data removed

**Invariant Enforcement Tests**:
- Verify INV-01: chunk belongs to exactly one document in exactly one workspace
- Verify INV-02: answer never cites source outside active workspace
- Verify INV-03: document never partially indexed at rest
- Verify INV-04: answer never cites content not retrieved for query
- Verify INV-05: re-processing resets document completely
- Verify INV-06: workspace deletion is total and irreversible
- Verify INV-07: system makes zero outbound network calls
- Verify INV-08: conversation history is append-only
- Verify INV-09: retrieval scope is always explicit

### Test Data

**Generators for Property Tests**:
- Random workspace names (valid and invalid)
- Random document paths (valid PDFs, invalid files)
- Random text (English, French, mixed, empty, very long)
- Random queries (English, French, mixed)
- Random workspace/document/session IDs (valid and invalid)

**Fixtures for Unit Tests**:
- Sample PDFs (text-native, image-only, corrupt, password-protected)
- Sample text in English and French
- Sample embeddings
- Sample database states (empty, populated, inconsistent)

### Performance Tests

While not part of the correctness properties, performance tests verify non-functional requirements:

- Ingestion of 20-page PDF completes in <10 seconds
- Query returns first token in <15 seconds
- Memory usage stays under 8GB during inference
- Database queries complete in <100ms
- Vector search completes in <500ms

### Test Execution

**Local Development**:
```bash
# Run all tests
go test ./...

# Run unit tests only
go test -short ./...

# Run property tests only
go test -run Properties ./...

# Run with coverage
go test -cover ./...

# Run specific property test
go test -run TestChunkTokenCountInvariant ./internal/pipeline
```

**CI Pipeline**:
1. Run unit tests (fast feedback)
2. Run property tests (comprehensive coverage)
3. Run integration tests (end-to-end validation)
4. Run invariant enforcement tests (system laws)
5. Generate coverage report
6. Fail build if coverage <80%

### Test Maintenance

- Property tests reference design document properties (traceability)
- Update tests when requirements change
- Add new property tests for new invariants
- Review failed property tests for counterexamples
- Use counterexamples to create regression unit tests

