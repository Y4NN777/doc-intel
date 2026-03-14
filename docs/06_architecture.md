# Architecture — Software Design
## Doc-Intel — Software Engineering Foundations

---

## 1. Architectural Style

Doc-Intel follows a **Layered Pipeline Architecture** with an embedded
**Agent Loop** in the query layer.

This choice is justified by three observations from the design phase:

**1. Data flows in one direction.**
Documents enter the system, get processed through sequential stages
(extract → chunk → embed → store), and never flow backwards.
A pipeline is the natural model for this.

**2. Concerns are strictly separated by depth.**
The user interface knows nothing about storage.
Storage knows nothing about reasoning.
Reasoning knows nothing about rendering.
Layers communicate only with the layer directly below them.

**3. The query path is non-linear.**
Unlike ingestion, a query does not follow a fixed sequence of steps.
The system must decide how many retrieval passes to make, when context
is sufficient, and when to generate a final answer. This decision loop
is the definition of an agent. It is embedded inside the query layer
and does not affect the layered structure of the rest of the system.

---

## 2. System High-Level Architecture

```
  +------------------------------------------------------------------+
  |                        USER INTERFACE LAYER                      |
  |                                                                  |
  |    +----------------------------------------------------------+  |
  |    |                          TUI                            |  |
  |    |   Commands | Progress display | Answer streaming        |  |
  |    +----------------------------------------------------------+  |
  +---------------------------|--------------------------------------+
                              |
                              | (user commands flow down)
                              | (responses stream up)
                              |
  +------------------------------------------------------------------+
  |                        ORCHESTRATION LAYER                       |
  |                                                                  |
  |   +-------------------+          +----------------------------+  |
  |   |  Workspace Manager|          |     Document Manager       |  |
  |   |  create           |          |     add / delete           |  |
  |   |  switch           |          |     reprocess              |  |
  |   |  delete (cascade) |          |     mark read              |  |
  |   +-------------------+          +----------------------------+  |
  |                                                                  |
  +---------------------------|--------------------------------------+
                              |
                              | (orchestration delegates down)
                              |
       +----------------------+---------------------+
       |                                            |
       v                                            v
  +--------------------+                 +---------------------+
  |   INGESTION LAYER  |                 |    QUERY LAYER      |
  |                    |                 |                     |
  |  +-------------+   |                 |  +---------------+  |
  |  |   Parser    |   |                 |  |    Agent      |  |
  |  | text layer  |   |                 |  |    Loop       |  |
  |  | OCR fallback|   |                 |  |               |  |
  |  +------+------+   |                 |  | plan -> search|  |
  |         |          |                 |  | -> generate   |  |
  |  +------v------+   |                 |  | -> cite       |  |
  |  |   Chunker   |   |                 |  | -> score      |  |
  |  | semantic    |   |                 |  +-------+-------+  |
  |  | split +     |   |                 |          |          |
  |  | overlap     |   |                 |  +-------v-------+  |
  |  +------+------+   |                 |  |   Retriever   |  |
  |         |          |                 |  | semantic +    |  |
  |  +------v------+   |                 |  | keyword merge |  |
  |  |   Embedder  |   |                 |  +---------------+  |
  |  | local model |   |                 |                     |
  |  +------+------+   |                 +---------------------+
  |         |          |                           |
  |  +------v------+   |                           |
  |  |  Pipeline   |   |                           |
  |  | transaction |   |                           |
  |  | boundary    |   |                           |
  +-----|----------+---+                           |
        |                                          |
        +-------------------+-----------------------+
                            |
                            v
  +------------------------------------------------------------------+
  |                        PERSISTENCE LAYER                         |
  |                                                                  |
  |   +-------------------------+    +---------------------------+   |
  |   |         Store           |    |       Vector Index        |   |
  |   |        (SQLite)         |    |        (FAISS)            |   |
  |   |                         |    |                           |   |
  |   |  workspaces             |    |  ANN search               |   |
  |   |  documents              |    |  workspace-scoped         |   |
  |   |  chunks (text)          |    |  embeddings               |   |
  |   |  conversation history   |    |                           |   |
  |   |  session metadata       |    |                           |   |
  |   |                         |    |                           |   |
  |   |  transaction control    |    |                           |   |
  |   |  referential integrity  |    |                           |   |
  |   +-------------------------+    +---------------------------+   |
  |                                                                  |
  +------------------------------------------------------------------+
                            |
                            | (local IPC only — no network)
                            |
  +------------------------------------------------------------------+
  |                     EXTERNAL RUNTIME LAYER                       |
  |                                                                  |
  |   +---------------------+    +------------------------------+   |
  |   |    Local LLM        |    |     Embedding Model          |   |
  |   |                     |    |                              |   |
  |   |  quantized binary   |    |  local binary                |   |
  |   |  text generation    |    |  vector generation           |   |
  |   |  no network         |    |  no network                  |   |
  |   +---------------------+    +------------------------------+   |
  |                                                                  |
  |   +------------------------------+                              |
  |   |    File System               |                              |
  |   |  PDF source files            |                              |
  |   |  read-only during ingestion  |                              |
  |   +------------------------------+                              |
  |                                                                  |
  +------------------------------------------------------------------+
```

---

## 3. Layer Descriptions

### Layer 1 — User Interface

**Role**: The only entry point for the user. Renders output,
accepts input, streams answers.

**Rule**: This layer knows about all managers but no manager
knows about this layer. Communication is strictly top-down.

**Key concern**: Streaming. The user sees answer tokens
appear progressively. The TUI must handle a streaming
response from the query layer without blocking.

---

### Layer 2 — Orchestration

**Role**: Receives user commands and routes them to the
correct pipeline or query path. Owns workspace and
document lifecycle.

**Rule**: This layer does not process documents and does
not reason. It delegates. It is the traffic controller,
not the worker.

**Key concern**: Cascade correctness. When a workspace
or document is deleted, the orchestration layer must
ensure the persistence layer removes all associated
data atomically.

---

### Layer 3 — Ingestion (left path)

**Role**: Transforms raw PDF files into indexed, searchable
chunks. Runs only on explicit user command.

**Rule**: This path is strictly sequential. Each stage
receives the output of the previous stage. No stage
skips another. The pipeline wraps the entire sequence
in a transaction boundary — either all stages complete
for a document, or none are persisted.

```
  PDF file
     |
     v
  Parser          extract text layer
     |             fallback to OCR
     v             detect language
  Chunker         split into semantic units
     |             max 512 tokens, 64 overlap
     v
  Embedder        generate vectors
     |             local model, no network
     v
  Pipeline        commit atomically
     |             or rollback entirely
     v
  Store + VectorIndex
```

---

### Layer 3 — Query (right path)

**Role**: Accepts a natural language question, reasons over
the indexed corpus, and produces a grounded answer with
citations and confidence.

**Rule**: This path is non-linear. The agent loop decides
how many retrieval passes are needed before generating
an answer. The number of passes is bounded to 5.
Citations are resolved from retrieved chunks only —
never generated by the LLM.

```
  User question
       |
       v
  +--------------------+
  |    Agent Loop      |
  |                    |
  |  1. retrieve       | <---+
  |  2. evaluate       |     | (up to 5 passes)
  |     sufficient?    | ----+
  |  3. generate       |
  |  4. cite           |
  |  5. score          |
  +--------------------+
       |
       v
  Answer + sources + confidence
```

---

### Layer 4 — Persistence

**Role**: Owns all durable state. Enforces referential
integrity. Provides transaction control to the layers
above it.

**Two concerns, two components**:

```
  Store (SQLite)               Vector Index (FAISS)
  --------------               --------------------
  Structured data              Unstructured search
  workspaces, documents,       embeddings, ANN search
  chunks (text), history       workspace-scoped

  Transactional                Eventually consistent
  (all or nothing)             (insert + search)
```

**Rule**: No layer above may write to persistent state
without going through this layer. The Store is the
only component that holds transaction boundaries.

---

### Layer 5 — External Runtime

**Role**: Local binaries called by the layers above.
They are not part of Doc-Intel's codebase. They are
treated as pure functions — input in, output out.

**Rule**: All communication with this layer is local IPC.
No network address is ever configured. These binaries
have no network access during Doc-Intel's operation.

```
  Local LLM            text in  ->  text out
  Embedding Model      text in  ->  vector out
  File System          path in  ->  bytes out
```

---

## 4. The Two Data Flows

The architecture supports exactly two data flows.
All use cases are instances of one of these two flows.

### Flow A — Ingestion

```
  User command
      |
      v
  TUI -> WorkspaceManager/DocManager
      |
      v
  Pipeline
      |
      v
  Parser -> Chunker -> Embedder
      |
      v
  Store + VectorIndex
      |
      v
  Status report streamed back to TUI
```

### Flow B — Query

```
  User question
      |
      v
  TUI -> QueryOrchestrator
      |
      v
  Agent Loop
  [Retriever <-> Store/VectorIndex] (up to 5x)
      |
      v
  LLM (generate)
      |
      v
  CitationExtractor + ConfidenceScorer
      |
      v
  Answer streamed back to TUI
  Turn persisted to SessionManager -> Store
```

---

## 5. Architectural Constraints Enforced by Design

```
  Constraint             How the architecture enforces it
  ----------             --------------------------------
  No network             External Runtime layer has no network path.
                         All calls are local IPC. Enforced at entry.

  Single binary          All layers compile into one executable.
                         No server, no daemon, no separate process.

  Atomic ingestion       Pipeline layer owns the transaction boundary.
                         Store commits or rolls back as a unit.

  Workspace isolation    Retriever always receives wsID.
                         FAISS index namespaces by wsID.
                         No cross-workspace path exists.

  Append-only history    SessionManager has no update path.
                         Only AppendTurn() writes to history.

  Bounded agent loop     Agent Loop has a hard ceiling of 5 passes.
                         After 5, it generates from available context.
```

---

## 6. Project Structure

```
  doc-intel/
  |
  +-- main              Entry point. Wires all modules. Enforces INV-07.
  |
  +-- domain/           Shared entities. No dependencies. No logic.
  |                     Workspace, Document, Chunk, Vector, Session,
  |                     Turn, Answer, Citation, QueryRequest, ScoredChunk.
  |                     All enumerations: DocStatus, Language, Confidence,
  |                     QueryType, ChunkSource.
  |
  +-- tui/              USER INTERFACE LAYER
  |                     Renders interface. Streams answers. Dispatches
  |                     user commands to the correct module. Single
  |                     entry point — nothing bypasses it.
  |
  +-- workspace/        ORCHESTRATION LAYER
  |                     Workspace lifecycle only: create, list,
  |                     switch, delete. Delegates cascade delete
  |                     to store.
  |
  +-- docmanager/       ORCHESTRATION LAYER
  |                     Document lifecycle only: add, list, delete,
  |                     reprocess, mark read. Never touches pipeline.
  |
  +-- pipeline/         INGESTION LAYER
  |                     Owns the full ingestion sequence and its
  |                     transaction boundary. Contains parser,
  |                     chunker, embedder as internal components.
  |                     Nothing outside pipeline calls parser,
  |                     chunker, or embedder directly.
  |
  +-- retriever/        QUERY LAYER — search side
  |                     Hybrid search: semantic + keyword, merged
  |                     and ranked. Always scoped by wsID.
  |                     Never called without a workspace scope.
  |
  +-- query/            QUERY LAYER — reasoning side
  |                     Agent loop: plan, retrieve, generate, cite,
  |                     score. Contains all sub-components of the
  |                     reasoning process. The only module that
  |                     calls the LLM.
  |
  +-- session/          QUERY LAYER — memory side
  |                     Conversation history: persist, list, load,
  |                     resume. Append-only. No update path.
  |
  +-- store/            PERSISTENCE LAYER — structured (SQLite)
  |                     All reads and writes for workspaces,
  |                     documents, chunks, and history.
  |                     Owns all transaction control.
  |                     Single writer — no module writes
  |                     persistent state without going through here.
  |
  +-- vectorindex/      PERSISTENCE LAYER — search (FAISS)
                        ANN search over chunk embeddings.
                        Namespaced by wsID.
                        Insert, search, delete.
```

---

## 7. Module Dependency Rules

```
  domain/        no dependencies
  store/         domain/
  vectorindex/   domain/
  pipeline/      domain/, store/, vectorindex/
  retriever/     domain/, store/, vectorindex/
  session/       domain/, store/
  workspace/     domain/, store/
  docmanager/    domain/, store/
  query/         domain/, retriever/, session/, store/
  tui/           workspace/, docmanager/, pipeline/, query/, session/
  main           tui/
```

**The rules stated simply:**
- `domain/` is the foundation — no one is below it
- `store/` and `vectorindex/` know nothing above themselves
- Data flows downward. No module calls a module above its layer.
- `query/` is the only module that calls the LLM
- `pipeline/` is the only module that calls the embedder

---

## 8. Invariant Enforcement Map

```
  Invariant   Module          What enforces it
  ---------   -------------   ----------------------------------------
  INV-01      store           Atomic cascade delete for doc + workspace
  INV-02      retriever       wsID is a required parameter, no default
  INV-03      pipeline        Transaction wraps the full ingest sequence
  INV-04      query           Citations resolved from retrieved chunks only
  INV-05      pipeline        Purge runs before new data is written
  INV-06      workspace       Delete delegates to store atomic cascade
  INV-07      main            Network block enforced at process start
  INV-08      session         No update path exists — append only
  INV-09      retriever       No search call accepted without wsID
```

**Technology Choices:**
- Store: SQLite (transactional structured data)
- Vector Index: FAISS (high-performance ANN search)

---

*Previous: [05 — Modeling: UML & C4](./05_modeling_uml_c4.md)*
*Next: Implementation.*
