# SRS — Software Requirements Specification
## Doc-Intel — Software Engineering Foundations

---

## 1. Actors

```
User ──────────────────────────────────────► Doc-Intel
       (workspace commands, document         (processes, indexes,
        management, natural language          reasons, responds)
        queries in EN or FR)
```

| Actor | Type | Role |
|-------|------|------|
| User | Human | The sole agent interacting with the system. Issues commands, adds documents, asks questions. |
| File System | External system | Source of PDF files. Read-only from the system's perspective during ingestion. |
| Local LLM | External binary | Called locally for text generation and reasoning. Never reaches the network. Pure function: text in, text out. No agency, no decision-making. |

No other actors exist. The system has no admin role, no API consumer, no remote client.

---

## 2. Functional Requirements

### 2.1 Workspace Management

**F-WS-01** — The system MUST allow the user to create a named workspace.

**F-WS-02** — The system MUST allow the user to list all existing workspaces.

**F-WS-03** — The system MUST allow the user to switch the active workspace.

**F-WS-04** — The system MUST allow the user to delete a workspace and ALL of its associated data: documents, chunks, embeddings, and conversation history.

**F-WS-05** — The system MUST isolate workspaces from each other. A query issued in workspace A MUST NOT return results originating from workspace B.

### 2.2 Document Management

**F-DM-01** — The system MUST allow the user to add a PDF file to the active workspace.

**F-DM-02** — The system MUST allow the user to list all documents in the active workspace, along with their processing status and read flag.

**F-DM-03** — The system MUST allow the user to delete a document from the active workspace. Deletion MUST remove all associated chunks, embeddings, and index entries atomically.

**F-DM-04** — The system MUST allow the user to re-process an existing document. Re-processing MUST purge all previous chunks and embeddings for that document before re-running the ingestion pipeline.

**F-DM-05** — The system MUST allow the user to mark a document as read or unread.

### 2.3 Ingestion Pipeline

**F-IN-01** — The system MUST extract text from text-native PDF pages directly from the text layer.

**F-IN-02** — The system MUST fall back to OCR for pages where no text layer is detected.

**F-IN-03** — The system MUST detect the primary language of a document (English or French) and store it as document metadata.

**F-IN-04** — The system MUST split extracted text into chunks. No chunk MUST exceed 512 tokens. Each chunk MUST overlap with the next by a minimum of 64 tokens.

**F-IN-05** — The system MUST generate a vector embedding for each chunk using a locally running embedding model.

**F-IN-06** — The system MUST persist all chunks and their embeddings atomically. If any step of the ingestion pipeline fails, the system MUST rollback all changes for that document and report the failure.

**F-IN-07** — The system MUST only trigger ingestion when explicitly commanded by the user. The system MUST NOT process documents in the background without user action.

### 2.4 Agent and Querying

**F-AG-01** — The system MUST allow the user to ask a natural language question scoped to the active workspace.

**F-AG-02** — The system MUST allow the user to scope a question to a single document within the active workspace.

**F-AG-03** — The system MUST retrieve relevant chunks using a combination of semantic similarity and keyword matching before passing context to the LLM.

**F-AG-04** — The system MUST NOT perform more than 5 retrieval operations per query before producing a final answer.

**F-AG-05** — The system MUST produce an answer that includes: the answer text, the source document filename, the source page number, and a confidence level (High / Medium / Low) for every response.

**F-AG-06** — The system MUST support cross-document reasoning. When a query requires information from multiple documents, the answer MUST reference all relevant sources.

**F-AG-07** — The system MUST support specific data extraction. When a user asks for a date, a name, a figure, or a clause, the system MUST return a targeted answer, not a general summary.

**F-AG-08** — The system MUST support document summarization. The summary MUST reference the source pages it draws from.

**F-AG-09** — The system MUST accept queries written in English or French, regardless of the language the source document is written in.

### 2.5 Conversation Memory

**F-MEM-01** — The system MUST maintain conversation context within a session. A follow-up question MUST resolve references from prior turns in the same session.

**F-MEM-02** — The system MUST persist conversation history per workspace across sessions.

**F-MEM-03** — The system MUST allow the user to list past sessions within the active workspace.

**F-MEM-04** — The system MUST allow the user to resume a past session and continue the conversation.

### 2.6 TUI

**F-TUI-01** — The system MUST stream the LLM answer token by token. The user MUST see text appear progressively, not all at once after a delay.

**F-TUI-02** — The system MUST display ingestion progress per document during processing.

**F-TUI-03** — The system MUST display workspace and document selection without requiring the user to exit and restart the tool.

---

## 3. Business Rules

**BR-01 — Workspace name uniqueness**
Two workspaces MUST NOT share the same name. The system MUST reject creation of a workspace whose name already exists.

**BR-02 — No orphaned data**
Deleting a workspace MUST cascade to all its documents, chunks, embeddings, and history. Deleting a document MUST cascade to all its chunks and embeddings. Partial states MUST NOT be persisted.

**BR-03 — Answer grounding**
An answer MUST NOT reference a source that was not part of the retrieval result for that query. The system MUST NOT fabricate citations.

**BR-04 — Processing state machine**
A document MUST exist in exactly one of three states at any time: `pending`, `indexed`, or `failed`. There is no intermediate state at rest. A document in `failed` state MUST be safe to re-process.

**BR-05 — Query isolation**
A query issued in workspace A MUST NOT access, read, or reason over documents belonging to workspace B, even if both workspaces contain documents with similar content.

**BR-06 — Explicit ingestion only**
The system MUST NOT modify the index, generate embeddings, or alter any stored state without an explicit user command.

---

## 4. Non-Functional Requirements

**NF-01 — Local execution**
The system MUST NOT make any outbound network connection during any operation: ingestion, querying, or session management. This is verifiable via `strace`.

**NF-02 — Single binary**
The system MUST compile and run as a single executable binary on Ubuntu 22.04+. No runtime interpreter, no dependency manager, no server process is required to run the tool.

**NF-03 — Ingestion performance**
The system MUST process a text-native PDF of 20 pages in under 10 seconds on the target hardware (Ryzen 3 CPU, 16GB RAM).

**NF-04 — Query latency**
The system MUST return a first token of the answer within 15 seconds of the user submitting a query on the target hardware.

**NF-05 — Memory ceiling**
The system MUST NOT consume more than 8GB of RAM during inference on the target hardware.

**NF-06 — Bilingual support**
The system MUST handle documents and queries written in English, French, or a mix of both without degraded answer quality.

---

## 5. Error Cases

**E-01 — PDF is unreadable**
*Condition*: The file at the given path is corrupt, password-protected, or not a valid PDF.
*Expected behavior*: The system MUST reject the file, set its status to `failed`, report a clear error message to the user, and leave the index unchanged.

**E-02 — OCR produces no text**
*Condition*: A page is an image but OCR returns empty or near-empty output.
*Expected behavior*: The system MUST store the page as an empty text entry, log a warning against that page number, and continue processing remaining pages.

**E-03 — Ingestion interrupted mid-document**
*Condition*: The process is killed or crashes during ingestion of a document.
*Expected behavior*: On next startup, the system MUST detect the document in an inconsistent state, reset it to `pending`, and notify the user that re-processing is required.

**E-04 — LLM produces no output**
*Condition*: The local LLM fails to return a response within a timeout period.
*Expected behavior*: The system MUST display a timeout error to the user, preserve the session context, and allow the user to retry the query.

**E-05 — Workspace not found**
*Condition*: The user attempts to switch to or operate on a workspace that does not exist.
*Expected behavior*: The system MUST display a clear error identifying the missing workspace name and list available workspaces.

**E-06 — Duplicate workspace name**
*Condition*: The user attempts to create a workspace with a name that already exists.
*Expected behavior*: The system MUST reject the creation and display the name conflict without modifying any existing workspace.

**E-07 — Query on empty workspace**
*Condition*: The user asks a question in a workspace that has no indexed documents.
*Expected behavior*: The system MUST inform the user that no documents are indexed in the current workspace and prompt them to add and process documents first.

---

*Previous: [01 — Requirements: PRD](./01_requirements_prd.md)*
*Next: [03 — Design: Contract & Invariants](./03_design_contract_invariants.md)*
