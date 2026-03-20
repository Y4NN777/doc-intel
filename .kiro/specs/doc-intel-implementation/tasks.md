# Implementation Plan: Doc-Intel

## Overview

This implementation plan follows the layered architecture from the design document, building from the foundation layer up through persistence, ingestion, query, orchestration, and UI layers. Each task is designed to be implementable independently where possible, with property-based tests for invariants and unit tests for specific behaviors.

The implementation order follows module dependencies:
1. Domain entities (foundation)
2. Persistence layer (Store, VectorIndex)
3. Ingestion layer (Parser, Chunker, Embedder, Pipeline)
4. Query layer (Retriever, QueryOrchestrator, SessionManager)
5. Orchestration layer (WorkspaceManager, DocumentManager)
6. UI layer (TUI)
7. Entry point (main)

## Tasks

- [x] 1. Set up project structure and domain entities
  - Create Go module structure following architecture design
  - Define all domain entities in `internal/domain/` package
  - Implement enumerations: DocStatus, Language, Confidence, QueryType, ChunkSource
  - Define core structs: Workspace, Document, Chunk, Vector, Session, Turn, QueryRequest, Answer, Citation, ScoredChunk
  - _Requirements: All requirements (foundation for entire system)_

- [x] 2. Implement Store (SQLite persistence layer)
  - [x] 2.1 Create Store interface and SQLite implementation
    - Implement `internal/store/store.go` with all CRUD operations
    - Create database schema with tables: workspaces, documents, chunks, sessions, turns, citations
    - Set up foreign key constraints for referential integrity
    - Create indexes for performance: workspace_id, document_id, session_id
    - Implement transaction control (BeginTx, Commit, Rollback)
    - _Requirements: 1.1, 1.3, 2.1, 2.2, 9.1, 9.2, 9.4, 22.1-22.4_

  - [ ]* 2.2 Write property test for workspace creation persistence
    - **Property 1: Workspace Creation Persistence**
    - **Validates: Requirements 1.1**

  - [ ]* 2.3 Write property test for workspace name uniqueness
    - **Property 2: Workspace Name Uniqueness**
    - **Validates: Requirements 1.2**

  - [ ]* 2.4 Write property test for workspace listing completeness
    - **Property 3: Workspace Listing Completeness**
    - **Validates: Requirements 1.3**

  - [x] 2.5 Implement cascade delete for workspaces and documents
    - Add CASCADE DELETE constraints to foreign keys
    - Implement atomic deletion in single transaction
    - _Requirements: 1.5, 2.3, 9.7, 22.5_

  - [ ]* 2.6 Write property test for workspace cascade deletion
    - **Property 4: Workspace Cascade Deletion**
    - **Validates: Requirements 1.5, 9.7, 22.5**

  - [ ]* 2.7 Write property test for document cascade deletion
    - **Property 7: Document Cascade Deletion**
    - **Validates: Requirements 2.3**

  - [x] 2.8 Implement full-text search for keyword retrieval
    - Create FTS5 virtual table for chunks
    - Implement SearchKeyword method with workspace scoping
    - _Requirements: 5.3, 6.1_

  - [ ]* 2.9 Write property test for referential integrity constraints
    - **Property 40-43: Referential Integrity**
    - **Validates: Requirements 22.1-22.4**

  - [ ]* 2.10 Write unit tests for Store operations
    - Test workspace CRUD operations
    - Test document CRUD operations with status transitions
    - Test chunk operations
    - Test session and turn operations
    - Test transaction commit and rollback
    - Test keyword search with workspace scoping

- [x] 3. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Implement VectorIndex (FAISS persistence layer)
  - [x] 4.1 Create VectorIndex interface and FAISS implementation
    - Implement `internal/vectorindex/index.go` with Insert, Search, Delete operations
    - Set up workspace-namespaced FAISS indexes
    - Implement metadata tracking (chunk_id → vector_index mapping)
    - Create file structure: `~/.docintel/workspaces/<workspace_id>/index.faiss`
    - _Requirements: 5.2, 6.1_

  - [ ]* 4.2 Write property test for retrieval workspace isolation (semantic)
    - **Property 16: Retrieval Workspace Isolation (Semantic)**
    - **Validates: Requirements 5.2**

  - [ ]* 4.3 Write unit tests for VectorIndex operations
    - Test vector insertion
    - Test ANN search with workspace scoping
    - Test vector deletion by chunk_id
    - Test workspace deletion
    - Test workspace isolation in search

- [ ] 5. Implement Parser (PDF text extraction)
  - [x] 5.1 Create Parser interface and implementation
    - Implement `internal/pipeline/parser.go` with Extract method
    - Integrate PDF text layer extraction library
    - Integrate OCR library for fallback
    - Implement language detection (English, French, unknown)
    - Return PageText structs with page number, text, and source annotation
    - _Requirements: 3.2, 3.3, 3.4_

  - [ ] 5.2 Implement error handling for invalid PDFs
    - Handle corrupt PDFs with descriptive errors
    - Handle password-protected PDFs with descriptive errors
    - Handle invalid file types with descriptive errors
    - _Requirements: 14.1, 14.2, 14.3_

  - [ ] 5.3 Implement OCR failure resilience
    - Continue processing when OCR fails on individual pages
    - Store empty text for failed pages
    - Log warnings for OCR failures
    - _Requirements: 15.1, 15.2, 15.3_

  - [ ]* 5.4 Write property test for OCR failure resilience
    - **Property 38: OCR Failure Resilience**
    - **Validates: Requirements 15.3**

  - [ ]* 5.5 Write unit tests for Parser
    - Test text-native PDF extraction
    - Test image-only PDF with OCR
    - Test corrupt PDF error handling
    - Test password-protected PDF error handling
    - Test English language detection
    - Test French language detection
    - Test OCR failure on single page

- [ ] 6. Implement Chunker (text splitting with overlap)
  - [ ] 6.1 Create Chunker interface and implementation
    - Implement `internal/pipeline/chunker.go` with Chunk method
    - Integrate tokenizer for token counting
    - Split text into chunks with max 512 tokens
    - Ensure minimum 64 token overlap between consecutive chunks
    - Preserve page number and source metadata
    - Generate unique chunk IDs
    - _Requirements: 3.5, 3.6_

  - [ ]* 6.2 Write property test for chunk token count invariant
    - **Property 9: Chunk Token Count Invariant**
    - **Validates: Requirements 3.5**

  - [ ]* 6.3 Write property test for chunk overlap invariant
    - **Property 10: Chunk Overlap Invariant**
    - **Validates: Requirements 3.6**

  - [ ]* 6.4 Write unit tests for Chunker
    - Test short text (single chunk)
    - Test long text (multiple chunks with overlap)
    - Test chunking at sentence boundaries
    - Test chunking with no sentence boundaries

- [ ] 7. Implement Embedder (vector generation)
  - [ ] 7.1 Create Embedder interface and implementation
    - Implement `internal/pipeline/embedder.go` with Embed method
    - Integrate local embedding model via IPC
    - Generate one vector per chunk
    - Support English and French text
    - Handle embedding failures gracefully
    - _Requirements: 3.7, 10.3_

  - [ ]* 7.2 Write property test for embedding generation completeness
    - **Property 11: Embedding Generation Completeness**
    - **Validates: Requirements 3.7**

  - [ ]* 7.3 Write property test for bilingual embedding support
    - **Property 32: Bilingual Embedding Support**
    - **Validates: Requirements 10.3**

  - [ ]* 7.4 Write unit tests for Embedder
    - Test embedding English text
    - Test embedding French text
    - Test embedding empty text (error)
    - Test embedding model failure handling

- [ ] 8. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 9. Implement Pipeline (ingestion orchestration)
  - [ ] 9.1 Create Pipeline interface and implementation
    - Implement `internal/pipeline/pipeline.go` with Ingest and Reprocess methods
    - Coordinate Parser → Chunker → Embedder sequence
    - Wrap entire sequence in Store transaction
    - On success: commit all chunks + embeddings, set status "indexed"
    - On failure: rollback all changes, set status "failed"
    - Report progress during processing
    - _Requirements: 3.1, 3.8, 3.9, 3.10, 13.1, 13.4_

  - [ ] 9.2 Implement reprocessing with purge
    - Purge existing chunks and embeddings before re-ingesting
    - Reset document status to "pending"
    - Execute full ingestion sequence
    - Ensure no old chunks coexist with new chunks
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [ ]* 9.3 Write property test for atomic ingestion commit
    - **Property 12: Atomic Ingestion Commit**
    - **Validates: Requirements 3.8, 3.10, 13.1**

  - [ ]* 9.4 Write property test for ingestion rollback on failure
    - **Property 13: Ingestion Rollback on Failure**
    - **Validates: Requirements 3.9, 13.4**

  - [ ]* 9.5 Write property test for reprocess purge completeness
    - **Property 14: Reprocess Purge Completeness**
    - **Validates: Requirements 4.1, 4.4**

  - [ ]* 9.6 Write property test for reprocess isolation
    - **Property 15: Reprocess Isolation**
    - **Validates: Requirements 4.3**

  - [ ]* 9.7 Write unit tests for Pipeline
    - Test ingestion of text-native PDF
    - Test ingestion of image-only PDF (OCR)
    - Test ingestion with parser failure (rollback)
    - Test ingestion with chunker failure (rollback)
    - Test ingestion with embedder failure (rollback)
    - Test reprocessing of indexed document

- [ ] 10. Implement Retriever (hybrid search)
  - [ ] 10.1 Create Retriever interface and implementation
    - Implement `internal/retriever/retriever.go` with Search method
    - Require explicit workspace_id parameter (no default)
    - Perform semantic search via VectorIndex
    - Perform keyword search via Store
    - Merge and rank results by relevance score
    - Support optional document_id for scoping
    - Return ScoredChunk with text, metadata, page number, and score
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 6.1, 6.2, 6.3_

  - [ ]* 10.2 Write property test for retrieval workspace isolation (keyword)
    - **Property 17: Retrieval Workspace Isolation (Keyword)**
    - **Validates: Requirements 5.3**

  - [ ]* 10.3 Write property test for retrieval document scoping
    - **Property 18: Retrieval Document Scoping**
    - **Validates: Requirements 5.5**

  - [ ]* 10.4 Write property test for hybrid search result ordering
    - **Property 19: Hybrid Search Result Ordering**
    - **Validates: Requirements 6.2**

  - [ ]* 10.5 Write property test for search result completeness
    - **Property 20: Search Result Completeness**
    - **Validates: Requirements 6.3**

  - [ ]* 10.6 Write unit tests for Retriever
    - Test search with workspace_id only
    - Test search with workspace_id and document_id
    - Test search without workspace_id (error)
    - Test hybrid search merges semantic and keyword results
    - Test results ordered by score

- [ ] 11. Implement SessionManager (conversation history)
  - [ ] 11.1 Create SessionManager interface and implementation
    - Implement `internal/session/manager.go` with Create, List, Load, AppendTurn methods
    - Create generates new session scoped to workspace
    - AppendTurn writes question + answer as immutable Turn
    - No update path (append-only)
    - Load retrieves all turns for session
    - List returns sessions scoped to workspace
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6_

  - [ ]* 11.2 Write property test for session creation workspace scoping
    - **Property 27: Session Creation Workspace Scoping**
    - **Validates: Requirements 9.1**

  - [ ]* 11.3 Write property test for turn append persistence
    - **Property 28: Turn Append Persistence**
    - **Validates: Requirements 9.2**

  - [ ]* 11.4 Write property test for session context loading
    - **Property 29: Session Context Loading**
    - **Validates: Requirements 9.3**

  - [ ]* 11.5 Write property test for session listing workspace scoping
    - **Property 30: Session Listing Workspace Scoping**
    - **Validates: Requirements 9.4**

  - [ ]* 11.6 Write property test for turn immutability
    - **Property 31: Turn Immutability**
    - **Validates: Requirements 9.6**

  - [ ]* 11.7 Write unit tests for SessionManager
    - Test session creation
    - Test turn appending
    - Test session turn loading
    - Test session listing in workspace
    - Test session resumption
    - Test turn immutability

- [ ] 12. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 13. Implement QueryOrchestrator (agent loop)
  - [ ] 13.1 Create QueryOrchestrator interface and implementation
    - Implement `internal/query/orchestrator.go` with Query, Summarize, Extract, Compare methods
    - Implement agent loop: plan → retrieve → evaluate → generate → cite → score
    - Bound to maximum 5 retrieval passes
    - Load session context for multi-turn conversations
    - Construct prompts containing only retrieved chunks
    - Call LLM for text generation via IPC
    - Extract citations from retrieved chunks only
    - Score confidence based on retrieval quality
    - Return Answer with text, citations, and confidence
    - Stream tokens progressively
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7, 7.8, 7.9, 7.10, 8.1, 8.2, 8.3, 8.4_

  - [ ] 13.2 Implement error handling for empty workspaces
    - Return error when no indexed documents exist
    - Provide actionable message to user
    - _Requirements: 17.1, 17.2_

  - [ ] 13.3 Implement LLM timeout handling
    - Return timeout error when LLM exceeds timeout
    - Preserve session context on timeout
    - Allow retry without re-entering query
    - _Requirements: 16.1, 16.2, 16.3_

  - [ ]* 13.4 Write property test for agent loop bounded iteration
    - **Property 21: Agent Loop Bounded Iteration**
    - **Validates: Requirements 7.4**

  - [ ]* 13.5 Write property test for prompt context grounding
    - **Property 22: Prompt Context Grounding**
    - **Validates: Requirements 7.5**

  - [ ]* 13.6 Write property test for citation grounding
    - **Property 23: Citation Grounding**
    - **Validates: Requirements 7.7, 7.8**

  - [ ]* 13.7 Write property test for answer completeness
    - **Property 24: Answer Completeness**
    - **Validates: Requirements 7.10**

  - [ ]* 13.8 Write property test for citation structure completeness
    - **Property 25: Citation Structure Completeness**
    - **Validates: Requirements 8.2**

  - [ ]* 13.9 Write property test for multi-document citation coverage
    - **Property 26: Multi-Document Citation Coverage**
    - **Validates: Requirements 8.4**

  - [ ]* 13.10 Write property test for bilingual LLM support
    - **Property 33: Bilingual LLM Support**
    - **Validates: Requirements 10.4**

  - [ ]* 13.11 Write property test for timeout session preservation
    - **Property 39: Timeout Session Preservation**
    - **Validates: Requirements 16.2**

  - [ ]* 13.12 Write unit tests for QueryOrchestrator
    - Test query with sufficient context (1 pass)
    - Test query with insufficient context (multiple passes)
    - Test query with maximum passes (5)
    - Test query with no indexed documents (error)
    - Test query with session context
    - Test query without session context (create new)
    - Test citation extraction from retrieved chunks
    - Test confidence scoring based on retrieval quality

- [ ] 14. Implement WorkspaceManager (orchestration layer)
  - [ ] 14.1 Create WorkspaceManager interface and implementation
    - Implement `internal/workspace/manager.go` with Create, List, Switch, Delete, GetActive methods
    - Create validates name uniqueness before persisting
    - Delete delegates to Store for atomic cascade removal
    - Switch updates active workspace in memory
    - All operations synchronous with immediate error returns
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6_

  - [ ]* 14.2 Write unit tests for WorkspaceManager
    - Test workspace creation with unique name
    - Test workspace creation with duplicate name (error)
    - Test workspace listing
    - Test switching to existing workspace
    - Test switching to non-existent workspace (error)
    - Test workspace deletion with no documents
    - Test workspace deletion with documents (cascade)

- [ ] 15. Implement DocumentManager (orchestration layer)
  - [ ] 15.1 Create DocumentManager interface and implementation
    - Implement `internal/docmanager/manager.go` with Add, List, Delete, MarkRead, GetByID methods
    - Add creates document record with status "pending"
    - Add validates file is valid PDF
    - Delete delegates to Store for atomic removal
    - MarkRead updates read flag only
    - All operations scoped by workspace
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

  - [ ]* 15.2 Write property test for document creation with pending status
    - **Property 5: Document Creation with Pending Status**
    - **Validates: Requirements 2.1**

  - [ ]* 15.3 Write property test for document listing workspace scoping
    - **Property 6: Document Listing Workspace Scoping**
    - **Validates: Requirements 2.2**

  - [ ]* 15.4 Write property test for document read flag update
    - **Property 8: Document Read Flag Update**
    - **Validates: Requirements 2.4**

  - [ ]* 15.5 Write unit tests for DocumentManager
    - Test adding valid PDF
    - Test adding non-PDF file (error)
    - Test listing documents in workspace
    - Test deleting document with no chunks
    - Test deleting document with chunks (cascade)
    - Test marking document as read/unread

- [ ] 16. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 17. Implement TUI (terminal user interface)
  - [ ] 17.1 Create TUI interface and implementation
    - Implement `internal/tui/tui.go` with command dispatch and rendering
    - Implement workspace commands: create, list, switch, delete
    - Implement document commands: add, list, delete, mark read
    - Implement ingestion commands: ingest, reprocess
    - Implement query commands: query, summarize, extract, compare
    - Implement session commands: list, resume
    - Display workspace and document selection
    - Stream answer tokens progressively
    - Display ingestion progress with stage and percentage
    - Display clear error messages with actionable suggestions
    - _Requirements: 11.1, 11.2, 11.3, 11.4_

  - [ ]* 17.2 Write unit tests for TUI
    - Test command parsing and dispatch
    - Test workspace command handling
    - Test document command handling
    - Test query command handling
    - Test session command handling
    - Test error message display
    - Test progress display

- [ ] 18. Implement main entry point and network isolation
  - [ ] 18.1 Create main entry point
    - Implement `cmd/doc-intel/main.go` with initialization and wiring
    - Initialize all components in dependency order
    - Wire components together
    - Enforce network isolation at process start
    - Validate configuration (reject network addresses)
    - Run consistency check at startup
    - Start TUI event loop
    - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5, 13.2, 18.1, 18.2, 18.3_

  - [ ] 18.2 Implement configuration parser and pretty printer
    - Parse configuration into Configuration object
    - Validate configuration (no network addresses)
    - Pretty print Configuration objects
    - _Requirements: 23.1, 23.2, 23.4_

  - [ ] 18.3 Implement inconsistent state recovery at startup
    - Detect documents with chunks but status "pending"
    - Reset inconsistent documents to "pending" and purge orphaned chunks
    - Detect and delete orphaned chunks with no parent document
    - Detect and delete sessions referencing deleted workspaces
    - Log all recovery actions
    - _Requirements: 13.2_

  - [ ]* 18.4 Write property test for network isolation
    - **Property 34: Network Isolation**
    - **Validates: Requirements 12.1**

  - [ ]* 18.5 Write property test for network address rejection
    - **Property 35: Network Address Rejection**
    - **Validates: Requirements 12.4**

  - [ ]* 18.6 Write property test for inconsistent state recovery
    - **Property 36: Inconsistent State Recovery**
    - **Validates: Requirements 13.2**

  - [ ]* 18.7 Write property test for atomic deletion
    - **Property 37: Atomic Deletion**
    - **Validates: Requirements 13.3**

  - [ ]* 18.8 Write property test for configuration round-trip
    - **Property 44: Configuration Round-Trip**
    - **Validates: Requirements 23.3**

  - [ ]* 18.9 Write unit tests for main entry point
    - Test component initialization
    - Test component wiring
    - Test network isolation enforcement
    - Test configuration validation
    - Test consistency check at startup

- [ ] 19. Integration testing and end-to-end validation
  - [ ]* 19.1 Write integration test: create workspace → add document → ingest → query
    - Test full flow from workspace creation to query answer
    - Verify answer contains citations from ingested document
    - _Requirements: All requirements (end-to-end validation)_

  - [ ]* 19.2 Write integration test: multi-document query
    - Create workspace with multiple documents
    - Ingest all documents
    - Query across documents
    - Verify citations from multiple documents
    - _Requirements: 8.4_

  - [ ]* 19.3 Write integration test: multi-turn conversation
    - Create session
    - Submit multiple queries in same session
    - Verify context preservation across turns
    - Resume session and continue conversation
    - _Requirements: 9.1, 9.2, 9.3, 9.5_

  - [ ]* 19.4 Write integration test: document reprocessing
    - Ingest document
    - Query and verify results
    - Reprocess document
    - Verify old chunks removed
    - Query and verify new results
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [ ]* 19.5 Write integration test: workspace deletion cascade
    - Create workspace with documents, chunks, embeddings, sessions, turns
    - Delete workspace
    - Verify all associated data removed from Store and VectorIndex
    - _Requirements: 1.5, 9.7, 22.5_

  - [ ]* 19.6 Write invariant enforcement tests
    - Test INV-01: chunk belongs to exactly one document in exactly one workspace
    - Test INV-02: answer never cites source outside active workspace
    - Test INV-03: document never partially indexed at rest
    - Test INV-04: answer never cites content not retrieved for query
    - Test INV-05: re-processing resets document completely
    - Test INV-06: workspace deletion is total and irreversible
    - Test INV-07: system makes zero outbound network calls
    - Test INV-08: conversation history is append-only
    - Test INV-09: retrieval scope is always explicit

- [ ] 20. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Property tests validate universal correctness properties from the design document
- Unit tests validate specific examples and edge cases
- Integration tests validate end-to-end flows and invariant enforcement
- The implementation follows strict module dependency order to avoid circular dependencies
- All 44 correctness properties from the design document are covered by property tests
- Network isolation is enforced at process start (INV-07)
- All state transitions are atomic (INV-03)
- Workspace isolation is enforced at all layers (INV-02, INV-09)
