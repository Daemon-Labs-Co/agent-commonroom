# Security and architecture risk register

This register describes the current source boundaries and remaining risks. It
replaces the dated review transcripts and incremental remediation notes. Finding
IDs remain stable for references; they do not mean the original defect is still
present. This is a source assessment, not penetration-test certification, SOC 2
compliance or proof of live provider/container enforcement.

## Supported trust boundary

Use one controller, a private PostgreSQL runtime role and the supported Linux
rootless-Docker worker. HTTP remains loopback-only; access from another machine
uses the documented SSH tunnel. Public hosting, hostile multi-tenant isolation
and multiple concurrent controllers are not supported deployment guarantees.
Authenticated humans and approved workers remain distinct from model input.

Managed development admission remains blocked on production acceptance of the
credential-separated path. Internal Codex execution now assembles the audited
broker, fresh engine/tool pair, bootstrap, private credential delivery and provider
gateway; synthetic end-to-end tests exercise that assembly. This does not enable
public production dispatch. See [provider/tool separation](docs/provider-tool-separation.md).

## Security findings

| ID | Current protection | Remaining risk / completion evidence |
| --- | --- | --- |
| SEC-01 — Identity and authorization | Authenticated enabled-human API sessions, server-side room/owner checks and explicit actor attribution; agent API credentials rejected. | Capability-scoped agent/service identities and enterprise authorization are not implemented. Review every new route independently. |
| SEC-02 — Browser and listener boundary | Loopback listener, host/origin checks, authenticated sessions and mutation CSRF checks. | These do not provide public TLS deployment, reverse-proxy trust or hostile-tenant protection. |
| SEC-03 — Resource scope | Human room access and personal-room controls are enforced; owner personal-room reads are audited. | Project authorization and tenant isolation need distinct models; a company selector is not an isolation boundary. |
| SEC-04 — Execution and credential isolation | Mandatory rootless containers, legacy native development refusal, separate engine/tool implementation and audited read/write callbacks. | Reviewed image provenance, production admission and live hostile-workload/provider acceptance remain open. Containers do not defeat a compromised host/kernel. |
| SEC-05 — Audit integrity | Restricted runtime DB privileges, append-only trigger, hash chain, protected metadata and format-downgrade refusal. | A privileged DB administrator can rewrite or roll back a valid chain. Independent immutable storage and externally protected head/sequence anchors remain required. |
| SEC-06 — Incident shutdown | Integrity/storage failures latch execution off and cancel active contexts; explicit owner recovery rechecks integrity and worker state. | Cancellation alone is not physical shutdown proof. Transport/daemon failures leave jobs unresolved and must be reconciled. |
| SEC-07 — Untrusted browser content | Rendering and validation have regression coverage. | Every new HTML/attribute insertion remains a review boundary. Earlier review examples are not a current exploit claim. |
| SEC-08 — Native-session binding | Connection provenance, context-policy resets and recovery checks constrain continuation. | Provider changes, upgrades and imported session contents require compatibility review; an imported transcript is untrusted context. |
| SEC-09 — Worker transport and cancellation | Approved SSH destination, bounded protocols, durable lifecycle journals, creator fencing and exact-ID cleanup. | Live transport-loss/restart termination acceptance and complete production pair recovery remain open; SSH alone does not isolate execution. |
| SEC-10 — Resource exhaustion | Context/output/storage-wait budgets, turn admission, launch budgets, physical container reservations and gateway connection ceilings. | Hard aggregate 5 GB storage, paid limits, spend accounting, scalable replay and public volumetric defense remain unfinished. |
| SEC-11 — Attribution | New messages carry actor identity/type and cannot gain human styling from a name. | Legacy attribution may be inferred; comprehensive causal decision/evidence provenance remains incomplete. |
| SEC-12 — Context disclosure | Server-side destination filtering, source-room checks, human-only memory and explicit cross-room Application export grants. | Prompt injection is not eliminated. Project sharing, automatic secret detection and provider-side deletion are not established. |
| SEC-13 — Atomicity and replay | Transactional event batches, revisions and HTTP idempotency receipts protect supported operations. | No exactly-once external side effects or distributed controller ownership; ambiguous work requires explicit recovery. |
| SEC-14 — Deployment and supply chain | Separate migrations/runtime credentials, private-file checks, sanitized launch environments and immutable-image requirements. | Published reviewed images, repeatable image/provider/OS scans, upgrade compatibility and operational secret lifecycle still need acceptance. |

## Architecture findings

| ID | Current state and remaining direction |
| --- | --- |
| ARCH-01 — Durable execution | Durable worker intent, versioned journals and shutdown evidence exist. Controller dispatch ownership remains single-process; no distributed durable queue or automatic side-effect replay. |
| ARCH-02 — Identity model | Human membership and organizational hierarchy exist. Connections, room profiles and agent identity still require careful separation before enterprise roles or multiple companies are supported. |
| ARCH-03 — Module boundaries | Startup, persistence, projection, mutations and worker responsibilities are split into focused files. The executable remains one Go package; file extraction alone does not establish independent domain services. |
| ARCH-04 — Read scaling | Bounded storage contexts and history endpoints reduce unbounded waits. Whole-state replay remains a scaling concern; indexed projections/checkpoints need independent integrity and migration design. |
| ARCH-05 — Task lifecycle | Atomic task dispatch and correlated outcomes exist. Complete proposal, review, dependent-blocking and human-override semantics must follow the gold specification. |
| ARCH-06 — Adapter extensibility | Codex/Claude paths and an internal reviewed Codex callback bridge exist. Required isolated Gemini, Copilot and other provider paths remain unfinished or unverified; saved transport labels are not working adapters. |
| ARCH-07 — Workflows | Sequential pinned workflow execution and human gates exist. This is not a distributed/resumable swarm, general parallel join engine or paid consensus implementation. |
| ARCH-08 — Browser composition | Plain JavaScript scripts share globals and order-sensitive wrappers. Request scoping, draft preservation, accessibility and module boundaries need continued regression coverage. |
| ARCH-09 — Memory and provenance | Disclosure policy and source attribution constrain current memory. A real Project boundary remains absent. Paid Fulginn/Veriscribe and consensus are separate future scope. |
| ARCH-10 — Verification evidence | Human statements and provider-reported text must remain distinguishable from trusted observations. Bounded command execution, revision-bound artifacts and independent check results remain P0 work. |
| ARCH-11 — Operations | Explicit migrations, private startup state, worker diagnostics and fencing exist. Production provider/image acceptance, credential refresh, reviewed upgrades, idle policy and approved workspace disposal remain unfinished. |
| ARCH-12 — Tests and documentation | Synthetic authenticated, transaction, lifecycle and callback suites exist. Pending acceptance contracts are not completed tests. Automated tests mock DB/runtime/provider boundaries; manual live acceptance is separate. |

## Verification and response

Run `make test` during development and `make check` before completion. Use only
synthetic data, temporary filesystem fixtures and mocked database interactions;
do not create test schemas or launch real provider agents in automated suites.
[Remaining P0 acceptance contracts](docs/p0-acceptance-tests.md) identify pending
behavior and separate live checks. Passing a synthetic protocol peer does not
prove OAuth compatibility, image contents or kernel enforcement.

For an incident, stop new work, preserve audit/journal evidence and inspect exact
worker identities. Do not delete history, relax isolation or retry an ambiguous
side effect to make the system appear recovered. Follow [database recovery](docs/database-security.md),
[worker recovery](docs/worker-jobs.md) and the [security policy](SECURITY.md).
