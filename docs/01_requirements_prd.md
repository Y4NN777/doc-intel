# PRD — Doc-Intel
## Software Engineering Foundations

---

## 1. Problem

A developer who works with private documents accumulates them across multiple projects and topics over time. Knowledge stays trapped inside isolated files. Finding a specific piece of information requires remembering which document contains it, opening it, and reading through it manually.

Existing local tools that solve the retrieval problem either send documents to external APIs, require a running server, depend on a heavy runtime, or expose a browser interface. None of them fit a developer who lives in the terminal and has privacy requirements.

Even in tools that do support local document chat, conversations are stateless. Each session starts from zero. A user who asked a question yesterday cannot resume that thread today.

> A developer opens a terminal to find information across their private documents.
> They have no tool that is simultaneously intelligent, offline, terminal-native,
> and organized around the way they think: in projects and topics.
> They resort to manual reading. Time is lost. Context is lost. Knowledge stays trapped.

---

## 2. Target Users

**Primary user: the terminal-native developer**

- Works on Linux, primarily in a terminal environment
- Manages personal and professional documents organized by project or topic
- Has privacy requirements that rule out any cloud-based tool
- Does not want to manage a running server or open a browser just to query a document
- Works in English, French, or both

**This is not a tool for:**
- Teams or multi-user environments
- Non-technical users expecting a graphical installer
- Users with no privacy concern about their documents

---

## 3. Use Cases

**UC-01 — Organize documents by project**
The user groups documents into isolated, named workspaces that reflect their natural organization.

**UC-02 — Add and manage documents**
The user adds documents to a workspace, removes them, re-processes them, and sees what is currently indexed.

**UC-03 — Process documents on demand**
The user triggers document processing explicitly — never automatically without their knowledge.

**UC-04 — Ask questions in natural language**
The user asks questions about document content in English or French and receives answers grounded in the actual content.

**UC-05 — Scope a question**
The user directs a question at a specific workspace or narrows it to a single document within that workspace.

**UC-06 — Receive sourced, confident answers**
Every answer includes the source document name, page number, and a confidence level.

**UC-07 — Extract specific data**
The user asks for specific data points — dates, names, figures — and receives targeted answers.

**UC-08 — Compare across documents**
The user asks questions that require synthesizing information from multiple documents in the same workspace.

**UC-09 — Summarize a document**
The user requests a summary of a single document and receives an overview with source page references.

**UC-10 — Track read status**
The user marks documents as read and sees which ones in a workspace have not been reviewed.

**UC-11 — Resume past conversations**
The user returns to a previous session within a workspace and continues a conversation from a prior session.

---

## 4. Out of Scope

| Excluded | Reason |
|----------|--------|
| Web or graphical UI | Contradicts the terminal-native constraint. Planned v2. |
| Multi-user support | Personal tool. Auth and access control are out of scope. |
| Cloud sync or remote access | Contradicts the local-only privacy constraint. |
| Non-PDF formats | Scope control. PDF is the primary use case. v2+. |
| Automatic background ingestion | The user controls when processing happens. |
| Internet search or external knowledge | All knowledge comes from the user's own documents only. |
| Model fine-tuning or training | Doc-Intel reasons over documents. It does not modify models. |

---

## 5. Success Criteria

**SC-01** — A user creates two workspaces. Documents and conversations from one are never visible in the other.

**SC-02** — A user removes a document. No trace of its content appears in any subsequent answer.

**SC-03** — A user asks a question. The answer references the correct filename and page number without fabricating content.

**SC-04** — A user asks a question requiring information from two documents. The answer correctly synthesizes both.

**SC-05** — At no point during any operation does the system make an outbound network connection.

**SC-06** — A user closes the tool, reopens it the next day, and finds their previous conversation available to continue.

**SC-07** — A user asks a question in French about an English document and receives a coherent, relevant answer.

**SC-08** — A 20-page text-native PDF is processed and a first answer is returned within 15 seconds on target hardware.

---

*Previous: [00 — Foundation: Engineering Mindset](./00_foundation_engineering_mindset.md)*
*Next: [02 — Requirements: SRS](./02_requirements_srs.md)*
