# VectorIndex Implementation

This package provides two implementations of the VectorIndex interface:

## 1. Pure Go Implementation (Default)

**File**: `index.go` (build tag: `!faiss`)

- Uses cosine similarity for vector search
- No external dependencies
- Suitable for development and testing
- Automatically used when FAISS is not available

**Usage**:
```bash
make test
make build
```

## 2. FAISS CGO Implementation (Production)

**File**: `index_faiss.go` (build tag: `faiss`)

- Uses Facebook's FAISS library via CGO
- High-performance ANN search
- Requires FAISS C library installed
- Recommended for production use

**Setup**: See `../../FAISS_SETUP.md` for installation instructions

**Usage**:
```bash
make test-faiss
make build-faiss
```

## Switching Between Implementations

The implementation is selected at compile time using Go build tags:

- **Default** (no tags): Pure Go implementation
- **With `-tags faiss`**: FAISS CGO implementation

Both implementations:
- Implement the same `Index` interface
- Use the same file structure and metadata format
- Are fully compatible (can switch between them)
- Enforce workspace isolation (INV-02, INV-09)

## File Structure

```
~/.docintel/workspaces/
└── <workspace_id>/
    ├── index.faiss      # FAISS index file (or JSON for pure Go)
    └── metadata.json    # Chunk ID to vector index mapping
```

## Testing

```bash
# Test with pure Go implementation
make test-vectorindex

# Test with FAISS (requires FAISS installed)
make test-faiss
```
