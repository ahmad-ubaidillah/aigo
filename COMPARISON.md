# Aigo vs Hermes — Comparison & Roadmap

## 📊 Feature Matrix (v0.2.0)

| Feature | Hermes (Python) | Aigo (Go) | Gap |
|---------|-----------------|-----------|-----|
| **Core Engine** | | | |
| Agent loop | ✅ ReAct | ✅ ReAct | = |
| Loop detection | ✅ FNV-1a | ✅ FNV-1a | = |
| Tool registry | ✅ 20+ tools | ✅ 14 tools | 🟡 Aigo needs more |
| Streaming | ✅ SSE | ✅ SSE | = |
| Multi-provider | ✅ 5+ | ✅ OpenAI-compat | = |
| Provider failover | ✅ Account pool | ✅ Basic | 🟡 |
| Prompt caching | ✅ (Nous models) | ❌ | 🔴 |
| **Channels** | | | |
| Telegram | ✅ | ✅ | = |
| Discord | ✅ | ✅ | = |
| Slack | ❌ | ✅ | 🟢 Aigo wins |
| WebSocket | ✅ | ✅ | = |
| WhatsApp | ✅ (baileys) | ❌ | 🔴 |
| Signal | ✅ | ❌ | 🔴 |
| **Memory** | | | |
| Daily notes | ✅ Markdown | ✅ FTS5/Markdown | 🟢 Aigo wins |
| Long-term | ✅ Markdown | ✅ FTS5/Markdown | 🟢 Aigo wins |
| Session storage | ✅ JSONL | ✅ FTS5 | 🟢 Aigo wins |
| Full-text search | ✅ Basic grep | ✅ FTS5 BM25 | 🟢 Aigo wins |
| Auto-memory | ❌ | ✅ session pkg | 🟢 Aigo wins |
| Auto-learning | ❌ | ✅ Correction detect | 🟢 Aigo wins |
| Vector search | ✅ (zvec) | ✅ (sqlite-vec SimHash) | 🟡 |
| **Skills** | | | |
| Skill system | ✅ 1100+ skills | ❌ | 🔴 |
| Skill install | ✅ Hub + LobeHub | ❌ | 🔴 |
| Skill search | ✅ hermes skills search | ❌ | 🔴 |
| **Web UI** | | | |
| Dashboard | ✅ hermes-chat | ✅ Built-in | = |
|| Chat interface | ✅ SSE streaming | ✅ SSE streaming | = ||
| Settings editor | ✅ | ✅ | = |
| Mobile responsive | 🟡 Basic | ✅ Mobile-first | 🟢 Aigo wins |
| Typing indicator | ✅ | ✅ | = |
| Copy button | ❌ | ✅ | 🟢 Aigo wins |
| **Advanced** | | | |
| MCP client | ✅ | ✅ | = |
| Cron scheduler | ✅ | ✅ | = |
| Web search | ✅ (multiple) | ✅ DuckDuckGo | 🟡 |
| Web fetch | ✅ (crawl4ai) | ✅ HTML→text | 🟡 |
| Context compression | ✅ | ✅ | = |
| Semantic routing | ✅ (98% accuracy) | ✅ (TF-IDF) | 🟡 |
| Sub-agent delegation | ✅ | ✅ (Sisyphus/Hephaestus/Oracle/Explore) | 🟡 |
| PLUR learning | ✅ (engrams) | ❌ | 🔴 |
| Session search | ✅ (zvec) | 🟡 (FTS5) | 🟡 |
| **Performance** | | | |
| Binary size | N/A (Python) | 12MB static | 🟢 Aigo wins |
| Startup time | ~2-5s | <100ms | 🟢 Aigo wins |
| Memory usage | ~50-100MB | ~20-30MB | 🟢 Aigo wins |
| External deps | Python, venv | Zero | 🟢 Aigo wins |
| **Platforms** | | | |
| VPS/Server | ✅ | ✅ | = |
| Embedded/IoT | ❌ | 🟡 (Aizen planned) | 🟢 |
| Raspberry Pi | ❌ | 🟡 (Aizen planned) | 🟢 |

## 🎯 What Aigo Needs to Match Hermes

### Priority 1 (Must Have)
1. **Streaming SSE in WebUI** — Hermes has real-time token streaming
2. **More channels** — WhatsApp (baileys port to Go), Signal
3. ~~**Semantic router** — intent-based routing like Hermes (98% accuracy)**~~ ✅ Implemented (TF-IDF)
4. ~~**Skill system** — load/manage/install skills like Hermes (1100+ available)**~~ 🟡 Basic (needs more skills)
5. ~~**Sub-agent delegation** — spawn child agents for complex tasks**~~ ✅ Implemented

### Priority 2 (Nice to Have)
6. **Vector search** — integrate a lightweight vector DB (hnswlib/brute-force)
7. **Prompt caching** — when provider supports it (Nous, Anthropic)
8. **PLUR-like learning** — cross-domain principles extraction
9. **MCP server mode** — expose Aigo as MCP server for other agents

### Priority 3 (Future)
10. **Aizen (Zig)** — embedded version for IoT/edge
11. **Plugin system** — extensible plugin architecture
12. **Multi-agent orchestration** — coordinate multiple agents

## 🏆 Aigo's Advantages Over Hermes

1. **Single binary, zero deps** — `scp aigo server && ./aigo start`
2. **100x faster startup** — <100ms vs 2-5s
3. **Auto-memory** — learns from conversations automatically
4. **Auto-learning** — detects corrections and preferences
5. **FTS5 memory** — proper full-text search with ranking
6. **Better mobile UI** — mobile-first responsive design
7. **Slack support** — which Hermes doesn't have
8. **Copy button** — which Hermes doesn't have
9. **Cross-compile** — build for ARM/RISC-V without deps

## 🛣️ Implementation Plan

### Phase 1: Core Parity (Current → 2 weeks)
- [ ] WebSocket SSE streaming in WebUI
- [ ] WhatsApp channel (port baileys patterns to Go)
- [ ] Skill system (SKILL.md loader)
- [ ] Semantic router (TF-IDF or embedding-based)

### Phase 2: Advanced Features (2-4 weeks)
- [ ] Sub-agent delegation
- [ ] Vector memory (lightweight)
- [ ] MCP server mode
- [ ] Provider failover with account pool

### Phase 3: Aizen (4-8 weeks)
- [ ] Zig embedded version
- [ ] <1MB binary target
- [ ] IoT/edge deployment
- [ ] Raspberry Pi support

## 📈 Token Efficiency Comparison

| Operation | Hermes (tokens) | Aigo (tokens) |
|-----------|-----------------|---------------|
| System prompt | ~500-1000 | ~400-600 |
| Memory injection | ~200-500 | ~100-300 (FTS5) |
| Tool schemas (14) | ~800 | ~600 |
| Session context | ~300-1000 | ~200-500 |
| **Total per turn** | **~1800-3300** | **~1300-2000** |

Aigo uses ~30-40% fewer tokens due to:
- More compact system prompt
- FTS5 ranking (only relevant results)
- Auto-memory (no need to search every time)
- Efficient tool schema format
