# Phase 6: Advanced Memory

**Goal:** Transform basic memory into intelligent, connected knowledge graph.
**Duration:** 3 weeks
**Dependencies:** Phase 3 (Memory & Learning)
**Status:** 📋 Planned

---

## 6.1 Memory Graph

### 6.1.1 Graph Data Structure
- [ ] Create `internal/memory/graph.go` — node/edge data structure
- [ ] Node types: entity, event, pattern, preference, fact
- [ ] Edge types: updates, extends, derives, related
- [ ] Graph traversal: BFS, DFS, shortest path
- [ ] SQLite storage for graph (nodes + edges tables)
- [ ] API: AddNode, AddEdge, QueryNeighbors, FindPath

### 6.1.2 Node Types
- [ ] EntityNode — people, places, things, concepts
- [ ] EventNode — actions with timestamps and context
- [ ] PatternNode — repeated behaviors or sequences
- [ ] PreferenceNode — user preferences and settings
- [ ] FactNode — extracted facts from conversations

### 6.1.3 Edge Types
- [ ] UpdatesEdge — node A updates node B
- [ ] ExtendsEdge — node A extends node B
- [ ] DerivesEdge — node A derives from node B
- [ ] RelatedEdge — node A is related to node B
- [ ] CausalEdge — node A causes node B

### 6.1.4 Graph Operations
- [ ] AddNode(node) — add new node to graph
- [ ] AddEdge(from, to, type) — create relationship
- [ ] QueryNeighbors(nodeID, edgeType) — find connected nodes
- [ ] FindPath(from, to) — shortest path between nodes
- [ ] Subgraph(nodeIDs) — extract subgraph
- [ ] Serialize() / Deserialize() — save/load graph

---

## 6.2 6-Category Extraction

### 6.2.1 Profile Extraction
- [ ] Extract user info (name, role, expertise)
- [ ] Extract preferences (tools, languages, frameworks)
- [ ] Extract working style (TDD, documentation, etc.)
- [ ] Store in ProfileNode with confidence score

### 6.2.2 Entity Extraction
- [ ] Extract people, places, organizations
- [ ] Extract technical entities (repos, services, APIs)
- [ ] Link entities to existing graph nodes
- [ ] Deduplicate similar entities

### 6.2.3 Event Extraction
- [ ] Extract actions with timestamps
- [ ] Link events to entities involved
- [ ] Extract outcomes (success, failure, partial)
- [ ] Store event context (what, where, when)

### 6.2.4 Case Extraction
- [ ] Extract problem-solution pairs
- [ ] Link cases to relevant entities
- [ ] Extract conditions for solution applicability
- [ ] Store case effectiveness metrics

### 6.2.5 Pattern Extraction
- [ ] Detect repeated behaviors across sessions
- [ ] Extract common workflows
- [ ] Identify anti-patterns to avoid
- [ ] Store pattern frequency and confidence

---

## 6.3 Memory Archival

### 6.3.1 Cold Memory Detection
- [ ] Track last access time per node
- [ ] Define cold threshold (default 30 days)
- [ ] Score nodes by relevance (access frequency, connections)
- [ ] Identify candidates for archival

### 6.3.2 Archive Process
- [ ] Compress node data to JSON
- [ ] Store in archive directory with date partitioning
- [ ] Maintain index for fast lookup
- [ ] Update graph to mark nodes as archived

### 6.3.3 Restore Process
- [ ] Search archive by query
- [ ] Decompress and restore nodes to active graph
- [ ] Rebuild edge connections
- [ ] Update access timestamps

### 6.3.4 Memory TTL
- [ ] Configurable TTL per node type
- [ ] Auto-delete expired nodes
- [ ] Warning before deletion
- [ ] Export before deletion option

---

## Phase 6 Checklist

| Category | Tasks | Done | Progress |
|----------|-------|------|----------|
| Memory Graph | 20 | 0 | 0% |
| 6-Category Extraction | 20 | 0 | 0% |
| Memory Archival | 12 | 0 | 0% |
| **Total** | **52** | **0** | **0%** |
