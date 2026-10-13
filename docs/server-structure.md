# Server structure

The executable remains one Go package, with focused files for distinct
responsibilities. Go uses concrete structs and interfaces rather than abstract
classes. Splitting files improves ownership and reviewability; it does not by
itself promise a shorter compile time.

| File | Responsibility and principal methods |
| --- | --- |
| `cmd/server/main.go` | `main`: configuration, dependency creation, routes, authentication initialization and HTTP startup. |
| `cmd/server/application.go` | Concrete `server` dependency/coordination struct. `db` uses the `eventDatabase` interface; `auth` is the owner authentication service; `mu` protects controller state; active/cancellation maps and worker seams coordinate managed execution. |
| `cmd/server/event_identity.go` | `id` generates cryptographic IDs; `checksum` advances the local audit chain. |
| `cmd/server/event_store.go` | `auditChange`, `append`, `appendBatch`: transaction boundaries, actor attribution, protected audit metadata, canonical JSON, chain linking and response receipts. |
| `cmd/server/state_projection.go` | `load`: replay immutable events into the current `room.State`, including internal job/operation projections and presentation diagnostics. |
| `cmd/server/http_state.go` | `reply`, `session`, `state`: JSON responses, authenticated session metadata and room-authorized state views. |
| `cmd/server/events_handler.go` | `eventRequest`, `events`: common permissions, CSRF, request limits, retry resolution, mutation preparation and atomic persistence. |
| `cmd/server/event_mutations_identity.go` | `prepareIdentityEvent`: owner/company/profile/feature configuration payload validation. |
| `cmd/server/event_mutations_rooms.go` | `prepareRoomsEvent`: room, workflow, task and evidence validation and revision contracts. |
| `cmd/server/event_mutations_agents.go` | `prepareAgentsEvent`: connection and participant validation, execution provenance preservation and revisions. |
| `cmd/server/event_mutations_content.go` | `prepareContentEvent`: messages, attention acknowledgments, memory disclosure and room export policy validation. |
| `cmd/server/integrity_handler.go` | `integrity`, `integrityLocked`: human inspection, audit verification and fail-closed execution state. |
| `cmd/server/database.go` | Abstract persistence contract `eventDatabase`; production uses `pgxpool.Pool`, tests use an in-memory transactional mock. |
| `cmd/server/atomic_operations.go` | Request fingerprinting, durable replay receipts and transactional stale-state context. |
| `cmd/server/context_disclosure.go` | Mandatory provider/source filtering and legacy native-context continuation rules. |

Mutation preparation methods return a typed payload and success indicator. On
rejection they write the existing HTTP error and return false; they do not append
events or invoke providers. The orchestration handler performs the commit only
after successful preparation. Export policies and native-session resets commit
together. Existing public routes, event kinds and authorization boundaries are
preserved by the structural refactor.

`load` still performs whole-history replay. The `server` struct still coordinates
one controller, and event validation helpers remain in the same package. Domain
services, incremental projections and distributed workflow ownership would be
further architectural work; this extraction does not claim to implement them.

Verification uses the existing behavioral suite for authenticated event handling,
permission failures, state replay, integrity, cancellation, lifecycle, revisions
and new disclosure/atomicity regressions, plus race/static/UI checks. No optional
product flag applies to this behavior-preserving structural cleanup.
