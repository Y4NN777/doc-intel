# Foundation — Engineering Mindset
## Doc-Intel — Software Engineering Foundations

---

## Objective

This document establishes the **thinking framework** that governs every decision in this project.
It is not a technical document. It is a **mental contract** between the engineer and the problem.

Before writing a single requirement, a single schema, or a single line of code, the engineer must answer:

> **What dysfunction exists today that this system must eliminate?**

If this answer is vague, the architecture will be wrong.
If the architecture is wrong, the code is waste.

---

## 1. The Distinction That Matters

> **A Developer solves the problem they were given.**
> **An Engineer defines the problem before solving it.**

| Developer Mindset | Engineer Mindset |
|-------------------|-----------------|
| "What should I build?" | "What problem am I solving?" |
| Reaches for a framework | Defines the contract first |
| Adds features until it feels complete | Removes everything that doesn't serve a requirement |
| Treats privacy as a toggle | Treats privacy as an architectural constraint |
| Builds what is interesting | Builds what is necessary |

This project is built with the **engineer mindset**.
Every module has a defined responsibility.
Every responsibility maps to a requirement.
Every requirement is verifiable.

---

## 2. The Real Problem

### What this project is NOT solving

- "How to read PDFs" → a PDF viewer already exists.
- "How to search files" → grep already exists.
- "How to run an LLM locally" → llama.cpp already exists.
- "How to chat with PDFs locally" → AnythingLLM, PrivateGPT, GPT4All LocalDocs, Open WebUI all exist.

### Why those tools don't fit

Existing local PDF chat tools solve the problem for a different user profile:

| Tool | Stack | Interface | Deployment | Workspaces |
|------|-------|-----------|------------|------------|
| AnythingLLM | Node + Electron | Browser UI | Local server | Yes, but web-only |
| PrivateGPT | Python | Browser UI | Local server | No |
| GPT4All LocalDocs | C++ + Qt | Desktop GUI | Native app | No |
| Open WebUI | Python | Browser UI | Docker | No |

None of these fit a user who:
- Lives in the terminal and wants a **native TUI**, not a browser tab
- Wants a **single compiled binary** with zero runtime dependencies (no Node, no Python, no Docker)
- Needs **workspace-scoped persistent session history** with a clean engineering design
- Is building on **Ubuntu with Go** and wants to own and understand every layer of the stack

### What this project IS solving

> A developer who lives in the terminal accumulates private documents across different
> projects and topics. Existing local tools require a browser, a running server, or a
> heavy runtime. There is no tool that is simultaneously intelligent, fully local,
> terminal-native, and built as a single Go binary organized around workspaces
> with persistent conversation history.

This is the problem. Every feature in this system must trace back to it.
If a feature does not reduce the cost of reasoning over personal documents, it does not belong in v1.

---

## 3. The Non-Negotiable Constraint

Privacy in this system is **not a feature**. It is the reason the system exists in the form it does.

### Real-World Analogy

> **A safe vs. a filing cabinet**
>
> A filing cabinet organizes documents. A safe organizes documents *and* guarantees
> no one else can access them. If you need a safe, building a filing cabinet with a
> lock added later is the wrong approach. The constraint must drive the architecture
> from the start.

Doc-Intel is a safe, not a filing cabinet with a lock bolted on.

This means:
- Local-only is the **architecture**, not a deployment option.
- There is no "add cloud sync later" — that is a different product with a different threat model.
- Zero outbound network calls is an **invariant**, not a best practice.

---

## 4. What "Done" Looks Like

This system is done when a user can:

1. Create a workspace for a topic or project.
2. Add their own private PDF documents to it.
3. Process those documents with a single command.
4. Ask questions in natural language — English or French.
5. Receive grounded answers with source references and a confidence level.
6. Resume that conversation in a future session.

All of this without a single byte of their documents leaving their machine.

---

## 5. Chain of Causality

This project follows a strict chain. Skipping any step is not a shortcut — it is an error.

```
Intent → Requirement → Contract → Responsibility → Architecture → Code
```

| Step | Document | Question it answers |
|------|----------|-------------------|
| Intent | This document (00) | Why must this system exist? |
| Requirement | PRD (01) | What problem does it solve, for whom? |
| Specification | SRS (02) | What must the system do, verifiably? |
| Contract | Invariants (03) | What does the system guarantee and forbid? |
| Transition | Req → Arch (04) | Which module owns which guarantee? |
| Modeling | UML & C4 (05) | How do the responsibilities connect visually? |

---

## 6. Validation

This document is valid if:

1. Two engineers reading it can propose **completely different technical implementations** and both would correctly solve the stated problem.
2. If only one implementation seems possible after reading this, the document has leaked solution details and must be revised.

---

*Next: [01 — Requirements: PRD](./01_requirements_prd.md)*
