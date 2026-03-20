# Requirements Document

## Introduction

Doc-Intel is a terminal-native document intelligence system that enables developers to organize private PDF documents into workspace-scoped collections, process them locally, and ask natural language questions with grounded answers. The system enforces zero network calls, workspace isolation, and atomic operations while supporting English and French documents and queries.

This requirements document translates the comprehensive engineering documentation into actionable implementation requirements following EARS patterns and INCOSE quality rules.

## Glossary

- **System**: The Doc-Intel application as a whole
- **Store**: The SQLite-based persistence layer for structured data
- **VectorIndex**: The FAISS-based persistence layer for embeddings and ANN search
- **Pipeline**: The ingestion orchestration component that coordinates extract → chunk → embed → store
- **Retriever**: The hybrid search component combining semantic and keyword retrieval
- **QueryOrchestrator**: The agent loop component that plans, retrieves, generates, cites, and scores answers
- **WorkspaceManager**: The component managing workspace lifecycle operations
- **DocumentManager**: The component managing document lifecycle operations
- **SessionManager**: The component managing conversation history
- **TUI**: The terminal user interface component
- **Parser**: The subcomponent that extracts text from PDFs (text layer or OCR)
- **Chunker**: The subcomponent that splits text into semantic units with overlap
- **Embedder**: The subcomponent that generates vector embeddings from text
- **LLM**: The local language model binary used for text generation
- **EmbeddingModel**: The local embedding model binary used for vector generation

## Requirements

### Requirement 1: Workspace Lifecycle Management

**User Story:** As a developer, I want to create and manage isolated workspaces, so that I can organize documents by project or topic without cross-contamination.

#### Acceptance Criteria

1. WHEN the user creates a workspace with a unique name, THE WorkspaceManager SHALL persist the workspace to the Store
2. WHEN the user attempts to create a workspace with a name that already exists, THE WorkspaceManager SHALL reject the operation and return an error
3. WHEN the user lists workspaces, THE WorkspaceManager SHALL retrieve all workspaces from the Store and display them via the TUI
4. WHEN the user switches to an existing workspace, THE WorkspaceManager SHALL set that workspace as active for subsequent operations
5. WHEN the user deletes a workspace, THE WorkspaceManager SHALL instruct the Store to atomically remove the workspace and all associated documents, chunks, embeddings, and conversation history
6. WHEN the user attempts to switch to a workspace that does not exist, THE WorkspaceManager SHALL return an error identifying the missing workspace

### Requirement 2: Document Lifecycle Management

**User Story:** As a developer, I want to add, list, and remove PDF documents from my workspace, so that I can control what content is available for querying.

#### Acceptance Criteria

1. WHEN the user adds a PDF file to the active workspace, THE DocumentManager SHALL create a document record with status "pending" in the Store
2. WHEN the user lists documents in the active workspace, THE DocumentManager SHALL retrieve all documents scoped to that workspace from the Store
3. WHEN the user deletes a document, THE DocumentManager SHALL instruct the Store to atomically remove the document and all associated chunks and embeddings
4. WHEN the user marks a document as read or unread, THE DocumentManager SHALL update the document's read flag in the Store
5. WHEN the user attempts to add a file that is not a valid PDF, THE DocumentManager SHALL reject the operation and return an error

### Requirement 3: Document Ingestion Pipeline

**User Story:** As a developer, I want to process PDF documents into searchable chunks, so that I can query their content with natural language.

#### Acceptance Criteria

1. WHEN the user triggers ingestion for a pending document, THE Pipeline SHALL coordinate the Parser, Chunker, and Embedder in sequence
2. WHEN the Parser processes a text-native PDF page, THE Parser SHALL extract text directly from the text layer
3. WHEN the Parser encounters a page with no text layer, THE Parser SHALL fall back to OCR for that page
4. WHEN the Parser completes extraction, THE Parser SHALL detect the document's primary language (English, French, or unknown) and store it as metadata
5. WHEN the Chunker receives extracted text, THE Chunker SHALL split it into chunks where no chunk exceeds 512 tokens
6. WHEN the Chunker creates chunks, THE Chunker SHALL ensure each chunk overlaps with the next by a minimum of 64 tokens
7. WHEN the Embedder receives chunks, THE Embedder SHALL generate a vector embedding for each chunk using the local EmbeddingModel
8. WHEN all pipeline stages complete successfully, THE Pipeline SHALL instruct the Store to commit all chunks and embeddings atomically
9. IF any pipeline stage fails, THEN THE Pipeline SHALL instruct the Store to rollback all changes for that document and set its status to "failed"
10. WHEN ingestion completes successfully, THE Pipeline SHALL update the document status to "indexed" in the Store

### Requirement 4: Document Re-processing

**User Story:** As a developer, I want to re-process an existing document, so that I can update its indexed content after modifications.

#### Acceptance Criteria

1. WHEN the user triggers re-processing for an indexed document, THE Pipeline SHALL instruct the Store to purge all existing chunks and embeddings for that document
2. WHEN the purge completes, THE Pipeline SHALL reset the document status to "pending" and execute the full ingestion sequence
3. WHILE re-processing is in progress, THE System SHALL NOT allow queries to access the document's old chunks
4. WHEN re-processing completes, THE System SHALL ensure no old chunks coexist with new chunks for the same document

### Requirement 5: Workspace-Scoped Retrieval

**User Story:** As a developer, I want my queries to only search within the active workspace, so that I never receive results from unrelated projects.

#### Acceptance Criteria

1. WHEN the Retriever performs a search, THE Retriever SHALL require an explicit workspace_id parameter
2. WHEN the Retriever queries the VectorIndex, THE Retriever SHALL scope the search to embeddings belonging to the specified workspace_id
3. WHEN the Retriever queries the Store for keyword matching, THE Retriever SHALL scope the search to chunks belonging to the specified workspace_id
4. THE Retriever SHALL NOT provide any search operation that crosses workspace boundaries
5. WHEN a document_id is provided, THE Retriever SHALL further scope the search to chunks belonging to that specific document within the workspace

### Requirement 6: Hybrid Search Retrieval

**User Story:** As a developer, I want the system to find relevant content using both semantic similarity and keyword matching, so that I get comprehensive results.

#### Acceptance Criteria

1. WHEN the Retriever performs a search, THE Retriever SHALL execute both semantic search via VectorIndex and keyword search via Store
2. WHEN the Retriever receives results from both search methods, THE Retriever SHALL merge and rank them by relevance score
3. WHEN the Retriever returns results, THE Retriever SHALL include the chunk text, document metadata, page number, and relevance score for each result

### Requirement 7: Query Orchestration with Agent Loop

**User Story:** As a developer, I want the system to intelligently retrieve context and generate grounded answers, so that I receive accurate responses with source citations.

#### Acceptance Criteria

1. WHEN the user submits a query, THE QueryOrchestrator SHALL initiate an agent loop to plan, retrieve, and generate an answer
2. WHEN the QueryOrchestrator plans retrieval, THE QueryOrchestrator SHALL call the Retriever with the workspace_id and optional document_id
3. WHEN the QueryOrchestrator evaluates retrieved context, THE QueryOrchestrator SHALL determine if sufficient information is available to answer the query
4. IF context is insufficient, THEN THE QueryOrchestrator SHALL perform additional retrieval passes up to a maximum of 5 total passes
5. WHEN the QueryOrchestrator generates an answer, THE QueryOrchestrator SHALL construct a prompt containing only the retrieved chunks as context
6. WHEN the QueryOrchestrator calls the LLM, THE QueryOrchestrator SHALL pass the prompt and receive generated text
7. WHEN the QueryOrchestrator extracts citations, THE QueryOrchestrator SHALL map answer content to source chunks that were part of the retrieval result
8. THE QueryOrchestrator SHALL NOT cite any chunk that was not retrieved for the current query
9. WHEN the QueryOrchestrator scores confidence, THE QueryOrchestrator SHALL assign a confidence level (High, Medium, or Low) based on retrieval quality and answer grounding
10. WHEN the QueryOrchestrator completes, THE QueryOrchestrator SHALL return an Answer containing text, citations, and confidence level

### Requirement 8: Answer Grounding and Citation

**User Story:** As a developer, I want every answer to include source references, so that I can verify the information and understand its origin.

#### Acceptance Criteria

1. WHEN the QueryOrchestrator produces an answer, THE QueryOrchestrator SHALL include at least one citation for every factual claim
2. WHEN the QueryOrchestrator creates a citation, THE Citation SHALL include the document filename, page number, and a text excerpt from the source chunk
3. THE QueryOrchestrator SHALL NOT generate citations from the LLM's general knowledge
4. WHEN multiple documents contribute to an answer, THE QueryOrchestrator SHALL include citations from all relevant sources

### Requirement 9: Conversation Session Management

**User Story:** As a developer, I want to maintain conversation context within a session and resume past conversations, so that I can have multi-turn interactions with my documents.

#### Acceptance Criteria

1. WHEN the user starts a new query without specifying a session, THE SessionManager SHALL create a new session scoped to the active workspace
2. WHEN the QueryOrchestrator completes a query, THE SessionManager SHALL append the question and answer as a Turn to the current session
3. WHEN the user submits a follow-up query in the same session, THE QueryOrchestrator SHALL include previous turns as context for reference resolution
4. WHEN the user lists past sessions, THE SessionManager SHALL retrieve all sessions scoped to the active workspace from the Store
5. WHEN the user resumes a past session, THE SessionManager SHALL load all turns from that session and set it as the active session
6. THE SessionManager SHALL NOT modify existing turns after they are written
7. WHEN a workspace is deleted, THE SessionManager SHALL ensure all associated sessions and turns are deleted atomically

### Requirement 10: Bilingual Query Support

**User Story:** As a developer, I want to ask questions in English or French regardless of the document language, so that I can work in my preferred language.

#### Acceptance Criteria

1. WHEN the user submits a query in English, THE QueryOrchestrator SHALL process it correctly regardless of whether source documents are in English or French
2. WHEN the user submits a query in French, THE QueryOrchestrator SHALL process it correctly regardless of whether source documents are in English or French
3. WHEN the EmbeddingModel generates embeddings, THE EmbeddingModel SHALL support both English and French text
4. WHEN the LLM generates answers, THE LLM SHALL support both English and French queries and responses

### Requirement 11: Terminal User Interface

**User Story:** As a developer, I want a terminal-native interface with streaming answers and progress display, so that I can work efficiently without leaving my terminal.

#### Acceptance Criteria

1. WHEN the LLM generates an answer, THE TUI SHALL stream tokens progressively as they are produced
2. WHEN the Pipeline processes a document, THE TUI SHALL display ingestion progress including current stage and completion percentage
3. WHEN the user issues a command, THE TUI SHALL display workspace and document selection without requiring process restart
4. WHEN an error occurs, THE TUI SHALL display a clear error message identifying the problem and suggesting corrective action

### Requirement 12: Network Isolation

**User Story:** As a developer, I want absolute certainty that no document content leaves my machine, so that I can trust the system with sensitive information.

#### Acceptance Criteria

1. THE System SHALL NOT initiate any outbound network connection during any operation
2. WHEN the System calls the LLM, THE System SHALL use local inter-process communication only
3. WHEN the System calls the EmbeddingModel, THE System SHALL use local inter-process communication only
4. THE System SHALL NOT accept any network address as configuration for the LLM or EmbeddingModel
5. THE System SHALL be verifiable as network-isolated via system call tracing tools

### Requirement 13: Atomic State Transitions

**User Story:** As a developer, I want the system to never leave documents in partial states, so that query results are always predictable and complete.

#### Acceptance Criteria

1. WHEN the Store commits document ingestion, THE Store SHALL ensure all chunks and embeddings are written atomically
2. IF the Store detects a document in an inconsistent state at startup, THEN THE Store SHALL reset it to "pending" status
3. WHEN the Store performs a workspace or document deletion, THE Store SHALL remove all associated data in a single transaction
4. THE Store SHALL NOT persist any intermediate state during ingestion that would be visible to queries

### Requirement 14: Error Handling for Invalid PDFs

**User Story:** As a developer, I want clear error messages when a PDF cannot be processed, so that I can take corrective action.

#### Acceptance Criteria

1. WHEN the Parser encounters a corrupt PDF, THE Parser SHALL set the document status to "failed" and return a descriptive error message
2. WHEN the Parser encounters a password-protected PDF, THE Parser SHALL set the document status to "failed" and return a descriptive error message
3. WHEN the Parser encounters an invalid file, THE Parser SHALL set the document status to "failed" and return a descriptive error message
4. WHEN a document is in "failed" status, THE System SHALL allow the user to retry processing or remove the document

### Requirement 15: Error Handling for OCR Failures

**User Story:** As a developer, I want the system to continue processing when OCR fails on individual pages, so that I don't lose the entire document.

#### Acceptance Criteria

1. WHEN the Parser's OCR produces no text for a page, THE Parser SHALL store an empty text entry for that page
2. WHEN the Parser's OCR fails on a page, THE Parser SHALL log a warning with the page number
3. WHEN the Parser's OCR fails on a page, THE Parser SHALL continue processing remaining pages

### Requirement 16: Error Handling for LLM Timeouts

**User Story:** As a developer, I want to retry queries when the LLM times out, so that temporary failures don't lose my conversation context.

#### Acceptance Criteria

1. WHEN the LLM fails to return a response within the timeout period, THE QueryOrchestrator SHALL return a timeout error to the user
2. WHEN a timeout occurs, THE SessionManager SHALL preserve the session context
3. WHEN a timeout occurs, THE TUI SHALL allow the user to retry the query without re-entering it

### Requirement 17: Error Handling for Empty Workspaces

**User Story:** As a developer, I want clear guidance when I query an empty workspace, so that I understand why no results are returned.

#### Acceptance Criteria

1. WHEN the user submits a query in a workspace with no indexed documents, THE QueryOrchestrator SHALL return an error message
2. WHEN the error occurs, THE TUI SHALL inform the user that no documents are indexed and prompt them to add and process documents

### Requirement 18: Single Binary Deployment

**User Story:** As a developer, I want to run Doc-Intel as a single executable with no runtime dependencies, so that deployment is simple and portable.

#### Acceptance Criteria

1. THE System SHALL compile to a single executable binary
2. THE System SHALL run on Ubuntu 22.04+ without requiring Python, Node, Docker, or any interpreter
3. THE System SHALL embed all necessary Go dependencies at compile time

### Requirement 19: Performance - Ingestion

**User Story:** As a developer, I want fast document processing, so that I can start querying quickly.

#### Acceptance Criteria

1. WHEN the Pipeline processes a text-native PDF of 20 pages on target hardware (Ryzen 3, 16GB RAM), THE Pipeline SHALL complete ingestion in under 10 seconds

### Requirement 20: Performance - Query Latency

**User Story:** As a developer, I want fast query responses, so that I can maintain my flow while working.

#### Acceptance Criteria

1. WHEN the user submits a query on target hardware (Ryzen 3, 16GB RAM), THE System SHALL return the first token of the answer within 15 seconds

### Requirement 21: Performance - Memory Ceiling

**User Story:** As a developer, I want the system to run efficiently on modest hardware, so that I don't need expensive equipment.

#### Acceptance Criteria

1. WHEN the System performs inference on target hardware (Ryzen 3, 16GB RAM), THE System SHALL NOT consume more than 8GB of RAM

### Requirement 22: Data Integrity - Referential Integrity

**User Story:** As a developer, I want the system to maintain data consistency, so that I never encounter orphaned or corrupted data.

#### Acceptance Criteria

1. THE Store SHALL enforce that every chunk references a valid document_id
2. THE Store SHALL enforce that every document references a valid workspace_id
3. THE Store SHALL enforce that every turn references a valid session_id
4. THE Store SHALL enforce that every session references a valid workspace_id
5. WHEN a parent entity is deleted, THE Store SHALL cascade delete all child entities atomically

### Requirement 23: Parser and Pretty Printer for Configuration

**User Story:** As a developer, I want the system to correctly parse and serialize configuration, so that settings persist reliably across sessions.

#### Acceptance Criteria

1. WHEN the System reads configuration, THE Parser SHALL parse it into a Configuration object
2. WHEN the System writes configuration, THE PrettyPrinter SHALL format Configuration objects into valid configuration files
3. FOR ALL valid Configuration objects, parsing then printing then parsing SHALL produce an equivalent object (round-trip property)
4. WHEN invalid configuration is provided, THE Parser SHALL return a descriptive error

