# Contributing to Doc-Intel

Thanks for your interest in contributing to Doc-Intel.

## Engineering Philosophy

This project follows a strict engineering discipline:

> Define the problem before solving it.

Before contributing code, read the engineering docs in `docintel/docs/` to understand:
- The problem we're solving (and what we're NOT solving)
- The system invariants that must never be violated
- The architectural constraints that drive design decisions

## System Invariants

These rules are non-negotiable. Any contribution that violates them will be rejected:

1. Zero outbound network calls — ever
2. Workspace isolation — no cross-workspace data leakage
3. Atomic operations — no partial states at rest
4. Grounded answers only — citations must come from retrieved content
5. Explicit control — no background processing without user command

## Before You Start

1. Read `docintel/docs/00_foundation_engineering_mindset.md`
2. Review the requirements in `01_requirements_prd.md` and `02_requirements_srs.md`
3. Understand the system contract in `03_design_contract_invariants.md`
4. Check the architecture in `06_architecture.md`

## How to Contribute

### Reporting Issues

- Check existing issues first
- Provide clear reproduction steps
- Include your environment (OS, Go version, hardware)
- For bugs, explain what you expected vs what happened

### Proposing Features

Before opening a feature request, ask:
- Does this reduce the cost of reasoning over personal documents?
- Does it violate any system invariant?
- Does it require network access or cloud services?

If the answer to the last two is yes, it doesn't belong in v1.

### Code Contributions

1. Fork the repository
2. Create a feature branch from `main`
3. Write your code following the module structure in `06_architecture.md`
4. Ensure your changes don't violate any invariant
5. Add tests that verify correctness
6. Submit a pull request with a clear description

### Code Standards

- Follow Go conventions and idioms
- Keep modules decoupled — respect the dependency rules in `06_architecture.md`
- Every component must have a single, clear responsibility
- No component should know about layers above it
- Write tests that verify invariants, not just happy paths

### Testing

- Unit tests for individual components
- Integration tests for data flows
- Property-based tests for invariants where applicable
- Performance tests for ingestion and query latency

## Module Ownership

Each module has a clear responsibility. Contributions must respect these boundaries:

- `domain/` — Shared entities only, no logic
- `store/` — All persistent writes, referential integrity
- `pipeline/` — Ingestion orchestration, transaction boundaries
- `retriever/` — Scoped search operations
- `query/` — Query orchestration, citation extraction
- `workspace/` — Workspace lifecycle
- `session/` — Conversation history
- `tui/` — User interface only

See `06_architecture.md` for detailed module responsibilities.

## What We're Looking For

- Bug fixes that maintain invariants
- Performance improvements
- Better error messages
- Test coverage improvements
- Documentation clarifications

## What We're NOT Looking For (v1)

- Web UI or REST API
- Cloud sync or remote features
- Multi-user support
- Non-PDF format support
- Automatic background processing

These may be considered for v2+.

## Questions?

Open an issue for discussion before starting major work.

---

**Remember**: Every line of code must trace back to a requirement. If it doesn't reduce the cost of reasoning over personal documents, it doesn't belong here.
