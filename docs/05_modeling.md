# Modeling — UML & C4
## Doc-Intel — Software Engineering Foundations

---

## Modeling Order

```
C4 Context  ->  UML Use Cases  ->  UML Sequence (all)  ->  C4 Container  ->  C4 Component
```

---

## 1. C4 — Level 1: System Context

Doc-Intel as a black box in its environment.

```
                    +---------------------------+
                    |     Local Machine         |
                    |                           |
  +-------+         |    +-------------+        |
  |       |         |    |             |        |
  | User  | ------> |    |  Doc-Intel  |        |
  |       | <------ |    |             |        |
  +-------+         |    +------+------+        |
  Terminal-native   |           |               |
  developer.        |    +------+------+        |
  EN / FR.          |    |      |      |        |
                    |    v      v      v        |
                    | +----+ +----+ +------+    |
                    | |LLM | |Emb | | File |    |
                    | |Bin | |Mod | |System|    |
                    | +----+ +----+ +------+    |
                    |  Local  Local   Local     |
                    |  no net no net  read-only |
                    +---------------------------+
```

---

## 2. UML — Use Case Diagram

```
                    +---------------------------------------------+
                    |                 Doc-Intel                   |
                    |                                             |
    +------+        |  (UC-01) Manage workspaces                 |
    |      |        |  (UC-02) Manage documents                  |
    | User | -----> |  (UC-03) Ingest documents                  |
    |      |        |  (UC-04) Ask a question (scoped)           |
    +------+        |  (UC-05) Summarize a document              |
                    |  (UC-06) Extract specific data             |
                    |  (UC-07) Compare across documents          |
                    |  (UC-08) Resume a past session             |
                    |                                             |
                    +---------------------------------------------+
```

---

## 3. UML — Sequence Diagrams

UML sequence notation used throughout:

```
  o            = actor (User)
 /|\
 / \

 +--------+   = object / component lifeline
 |        |
     |        = lifeline (vertical)
     |
  +--+--+     = activation box (component is active)
  |     |
  |     |
  +--+--+

 ------>       = synchronous call (solid arrow)
 - - - ->      = return message (dashed arrow)
 --x           = message lost / rejected
[condition]    = guard condition
```

---

### UC-01 — Manage Workspaces

```
    o
   /|\                +-----+          +---------+
   / \                | TUI |          |Workspace|
  User                +--+--+          | Manager |
    |                    |             +----+----+
    |                    |                  |
    | create <name>       |                  |
    | -----------------> |                  |
    |                    |  +------------+  |
    |                    |  |            |  |
    |                    +--+ Create(name)-->
    |                    |  |            |  |
    |                    |  +------------+  |
    |                    |                  |
    |                    | <- - - - - - - - |
    |                    |   ok / name      |
    |                    |   conflict error |
    | <- - - - - - - - - |                  |
    |   confirm / error  |                  |
    |                    |                  |
    | list               |                  |
    | -----------------> |                  |
    |                    +--+ List() ------->
    |                    | <- - - - - - - - |
    |                    |   []Workspace    |
    | <- - - - - - - - - |                  |
    |   workspace list   |                  |
    |                    |                  |
    | switch <name>      |                  |
    | -----------------> |                  |
    |                    +--+ Switch(name) -->
    |                    | <- - - - - - - - |
    |                    |   ok / not found |
    | <- - - - - - - - - |                  |
    |                    |                  |
    | delete <name>      |                  |
    | -----------------> |                  |
    |                    +--+ Delete(name) -->
    |                    |   [cascades all  |
    |                    |    data atomically]
    |                    | <- - - - - - - - |
    |                    |   ok             |
    | <- - - - - - - - - |                  |
```

---

### UC-02 — Manage Documents

```
    o
   /|\                +-----+          +--------+        +-----+
   / \                | TUI |          |  Doc   |        |Store|
  User                +--+--+          |Manager |        +--+--+
    |                    |             +----+---+           |
    |                    |                  |               |
    | add <path>          |                  |               |
    | -----------------> |                  |               |
    |                    +--+ Add(wsID,path)-->              |
    |                    |                  +-- Insert() --> |
    |                    |                  | <- - - - - - - |
    |                    |                  |  docID, status=pending
    |                    | <- - - - - - - - |               |
    |                    |   Document added |               |
    | <- - - - - - - - - |                  |               |
    |                    |                  |               |
    | list               |                  |               |
    | -----------------> |                  |               |
    |                    +--+ List(wsID) --->               |
    |                    |                  +-- Query() --> |
    |                    |                  | <- - - - - -  |
    |                    |                  |  []Document   |
    |                    | <- - - - - - - - |               |
    | <- - - - - - - - - |                  |               |
    |   doc list +status |                  |               |
    |                    |                  |               |
    | delete <docID>     |                  |               |
    | -----------------> |                  |               |
    |                    +--+ Delete(docID)->               |
    |                    |                  +-- DeleteAtomically() -->
    |                    |                  |   [chunks + vectors + record]
    |                    |                  | <- - - - - - - |
    |                    | <- - - - - - - - |               |
    | <- - - - - - - - - |                  |               |
    |                    |                  |               |
    | mark-read <docID>  |                  |               |
    | -----------------> |                  |               |
    |                    +--+ MarkRead(id) ->               |
    |                    |                  +-- Update() --> |
    |                    | <- - - - - - - - | <- - - - - -  |
    | <- - - - - - - - - |                  |               |
```

---

### UC-03 — Ingest Documents

```
    o
   /|\       +-----+      +--------+      +------+     +--------+      +-----+
   / \        | TUI |      |Pipeline|      |Parser|     |Embedder|      |Store|
  User        +--+--+      +---+----+      +--+---+     +---+----+      +--+--+
    |             |             |             |              |              |
    | ingest      |             |             |              |              |
    | ----------> |             |             |              |              |
    |             +-- Ingest(wsID) ---------->|              |              |
    |             |             |             |              |              |
    |             |   [for each document]     |              |              |
    |             |             |             |              |              |
    |             |             +-- Extract(path, lang) ---> |              |
    |             |             | <- - - - - -|              |              |
    |             |             |  []PageText |              |              |
    |             |             |             |              |              |
    |             |             +-- Chunk([]PageText)        |              |
    |             |             |   [<=512 tok, 64 overlap]  |              |
    |             |             |             |              |              |
    |             |             +-- Embed([]Chunk) --------> |              |
    |             |             | <- - - - - - - - - - - - - |              |
    |             |             |             []Vector        |              |
    |             |             |             |              |              |
    |             |             +-- BeginTransaction() --------------------> |
    |             |             +-- WriteChunks() --------------------------> |
    |             |             +-- WriteVectors() -------------------------> |
    |             |             +-- SetStatus(indexed) --------------------> |
    |             |             |             |              |              |
    |             |   [success] |             |              |              |
    |             |             +-- Commit() -----------------------------> |
    |             |             |             |              |              |
    |             |   [failure] |             |              |              |
    |             |             +-- Rollback() ---------------------------> |
    |             |             +-- SetStatus(failed) --------------------> |
    |             |             |             |              |              |
    |             | <- - - - -  |             |              |              |
    |             |  report     |             |              |              |
    | <- - - - -  |             |             |              |              |
    |  summary    |   [end for each]          |              |              |
```

---

### UC-04 — Ask a Question (Scoped)

```
    o
   /|\    +-----+    +---------+    +---------+    +-------+    +-------+
   / \     | TUI |    | Query   |    |Retriever|    |  LLM  |    |Session|
  User     +--+--+    | Orch.   |    +----+----+    +---+---+    |Manager|
    |          |      +----+----+         |              |        +---+---+
    |          |           |              |              |            |
    | question |           |              |              |            |
    | (wsID)   |           |              |              |            |
    | -------> |           |              |              |            |
    |          +-- GetContext(sessionID) -------------------------------->
    |          | <- - - - - - - - - - - - - - - - - - - - - - - - - - -|
    |          |           []PriorTurns   |              |            |
    |          |           |              |              |            |
    |          +-- Query(q, wsID, ctx) -> |              |            |
    |          |           |              |              |            |
    |          |  [up to 5 passes]        |              |            |
    |          |           |              |              |            |
    |          |           +-- Search(wsID, query) ----> |            |
    |          |           | <- - - - - - |              |            |
    |          |           |  []ScoredChunk              |            |
    |          |           |              |              |            |
    |          |           +-- ShouldContinue(q, chunks) -----------> |
    |          |           | <- - - - - - - - - - - - - |            |
    |          |           |   continue | answer         |            |
    |          |           |              |              |            |
    |          |  [answer] |              |              |            |
    |          |           +-- Generate(q, chunks) ----------------> |
    |          |           | <- - - - - - - - - - - - - |            |
    |          |           |   raw answer text           |            |
    |          |           |              |              |            |
    |          |           +-- ExtractCitations(answer, chunks)       |
    |          |           +-- ScoreConfidence(chunks)                |
    |          |           |              |              |            |
    |          | <- - - -  |              |              |            |
    |          |  Answer{text, citations, confidence}    |            |
    |          |           |              |              |            |
    |          +-- AppendTurn(sessionID, q, answer) ----------------->
    |          |           |              |              |            |
    | <- - - - |           |              |              |            |
    | streamed answer       |              |              |            |
    | + sources             |              |              |            |
    | + confidence          |              |              |            |
```

---

### UC-05 — Summarize a Document

```
    o
   /|\    +-----+    +---------+    +---------+    +-------+
   / \     | TUI |    | Query   |    |Retriever|    |  LLM  |
  User     +--+--+    | Orch.   |    +----+----+    +---+---+
    |          |      +----+----+         |              |
    |          |           |              |              |
    | summarize <docID>    |              |              |
    | -------> |           |              |              |
    |          +-- Summarize(docID, wsID)>|              |
    |          |           |              |              |
    |          |           +-- FetchAll(docID) --------> |
    |          |           | <- - - - - - |              |
    |          |           |  []Chunk (all pages)        |
    |          |           |              |              |
    |          |           +-- Summarize(chunks) ----------------->  |
    |          |           | <- - - - - - - - - - - - - |
    |          |           |   summary text + page refs  |
    |          |           |              |              |
    |          | <- - - -  |              |              |
    |          |  Summary{text, pageRefs} |              |
    | <- - - - |           |              |              |
    | summary + page refs  |              |              |
```

---

### UC-06 — Extract Specific Data

```
    o
   /|\    +-----+    +---------+    +---------+    +-------+
   / \     | TUI |    | Query   |    |Retriever|    |  LLM  |
  User     +--+--+    | Orch.   |    +----+----+    +---+---+
    |          |      +----+----+         |              |
    |          |           |              |              |
    | extract <type> <q>   |              |              |
    | -------> |           |              |              |
    |          +-- Extract(type, q, wsID)>|              |
    |          |           |              |              |
    |          |           +-- Search(wsID, q) --------> |
    |          |           | <- - - - - - |              |
    |          |           |  []ScoredChunk              |
    |          |           |              |              |
    |          |           +-- ExtractTyped(type, chunks)-----------> |
    |          |           | <- - - - - - - - - - - - - |
    |          |           |   structured result         |
    |          |           |   (date | name | figure)    |
    |          |           |              |              |
    |          | <- - - -  |              |              |
    |          |  ExtractedData{value, source, confidence}
    | <- - - - |           |              |              |
    | targeted result      |              |              |
```

---

### UC-07 — Compare Across Documents

```
    o
   /|\    +-----+    +---------+    +---------+    +-------+
   / \     | TUI |    | Query   |    |Retriever|    |  LLM  |
  User     +--+--+    | Orch.   |    +----+----+    +---+---+
    |          |      +----+----+         |              |
    |          |           |              |              |
    | compare <q> (wsID)   |              |              |
    | -------> |           |              |              |
    |          +-- Compare(q, wsID) ----> |              |
    |          |           |              |              |
    |          |           +-- SearchMultiDoc(wsID, q) ->|
    |          |           | <- - - - - - |              |
    |          |           |  []ScoredChunk (>=2 docs)   |
    |          |           |              |              |
    |          |           +-- Compare(chunks) ---------------------->|
    |          |           | <- - - - - - - - - - - - - |
    |          |           |   comparison text           |
    |          |           |   per-doc citations         |
    |          |           |              |              |
    |          | <- - - -  |              |              |
    |          |  Answer{text, citations per doc, confidence}
    | <- - - - |           |              |              |
    | comparison + sources |              |              |
    | per document         |              |              |
```

---

### UC-08 — Resume a Past Session

```
    o
   /|\    +-----+    +-------+    +-----+
   / \     | TUI |    |Session|    |Store|
  User     +--+--+    |Manager|    +--+--+
    |          |      +---+---+       |
    |          |          |           |
    | list sessions (wsID)|           |
    | -------> |          |           |
    |          +-- List(wsID) ------> |
    |          |          +-- Query()->|
    |          |          | <- - - - -|
    |          |          |  []Session|
    |          | <- - - - |           |
    | <- - - - |          |           |
    | session list        |           |
    |                     |           |
    | resume <sessionID>  |           |
    | -------> |          |           |
    |          +-- Load(sessionID) -> |
    |          |          +-- Query()->|
    |          |          | <- - - - -|
    |          |          |  []Turn   |
    |          | <- - - - |           |
    |          |          |           |
    | <- - - - |          |           |
    | session context     |           |
    | loaded in TUI       |           |
    |                     |           |
    | [continues as UC-04]|           |
```

---

## 4. C4 — Level 2: Container

The black box is opened one level.

```
+----------------------------------------------------------------------+
|                     Doc-Intel  (Single Binary)                       |
|                                                                      |
|   +-------------------------------------------------------------+    |
|   |                          TUI                               |    |
|   |              (only entry point for the User)               |    |
|   +----+----------------+---------------+-----------+----------+    |
|        |                |               |           |               |
|        v                v               v           v               |
|   +---------+    +----------+    +----------+  +---------+          |
|   |Workspace|    | Pipeline |    |  Query   |  | Session |          |
|   | Manager |    |          |    |  Orch.   |  | Manager |          |
|   +---------+    +----------+    +----------+  +---------+          |
|        |              |               |              |              |
|        |              |               |              |              |
|        +--------------+---------------+--------------+              |
|                                |                                    |
|                                v                                    |
|                    +-----------+----------+                         |
|                    |         Store        |                         |
|                    | workspaces, documents|                         |
|                    | chunks, history      |                         |
|                    +----------+-----------+                         |
|                               |                                     |
|                               v                                     |
|                    +--------------------+                           |
|                    |    Vector Index    |                           |
|                    |      (FAISS)       |                           |
|                    | ANN search over    |                           |
|                    | chunk embeddings   |                           |
|                    +--------------------+                           |
|                                                                      |
+----------------------------------------------------------------------+
          |                    |                    |
          v                    v                    v
    [ Local LLM ]    [ Embedding Model ]    [ File System ]
```

---

## 5. C4 — Level 3: Component (Query Orchestrator)

The only container whose internal structure is complex enough to warrant decomposition now.

```
+------------------------------------------------------------------+
|                      Query Orchestrator                          |
|                                                                  |
|   +--------------------+                                        |
|   |      Planner       |  decides: retrieve more or generate    |
|   |                    |  bounded to 5 retrieval passes         |
|   +--------+-----------+                                        |
|            |                                                     |
|            v                                                     |
|   +--------------------+                                        |
|   |  Context Builder   |  ranks and assembles retrieved chunks  |
|   |                    |  into ordered prompt context           |
|   +--------+-----------+                                        |
|            |                                                     |
|            v                                                     |
|   +--------------------+                                        |
|   |     Prompter       |  constructs final prompt               |
|   |                    |  with context + session history        |
|   +--------+-----------+                                        |
|            |                                                     |
|            v                                                     |
|   +--------------------+                                        |
|   | Citation Extractor |  maps answer back to source chunks     |
|   |                    |  never generates citations             |
|   +--------+-----------+                                        |
|            |                                                     |
|            v                                                     |
|   +--------------------+                                        |
|   | Confidence Scorer  |  High / Medium / Low                   |
|   |                    |  based on retrieval scores             |
|   +--------------------+                                        |
|                                                                  |
+------------------------------------------------------------------+
      |              |               |                |
      v              v               v                v
  [ Store ]   [ Vector Index ]   [ LLM ]    [ Session Manager ]
```

---

## 6. UML — Class Diagrams

Derived strictly from the objects exchanged in the sequence diagrams.
Every class here appeared first as a message parameter or return value in Section 3.
No class is invented — all are observed.

### Notation

```
  +------------------+
  |   ClassName      |   = class name
  +------------------+
  | - attribute: Type|   = private attribute
  | + attribute: Type|   = public attribute
  +------------------+
  | + method(): Type |   = public method
  | - method(): Type |   = private method
  +------------------+

  A ---------> B        = association (A knows B)
  A <>-------> B        = aggregation (A has B, B can exist alone)
  A <*>------> B        = composition (A owns B, B cannot exist without A)
  A <|-------- B        = inheritance (B extends A)

  [0..1]  = zero or one
  [1]     = exactly one
  [0..*]  = zero or many
  [1..*]  = one or many
```

---

### Core Domain Entities

These are the data objects that flow between components.

```
  +-------------------+          +----------------------+
  |    Workspace      |          |      Document        |
  +-------------------+          +----------------------+
  | - id: UUID        |          | - id: UUID           |
  | - name: String    |          | - workspaceID: UUID  |
  | - createdAt: Date |          | - path: String       |
  | - lastUsedAt: Date|          | - language: Language |
  +-------------------+          | - status: DocStatus  |
  | + isActive(): Bool|          | - pageCount: Int     |
  +-------------------+          | - isRead: Bool       |
          |                      | - createdAt: Date    |
          |                      | - processedAt: Date  |
          |  1                   +----------------------+
          |                      | + isPending(): Bool  |
          <*>                    | + isIndexed(): Bool  |
          |  0..*                +----------------------+
          |                               |
          |                               | 1
          |                              <*>
          |                               | 0..*
          |                       +---------------+
          |                       |    Chunk      |
          |                       +---------------+
          |                       | - id: UUID    |
          |                       | - docID: UUID |
          |                       | - pageNumber: Int
          |                       | - text: String|
          |                       | - tokenCount: Int
          |                       | - source: ChunkSource
          |                       +---------------+
          |                               |
          |                               | 1
          |                              <*>
          |                               | 1
          |                       +---------------+
          |                       |    Vector     |
          |                       +---------------+
          |                       | - chunkID: UUID
          |                       | - values: []Float
          |                       | - dimensions: Int
          |                       +---------------+
```

---

### Value Objects and Enumerations

```
  +-------------------+     +-------------------+     +-------------------+
  |    DocStatus      |     |   ChunkSource     |     |    Language       |
  |   <<enumeration>> |     |   <<enumeration>> |     |   <<enumeration>> |
  +-------------------+     +-------------------+     +-------------------+
  | PENDING           |     | TEXT_LAYER        |     | EN                |
  | INDEXED           |     | OCR               |     | FR                |
  | FAILED            |     +-------------------+     | UNKNOWN           |
  +-------------------+                               +-------------------+

  +-------------------+     +-------------------+
  |    Confidence     |     |    PageText       |
  |   <<enumeration>> |     |   <<value object>>|
  +-------------------+     +-------------------+
  | HIGH              |     | - pageNumber: Int |
  | MEDIUM            |     | - text: String    |
  | LOW               |     | - source: ChunkSource
  +-------------------+     +-------------------+
```

---

### Query and Answer Objects

```
  +----------------------+          +----------------------+
  |     QueryRequest     |          |       Answer         |
  +----------------------+          +----------------------+
  | - workspaceID: UUID  |          | - text: String       |
  | - documentID: UUID?  |          | - confidence: Confidence
  | - text: String       |          | - sources: []Citation|
  | - sessionID: UUID?   |          | - sessionID: UUID    |
  | - type: QueryType    |          +----------------------+
  +----------------------+          | + hasSources(): Bool |
                                    +----------------------+
  +-------------------+                      |
  |    QueryType      |                      | 1
  |   <<enumeration>> |             <*>------+
  +-------------------+             |  1..*
  | QUESTION          |    +------------------+
  | SUMMARY           |    |     Citation     |
  | EXTRACTION        |    +------------------+
  | COMPARISON        |    | - chunkID: UUID  |
  +-------------------+    | - docName: String|
                           | - pageNumber: Int|
                           | - excerpt: String|
                           +------------------+
```

---

### Session and History Objects

```
  +----------------------+          +----------------------+
  |       Session        |          |        Turn          |
  +----------------------+          +----------------------+
  | - id: UUID           |          | - id: UUID           |
  | - workspaceID: UUID  |          | - sessionID: UUID    |
  | - startedAt: Date    |          | - question: String   |
  | - lastActiveAt: Date |          | - answer: Answer     |
  | - title: String      |          | - createdAt: Date    |
  +----------------------+          +----------------------+
  | + isActive(): Bool   |
  +----------------------+
          | 1
         <*>
          | 0..*
  +----------------------+
  |        Turn          |
  +----------------------+
```

---

### Component Interfaces

Derived from the method calls observed in the sequence diagrams.
These are the contracts each component exposes — not implementations.

```
  +-----------------------------+     +-----------------------------+
  | <<interface>>               |     | <<interface>>               |
  | WorkspaceManager            |     | DocumentManager             |
  +-----------------------------+     +-----------------------------+
  | + Create(name): Workspace   |     | + Add(wsID, path): Document |
  | + List(): []Workspace       |     | + List(wsID): []Document    |
  | + Switch(name): Workspace   |     | + Delete(docID): void       |
  | + Delete(name): void        |     | + Reprocess(docID): void    |
  +-----------------------------+     | + MarkRead(docID, b): void  |
                                      +-----------------------------+

  +-----------------------------+     +-----------------------------+
  | <<interface>>               |     | <<interface>>               |
  | Pipeline                    |     | Retriever                   |
  +-----------------------------+     +-----------------------------+
  | + Ingest(wsID): Report      |     | + Search(wsID, q): []ScoredChunk
  | + Reprocess(docID): Report  |     | + SearchDoc(docID, q): []ScoredChunk
  +-----------------------------+     | + SearchMultiDoc(wsID, q): []ScoredChunk
                                      +-----------------------------+

  +-----------------------------+     +-----------------------------+
  | <<interface>>               |     | <<interface>>               |
  | QueryOrchestrator           |     | SessionManager              |
  +-----------------------------+     +-----------------------------+
  | + Query(req): Answer        |     | + GetContext(sID): []Turn   |
  | + Summarize(docID): Answer  |     | + List(wsID): []Session     |
  | + Extract(req): Answer      |     | + Load(sID): []Turn         |
  | + Compare(req): Answer      |     | + AppendTurn(sID, turn): void
  +-----------------------------+     +-----------------------------+

  +-----------------------------+
  | <<interface>>               |
  | Store                       |
  +-----------------------------+
  | + WriteDocument(doc): void  |
  | + WriteChunks([]Chunk): void|
  | + WriteVectors([]Vector): void
  | + SetStatus(id, s): void    |
  | + DeleteDocument(id): void  |
  | + QueryChunks(wsID): []Chunk|
  | + BeginTransaction(): Tx    |
  | + Commit(tx): void          |
  | + Rollback(tx): void        |
  +-----------------------------+
```

---

### Full Dependency Map

```
         +-----+
         | TUI |
         +--+--+
            |
     +------+------+----------+-----------+
     |              |          |           |
     v              v          v           v
+--------+   +----------+  +-------+  +---------+
|Workspace|   | Pipeline |  | Query |  | Session |
|Manager  |   |          |  | Orch. |  | Manager |
+----+----+   +----+-----+  +---+---+  +----+----+
     |              |           |           |
     |              |      +----+----+      |
     |              |      |         |      |
     |              |      v         v      |
     |              |  +--------+ +------+  |
     |              |  |Retriever| | LLM |  |
     |              |  +--------+ +------+  |
     |              |      |                |
     +--------------+------+----------------+
                           |
                           v
                      +--------+
                      | Store  |
                      +---+----+
                          |
                          v
                   +--------------+
                   | Vector Index |
                   +--------------+
```

---

## 7. Traceability

```
  Component             Responsibility (04)        Invariants
  ------------------    -----------------------    -----------------
  Store                 Data Store                 INV-01
  Vector Index          Data Store (search)        INV-01
  Pipeline              Pipeline                   INV-03, INV-05
  Query Orchestrator    Query Orchestrator         INV-04
  Workspace Manager     Workspace Manager          INV-06
  Session Manager       Session Manager            INV-08
  TUI                   Entry point                INV-07
```

No component exists without a traced responsibility.
No responsibility from `04` is unmodeled.

---

*Previous: [04 — Transition: Requirements to Architecture](./04_transition_req_to_arch.md)*
*Design phase complete. Implementation begins.*
