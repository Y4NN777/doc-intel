# Doc-Intel

> Terminal-native document intelligence. Fully local. Zero network calls.

Ask questions about your private PDF documents using natural language. All processing happens on your machine.

## The Problem

You accumulate private documents across projects. Knowledge stays trapped in files. Existing local tools require browsers, servers, or heavy runtimes. None fit developers who live in the terminal.

## What It Does

- Organize PDFs into workspace-scoped collections
- Process documents locally with explicit control
- Ask questions in English or French
- Get grounded answers with source citations and confidence levels
- Resume conversations across sessions

## Non-Negotiables

- **Zero network calls**: Nothing leaves your machine
- **Single binary**: No runtime dependencies (no Python, Node, Docker)
- **Workspace isolation**: Documents and conversations never leak across boundaries
- **Atomic operations**: Documents are fully indexed or not indexed — no partial states

## Stack

- Go (single compiled binary)
- Ubuntu 22.04+
- Local LLM (quantized, ≤8GB RAM)
- Local embedding model
- SQLite (structured data) + FAISS (vector search)
- Terminal UI

## Status

Project initialized. Core module structure in place. Implementation in progress.

## Project Structure

```
docintel/
├── cmd/doc-intel/          # Entry point (enforces INV-07)
├── internal/
│   ├── domain/             # Shared entities (no dependencies)
│   ├── store/              # SQLite persistence layer
│   ├── vectorindex/        # FAISS vector search
│   ├── workspace/          # Workspace lifecycle (INV-06)
│   ├── docmanager/         # Document lifecycle
│   ├── pipeline/           # Ingestion orchestration (INV-03, INV-05)
│   ├── retriever/          # Scoped search (INV-02, INV-09)
│   ├── query/              # Query orchestration (INV-04)
│   └── session/            # Conversation history (INV-08)
└── docs/                   # Engineering documentation
```

## Documentation

Engineering docs in `docs/`:
- `00` — Engineering mindset and problem definition
- `01` — Product requirements (PRD)
- `02` — Software requirements (SRS)
- `03` — System contract and invariants
- `04` — Requirements to architecture mapping
- `05` — UML and C4 modeling
- `06` — Architecture and module design

## Getting Started

```bash
cd docintel
go build -o bin/doc-intel ./cmd/doc-intel
```

### Conversation Memory
- Persistent conversation history per workspace
- Resume past sessions and continue where you left off
- Context-aware follow-up questions within sessions

### Terminal-Native Experience
- Streaming answers that appear token by token
- Real-time ingestion progress display
- No browser required — pure TUI interface

## System Guarantees

Doc-Intel makes these promises regardless of internal implementation:

- **Grounded Answers**: Every answer derives from your indexed documents, never from general knowledge alone
- **Source Attribution**: Every answer includes document filename and page number
- **Workspace Isolation**: Queries in workspace A never access workspace B
- **Atomic Operations**: Documents are either fully indexed or not indexed — no partial states
- **Local Execution**: No document content, query, or response ever leaves your machine
- **Explicit Control**: System never modifies state without your explicit command

## Architecture

Doc-Intel follows a layered pipeline architecture with an embedded agent loop:

```
┌─────────────────────────────────────────────────────────┐
│                    User Interface (TUI)                 │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────┴────────────────────────────────┐
│              Orchestration Layer                        │
│   Workspace Manager  │  Document Manager                │
└────────────┬───────────┴──────────────┬─────────────────┘
             │                          │
      ┌──────┴──────┐            ┌──────┴──────┐
      │  Ingestion  │            │    Query    │
      │   Pipeline  │            │  Agent Loop │
      │             │            │             │
      │ Extract →   │            │ Retrieve →  │
      │ Chunk →     │            │ Reason →    │
      │ Embed →     │            │ Generate →  │
      │ Store       │            │ Cite        │
      └──────┬──────┘            └──────┬──────┘
             │                          │
┌────────────┴──────────────────────────┴─────────────────┐
│              Persistence Layer                           │
│   Store (SQLite)  │  Vector Index (FAISS)                │
└──────────────────────────────────────────────────────────┘
             │
┌────────────┴──────────────────────────────────────────┐
│         External Runtime (Local IPC Only)             │
│   Local LLM  │  Embedding Model  │  File System       │
└───────────────────────────────────────────────────────┘
```

## Technical Stack

- **Language**: Go (single compiled binary)
- **Target Platform**: Ubuntu 24.04+ (Linux), Mac & Windows ( Later )
- **LLM**: Local quantized model (≤8GB RAM footprint)
- **Embeddings**: Local embedding model
- **Storage**: SQLite (structured data) + Vector index (embeddings)
- **Interface**: Terminal UI (TUI)
- **Languages Supported**: English and French (documents and queries)

## Performance Targets

- Process 20-page text PDF in under 10 seconds
- First answer token within 15 seconds of query
- Memory usage under 8GB during inference
- Target hardware: Ryzen 3 CPU, 16GB RAM

## Use Cases

- **UC-01**: Organize documents by project in isolated workspaces
- **UC-02**: Add and manage documents within workspaces
- **UC-03**: Process documents on demand with explicit control
- **UC-04**: Ask questions in natural language (EN/FR)
- **UC-05**: Scope questions to specific workspaces or documents
- **UC-06**: Receive sourced answers with confidence levels
- **UC-07**: Extract specific data points (dates, names, figures)
- **UC-08**: Compare information across multiple documents
- **UC-09**: Summarize documents with page references
- **UC-10**: Track document read status
- **UC-11**: Resume past conversations from previous sessions

## Out of Scope (v1)

- Web or graphical UI (planned v2)
- Multi-user support or authentication
- Cloud sync or remote access
- Non-PDF document formats
- Automatic background ingestion
- Internet search or external knowledge
- Model fine-tuning or training

## System Invariants

These rules remain true under all conditions:

1. A chunk belongs to exactly one document in exactly one workspace
2. An answer never cites a source outside the active workspace
3. A document is never partially indexed at rest
4. An answer never cites content that wasn't retrieved for that query
5. Re-processing a document resets it completely
6. Workspace deletion is total and irreversible
7. The system makes zero outbound network calls
8. Conversation history is append-only within a session
9. Retrieval scope is always explicit

## Project Status

Currently in design phase. Implementation has not yet begun.

## Documentation

Comprehensive engineering documentation is available in `docintel/docs/`:

- `00_foundation_engineering_mindset.md` - Core philosophy and problem definition
- `01_requirements_prd.md` - Product requirements and use cases
- `02_requirements_srs.md` - Software requirements specification
- `03_design_contract_invariants.md` - System contract and invariants
- `04_transition_req_to_arch.md` - Requirements to architecture mapping
- `05_modeling.md` - UML and C4 diagrams
- `06_architecture.md` - Detailed architecture and module design

## License

[To be determined]

## Contributing


---

**Built with the engineer mindset**: Define the problem before solving it.
