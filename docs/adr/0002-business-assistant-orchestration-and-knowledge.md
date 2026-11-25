---
status: accepted
date: 2026-08-17
---

# Keep business-assistant orchestration in Go and treat LangGraph as an adapter

The Business Assistant will use an in-process Go Assistant Orchestrator with a Planner seam, Tool Registry seam, and deterministic synthesis. The default Rule Planner may plan at most three read-only Business Tool calls and every call is checked independently against its registered Casbin permission path. Dynamic schedules, announcements, Assets, and Invoices are always queried in real time through Business Tools; model output is non-authoritative wording and is discarded when it does not cover every successful Tool result.

Non-structured policies, handbooks, contracts, and meeting notes may enter the Knowledge Source index only with explicit Tenant, Department, User, and Role ownership. Existing documents without that ownership are not automatically indexed. This follows ADR-0001 and prevents a retrieval path from bypassing Data Scope.

LangGraph is not introduced into the default runtime. A LangGraph Planner adapter may call an external Python or JavaScript graph runtime later, but its output must pass through the same Tool Registry validation, three-Tool limit, authorization, observation, and deterministic synthesis. Adoption requires the external adapter to outperform the Rule Planner on the committed evaluation set while meeting deployment, latency, and failure-isolation targets.

## Consequences

- The Go process remains the authoritative business orchestration runtime.
- Rule, model, and LangGraph Planner adapters share one Planner interface.
- Tool schemas, read-only guarantees, and permission paths live in the Tool Registry.
- Provider failure or incomplete model wording cannot remove deterministic facts.
- Knowledge retrieval starts with PostgreSQL full-text search and a SQLite-compatible test adapter; embeddings can be added behind a retrieval seam when justified by evaluation.
- Historical documents remain excluded until ownership migration is implemented.
