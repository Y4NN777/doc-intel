# Transition — From Requirements to Architecture
## Doc-Intel — Software Engineering Foundations

---

## 1. The Method

Every guarantee from `03` must be assigned to exactly one owner.
Not a technology. Not a library. A **conceptual responsibility**.

The chain is:

```
Requirement → Guarantee → Responsibility → Component
```

If a guarantee has no owner, it will be violated.
If two components share ownership of the same guarantee, it will be violated by the one that blinks first.

---

## 2. Guarantee-to-Responsibility Translation

### INV-01 — A chunk belongs to exactly one document in exactly one workspace

**Guarantee**: The system guarantees that no chunk can exist without a valid parent document, and no document can exist without a valid parent workspace. Orphaned data is an illegal state.

**Who must carry this?**
Something must own the write path for all persistent data — documents, chunks, embeddings. It must enforce the parent-child relationship at the moment of writing, not after.

**Responsibility**: The **Data Store** owns all write operations. It enforces referential integrity on every insert. No other component writes to persistent storage directly.

**Component**: `Store`

---

### INV-02 — An answer never cites a source outside the active workspace

**Guarantee**: The system guarantees that retrieval is always scoped to the active workspace. Cross-workspace results are impossible, not just unlikely.

**Who must carry this?**
Something must own the act of finding relevant content. Scoping cannot be delegated to the caller — if every caller is responsible for remembering to scope, one will eventually forget.

**Responsibility**: The **Retriever** owns all search operations and applies workspace scope internally. The caller provides a query and a workspace ID. Scoping is not optional.

**Component**: `Retriever`

---

### INV-03 — A document is never partially indexed at rest

**Guarantee**: The system guarantees that ingestion is atomic. Either the full pipeline completes for a document, or nothing is written.

**Who must carry this?**
Something must own the orchestration of the ingestion steps — extraction, chunking, embedding, storage — and wrap them in a single transactional boundary. No individual step can own this because the guarantee spans all of them.

**Responsibility**: The **Pipeline** owns the ingestion transaction boundary. It coordinates all steps and instructs the Store to commit or rollback as a unit.

**Component**: `Pipeline`

---

### INV-04 — An answer never cites content that was not retrieved for that query

**Guarantee**: The system guarantees that citations in an answer are extracted from the retrieval result of that query, never generated from the LLM's general knowledge.

**Who must carry this?**
Something must own the construction of the prompt sent to the LLM. It must assemble context from retrieved chunks only, and extract citations from that same context. The LLM cannot be trusted to self-cite accurately.

**Responsibility**: The **Query Orchestrator** owns prompt construction and citation extraction. It passes only retrieved chunks as context to the LLM and maps the LLM's answer back to source metadata.

**Component**: `QueryOrchestrator`

---

### INV-05 — Re-processing resets a document completely

**Guarantee**: The system guarantees that when a document is re-processed, old and new state never coexist.

**Who must carry this?**
The same component that owns ingestion must own the purge before re-ingestion. Separating these two responsibilities would require coordination — and coordination is where invariants break.

**Responsibility**: The **Pipeline** owns re-processing. Before triggering ingestion, it instructs the Store to purge all existing chunks and embeddings for that document.

**Component**: `Pipeline` (same owner as INV-03)

---

### INV-06 — Workspace deletion is total and irreversible

**Guarantee**: The system guarantees that deleting a workspace removes all associated data atomically. No orphaned records survive.

**Who must carry this?**
Something must own workspace lifecycle — creation, switching, deletion — and know all the data that belongs to a workspace. This ownership must be centralized so deletion cascades correctly.

**Responsibility**: The **Workspace Manager** owns workspace lifecycle. It knows the full data footprint of a workspace and instructs the Store to delete everything in one atomic operation.

**Component**: `WorkspaceManager`

---

### INV-07 — Zero outbound network calls

**Guarantee**: The system guarantees that no component initiates a network connection to any external host during any operation.

**Who must carry this?**
This cannot be owned by a single component because it is a system-wide constraint. It must be enforced at the boundary of the process itself — not by trusting each component to behave.

**Responsibility**: The **Entry Point** enforces network isolation at startup. All calls to the LLM and embedding model go through local inter-process communication only. No component receives a network address as configuration.

**Component**: `Main` (process boundary enforcement)

---

### INV-08 — Conversation history is append-only

**Guarantee**: The system guarantees that conversation turns are written as they occur and never modified retroactively.

**Who must carry this?**
Something must own the persistence of conversation state — writing turns, reading past sessions, listing history. This is distinct from query logic.

**Responsibility**: The **Session Manager** owns conversation history. It appends turns to persistent storage immediately and exposes past sessions as read-only.

**Component**: `SessionManager`

---

### INV-09 — Retrieval scope is always explicit

**Guarantee**: The system guarantees there is no implicit global retrieval. Every search operation carries an explicit workspace boundary.

**Who must carry this?**
Already assigned. The Retriever owns this (see INV-02). It accepts no search call without a workspace ID.

**Component**: `Retriever` (same owner as INV-02)

---

## 3. Responsibility Grouping

Observing which responsibilities evolve together reveals the natural component boundaries.

| Component | Responsibilities owned | Invariants enforced |
|-----------|----------------------|---------------------|
| `Store` | All persistent reads and writes, referential integrity | INV-01 |
| `Retriever` | Scoped search, semantic + keyword retrieval | INV-02, INV-09 |
| `Pipeline` | Ingestion orchestration, transactional boundary, re-processing | INV-03, INV-05 |
| `QueryOrchestrator` | Prompt construction, context assembly, citation extraction | INV-04 |
| `WorkspaceManager` | Workspace lifecycle, data footprint, deletion cascade | INV-06 |
| `SessionManager` | Conversation history, session persistence, past session retrieval | INV-08 |
| `Main` | Process boundary, network isolation enforcement | INV-07 |

---

## 4. Coupling and Cohesion Check

**What changes together must live together.**
**What changes independently must stay apart.**

| Observation | Decision |
|-------------|----------|
| Pipeline and Store always interact during ingestion | Pipeline calls Store — they are coupled by design, not by accident |
| QueryOrchestrator and Retriever both serve a query | QueryOrchestrator calls Retriever — sequential dependency, not shared ownership |
| SessionManager and QueryOrchestrator both touch a query session | SessionManager persists; QueryOrchestrator reasons. Different rates of change — keep separate |
| WorkspaceManager and Store both touch workspace data | WorkspaceManager instructs Store — WorkspaceManager owns the decision, Store owns the execution |

No two components own the same guarantee. No guarantee is unowned.

---

## 5. What This Is Not

This document contains no:
- Programming language
- Library or framework name
- Database engine
- File format
- API design

Those decisions come in `05`. This document only answers:
**Who is responsible for what, and why.**

---

*Previous: [03 — Design: Contract & Invariants](./03_design_contract_invariants.md)*
*Next: [05 — Modeling: UML & C4](./05_modeling_uml_c4.md)*
