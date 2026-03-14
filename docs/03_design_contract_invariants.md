# Contract & Invariants — System Laws
## Doc-Intel — Software Engineering Foundations

---

## 1. System Contract

Doc-Intel makes the following promises to its environment regardless of how it is implemented internally.

```
┌─────────────────────────────────────────────────────┐
│                   System Contract                   │
│                                                     │
│  INPUT              SHIELD               OUTPUT     │
│  ─────────          ────────             ────────   │
│  PDF files    →   Invariants      →   Grounded      │
│  User queries      (always true)       answers      │
│  Commands                              with sources  │
│                                                     │
│  What passes through the shield is guaranteed.      │
│  What violates the shield is rejected.              │
└─────────────────────────────────────────────────────┘
```

### 1.1 What the system promises

| Promise | Description |
|---------|-------------|
| **Grounded answers** | Every answer is derived from content that exists in the user's indexed documents. The system never generates content from general knowledge alone without grounding it in a source. |
| **Source attribution** | Every answer includes the document filename and page number it draws from. |
| **Workspace isolation** | A query in workspace A will never return content from workspace B. Ever. |
| **Atomic state transitions** | A document is either fully indexed or not indexed at all. The system never leaves a document in a partial state at rest. |
| **Local execution** | No document content, query, or response ever leaves the machine. |
| **Explicit control** | The system never modifies its index or state without an explicit user command. |

### 1.2 What the system forbids

| Forbidden | Consequence if violated |
|-----------|------------------------|
| Outbound network calls | Privacy contract broken. Entire value proposition lost. |
| Fabricated citations | User makes decisions based on non-existent sources. Trust destroyed. |
| Cross-workspace data leakage | Workspace isolation — the core organizational guarantee — is void. |
| Partial document indexing at rest | Query results become unpredictable. Answers may be incomplete without warning. |
| Background state modification | User loses control over when and what gets processed. |

---

## 2. Invariants

These rules must remain true under all conditions — during refactors, stack changes, or feature additions. They are the laws of physics of this system. If any invariant is false at any point in time, the system is in an illegal state.

---

**INV-01 — A chunk belongs to exactly one document in exactly one workspace.**

```
chunk.document_id  →  documents.id        (must exist)
documents.workspace_id  →  workspaces.id  (must exist)
```

A chunk cannot exist without a parent document. A document cannot exist without a parent workspace. Orphaned data is illegal.

---

**INV-02 — An answer never cites a source outside the active workspace.**

The retrieval operation is scoped by `workspace_id` at query time. Post-hoc filtering is not acceptable. The workspace boundary is enforced at the point of retrieval, not after.

---

**INV-03 — A document is never partially indexed at rest.**

Ingestion is a transaction. Either all chunks and embeddings for a document are committed, or none are. On failure: full rollback. The document status is either `pending`, `indexed`, or `failed`. No other state exists at rest.

---

**INV-04 — An answer never cites content that was not retrieved for that query.**

The system MUST NOT reference a chunk in an answer unless that chunk was part of the retrieval result for that specific query session. Citations are not generated — they are extracted from retrieved context.

---

**INV-05 — Re-processing a document resets it completely.**

When a document is re-processed, all previous chunks and embeddings are deleted before new ones are written. Old state and new state MUST NOT coexist for the same document at any point.

---

**INV-06 — Workspace deletion is total and irreversible.**

Deleting a workspace removes all associated documents, chunks, embeddings, and conversation history in a single atomic operation. After deletion, no query or inspection can return data from that workspace.

---

**INV-07 — The system makes zero outbound network calls.**

This invariant has no exception. It is not configurable. It cannot be overridden by a flag or environment variable. It applies to ingestion, querying, summarization, and session management equally.

---

**INV-08 — Conversation history is append-only within a session.**

Turns are written to persistent storage as they occur. History is never modified retroactively. A turn, once written, is immutable.

---

**INV-09 — The retrieval scope is always explicit.**

Every retrieval operation carries an explicit `workspace_id` and optionally a `document_id`. There is no "global" retrieval mode that crosses workspace boundaries. Implicit scope does not exist.

---

## 3. Technical Constraints

These are the walls the design must respect. They are not implementation choices — they are given conditions that determine the architecture.

| Constraint | Impact on design |
|------------|-----------------|
| **Single-user, local machine** | No auth layer, no network stack, no multi-tenancy |
| **Target hardware: Ryzen 3, 16GB RAM, Ubuntu** | Model selection constrained to quantized models ≤ 8GB RAM footprint |
| **Single compiled binary** | No runtime dependency on Python, Node, Docker, or any interpreter |
| **English + French documents and queries** | OCR engine and embedding model must support both languages |
| **Manual ingestion only** | No filesystem watcher, no background process, no daemon |
| **TUI-only interface in v1** | No HTTP server, no REST API, no browser interface |

---

## 4. Coherence Checklist

Verifying that this contract is coherent with the PRD and SRS.

| Check | Status |
|-------|--------|
| All actors from SRS are reflected in the contract | ✓ User, File System, Local LLM |
| Every UC from PRD maps to at least one promise | ✓ UC-01→INV-06, UC-05→INV-02/09, UC-03→INV-03 |
| Every MUST NOT from SRS maps to a forbidden action | ✓ NF-01→INV-07, BR-03→INV-04, BR-02→INV-01 |
| No invariant introduces an architectural decision | ✓ No mention of SQLite, FAISS, Go, or any library |
| Technical constraints are conditions, not choices | ✓ None say "use X" — they say "must support Y" |

---

*Previous: [02 — Requirements: SRS](./02_requirements_srs.md)*
*Next: [04 — Transition: Requirements to Architecture](./04_transition_req_to_arch.md)*
