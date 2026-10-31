# Product functionality and edition inventory

## Input

The owner's latest scope decisions, FUNCTIONAL_SPEC.md, implemented core paths,
and read-only inspection of commonroom-commercial-services main
2f47977a144249090d0f87dabf179ce1e52a50b6. Latest human decisions override older
pricing/backlog text. Tier and implementation status are separate properties.
“Partial” means some code exists, not that the production workflow is working.

## Output

### Free core — confirmed scope

| Functionality | Current implementation / integration state |
| --- | --- |
| One human owner and one project | Owner authentication exists; a real project boundary still needs definition/integration. |
| Self-hosted controller/browser | Implemented prototype; restricted PostgreSQL runtime role and separate migration path. |
| Multiple agents and independent sessions/profiles | Implemented controls; supported runtime varies by provider. |
| Provider CLI/subscription authentication | Codex/Claude paths exist; intended Gemini, Copilot and llama.cpp coverage remains incomplete. No silent substitution with separately billed APIs. |
| One approved local/remote Linux worker | Rootless-Docker foundations exist; live production assembly incomplete. Remote transport does not authorize host model execution. |
| Room conversations, participation and context sharing | Implemented prototype with server permission/disclosure checks. |
| Room purpose, defined outcome and criteria | Implemented optional metadata; purpose does not grant execution authority. |
| Shared memory and revision/source scope | Basic implementation; enhanced commercial memory is separate. |
| First-connected governing agent, human replacement | Appointment/authority foundations exist; full automatic initial selection/integration remains incomplete. |
| Debate, objections and explicit human resolution | Intended core behavior; disagreement/dependency orchestration remains partial. Free agents do not automatically vote or approve by timeout. |
| Seven-part task definitions | Implemented versioned task/input/process/output/criteria/constraints/failure contracts. |
| Actual inputs and immutable execution history | Implemented value binding, provenance, dependency validation and recorded attempts. |
| Task creation from reviewed conversation proposals | Implemented bounded proposal/review foundations; general automatic extraction remains separate future work. |
| Bounded task assignment and delegation | Implemented controls; some production execution integration remains incomplete. |
| Sequential workflows and templates | Implemented supported runner/editor and human gates; not a general parallel engine. |
| Isolated coding and bounded commands | Substantial internal implementation; configured development currently blocks at production admission. Command profiles currently cover Go build/test. |
| File, bundle and patch deliverables | Contracts, capture/read/custody/manifest internals exist; required files block dispatch until the production storage path is connected. |
| Human review and acceptance | Implemented value/criterion contracts; actual file/patch review and delivery remain incomplete. |
| Basic repository commit/push | Confirmed core requirement; no completed production publisher. Current agent instructions prohibit commit/push. |
| Repository references and branch identity | Reference metadata/internal bootstrap exist; links are not proof of publishing or checks. |
| Storage accounting and finite limits | Implemented reservation/retention foundations; complete physical enforcement/inventory incomplete. Current documented free ceiling is5GB; tier capacity changes need explicit decisions. |
| Mandatory container/credential protection | Core infrastructure; never a paid bypass or optional off switch. |
| Authentication and protected audit history | Implemented owner/actor foundations, append-only transactions and integrity stops. Advanced commercial authority is separate. |
| Stop/cancel, diagnostics and explicit recovery | Implemented substantial paths; unknown external outcomes and live acceptance remain incomplete. |
| Idle shutdown and fresh containers | Internal policy/lifecycle foundations; production/provider acceptance still needed. |
| Reviewed harness/version upgrades | Signed review/admission foundations; real upgrade publication/compatibility rollout remains incomplete. |
| History, status, attention and interface preferences | Implemented prototype controls; complete server-wide search/export and notification escalation remain future work. |
| User/technical Help and setup/preflight diagnostics | Implemented, currently being reorganized by audience. Offline success is not live acceptance. |
| Managed secrets editor and explicit export | Confirmed free core; requested protected key/value storage, permission-checked reveal/edit/export, redacted audit and safe file publication are not complete. |
| Secret-exposure protection | Required for every adapter regardless of payment; comprehensive detection/interception remains requested work. |

### Paid add-ons — confirmed scope or explicitly recorded paid requirements

| Functionality | Current implementation / integration state |
| --- | --- |
| Multiple human users/team collaboration | Confirmed paid; existing historical human records also occur in the OSS prototype and must be preserved during extraction. |
| Companies/organizations and reporting hierarchy UI | Confirmed paid; existing core chart/model/routes require deliberate packaging migration. Security identity/scope is still core. |
| Multiple projects/project portfolio | Free is one project; paid multi-project layout requested, not a complete implemented project model. |
| Scoped company approvals, human authority and escalation | Paid requirements; explicit permissions rather than titles grant authority. Partial underlying data exists. |
| Consensus/voting and governed dispute resolution | Paid requirement; complete implementation absent. |
| Delegated timeout resolution | Paid requirement only within explicit granted limits; timeout/payment never authorizes an irreversible action by itself. |
| Vendor-supplied governing agent, replaceable by customer | Recorded paid plan; complete implementation absent. |
| Fulginn enhanced memory | Recorded paid plan; integration incomplete. |
| Veriscribe attributed organizational knowledge/wiki | Recorded paid plan; integration incomplete. |
| Multiple/distributed workers and fleet management | Confirmed paid; distinct from core access to one approved Linux worker. |
| External task-system integrations (GitHub Issues/Jira/etc.) | Confirmed paid; internal task creation/execution remains free. Repository publishing is a different core operation. |
| Monitoring clients' applications / remote exception intake | Paid monitoring requested; implementation/integration not complete. Basic CommonRoom reliability remains distinct. |
| Billing controls and Stripe commercial payments | Confirmed paid; checkout/webhooks/pricing enforcement are not implemented in inspected commercial main. |
| Startup-discount application/questionnaire | Commercial pricing workflow requested; implementation not verified. |
| Personal/Startup/Enterprise licensing policies | Commercial scope; latest pricing/lifetime availability and resource charges must be sourced from owner decisions, not stale repository proposals. No unlimited plans. |
| Enterprise compliance capabilities | Recorded future paid direction; no certification or implemented enterprise control set is claimed. |

### Existing commercial scaffolding — not approved product requirements

| Functionality | Verified state |
| --- | --- |
| Signed installation-bound entitlements | Private implementation validates Ed25519 signatures, validity/expiry and feature scope. |
| Private license files and current-access checks | Private implementation; not a checkout or license-issuer service. |
| Owner-only commercial status and feature settings | Private API/UI gating exists. |
| protectedAuditArchive | Default-off paid scaffold; gated operation returns HTTP501. Unapproved entitlement demonstration placeholder; no user-defined requirement. Exclude from the approved feature/module plan. Basic audit/retention remains core. |
| managedOrchestration | Default-off paid scaffold; gated operation returns HTTP501. Unapproved entitlement demonstration placeholder; no user-defined requirement. Exclude from the approved feature/module plan. Ordinary task/workflow execution remains core. |

### Placement or detailed scope still undecided

| Functionality | Outstanding decision |
| --- | --- |
| GitHub/GitLab/third repository provider coverage | Basic publishing is core; initial provider set and any provider-specific add-ons need confirmation. |
| Mercurial/CVS support | Mentioned as possibilities, not an accepted implementation or tier commitment. |
| Own-application exception reporting to working agents | Requested; consent/data minimization, reporting destination and automatic task behavior remain undefined. No transmission is authorized by this inventory. |
| Parallel workflow joins, advanced automatic recovery and cross-room handoff | Backlog scope; exact behavior/tier remains undecided. |
| MCP, skills and public plugin execution | Future/coming-soon scope; no completed runtime or settled tier split. Commercial add-on composition is a distinct architectural need. |
| Hosted offerings and separately charged infrastructure | Free hosting is an intention in the older specification; actual availability, hosting cost and resource pricing remain separate decisions. |
| Anonymous adoption telemetry and optional external decision services | Backlog ideas only; no default data collection or runtime authority approved. |
| Full history search/export and advanced human-attention preferences | Backlog requirements; tier and final behavior need confirmation. |

## Functionality

The commercial product is an add-on to the core, not a separately evolving copy of
the whole application. Models represent domain data/state; controllers coordinate
permitted use cases; views present outcomes and collect input. Worker transport,
provider adapters, persistence and cryptography are supporting infrastructure,
not business controllers merely because a type is named controller.

The inspected private repository contains an older app copy plus entitlement
scaffolding. Recent OSS artifact/governance work must not be presumed present
there. Packaging must preserve one authoritative core implementation, with private
commercial controllers/models/views extending explicit core contracts. The two HTTP501 entitlement placeholders are implementation scaffolding, not
approved features; plan their removal/replacement rather than inventing product
requirements from their names. Concrete MVC module/directory moves follow this
functionality inventory and use one authoritative core.

Current core feature-flag inventory (configuration is not a commercial entitlement):

`taskRepositoryLinks`, `taskProgress`, `taskAcceptanceReview`, `taskValueDependencies`, `taskDefinitions`, `roomOutcomeReview`, `roomOutcomeContext`, `roomOutcomes`, `automaticAgentRooms`, `taskDelegation`, `taskProposals`, `agentCoordination`, `idleStopNotices`, `idleAgentShutdown`, `workspaceDisposal`, `taskDispatch`, `contextPolicyEditor`, `jobDiagnostics`, `agentRecovery`, `agentBackgroundWork`, `agentSafetyControls`, `humanProfileImages`, `connectionSignIn`, `connectionTesting`, `helpWindowChoice`, `squadProfileActions`, `agentInvitationLayout`, `standardHelpIcons`, `featureFlagTabs`, `historyDateGroups`, `settingsTabs`, `roomSections`, `humanProfileDetails`, `topAlignedViews`, `compactProfileOverlays`, `companyOrgChart`, `newestMessagesFirst`, `roomViewportFit`, `featureFlagSwitches`, `workflowExecution`, `agentStatusDetails`, `squadMenuLayout`, `humanAttention`, `historyRefinement`.

## Properties and state changes

This inventory changes no runtime behavior, authority, feature defaults, pricing or
commercial access. Pending classifications remain explicit rather than silently
shipping paid functionality in core. Existing identities, messages, task records
and audit history must survive packaging/refactoring. Latest owner scope overrides
stale free/paid assumptions in archived documents.
