# MVC and commercial add-on organization — proposed mapping

## Input

The [approved functionality inventory](product-functionality.md), one independently
usable free core, private commercial add-ons and current source responsibilities.
This mapping follows feature inventory; it is not a claim that source modules have
already been extracted. Current application behavior and security contracts must
remain stable during each extraction.

## Output

A standard MVC responsibility map, grouped by functionality, with infrastructure
kept separate and commercial code importing a narrow supported core interface.

| Layer | Free core | Private commercial add-ons |
| --- | --- | --- |
| Models | Owner identity, project/room, agent profile, task definition/execution, workflow, output/acceptance, repository revision, storage reservation, audit event, secret reference | Team/company membership and hierarchy, multi-project portfolio, worker fleet, external tracker mappings, billing/license/entitlement, advanced consensus/authority, enhanced memory/attribution |
| Controllers | Authenticate/authorize requests, validate typed commands, coordinate core application use cases, return results | Coordinate paid team/company/fleet/tracker/billing use cases through exported core contracts; validate entitlement as well as normal permissions |
| Views | Single-project workspace, task/workflow/results, agent configuration, storage, protected secrets controls, history/settings and Help | Team/company/project portfolio, fleet management, external task sync, subscription and approved advanced-feature panels |
| Supporting application logic | Execution orchestration, task validation, acceptance/recovery policies and publication coordination | Paid policies/use cases composed with the same core execution and security mechanisms |
| Infrastructure | Event/database adapters, isolated worker runtime, provider adapters, filesystem/quota/custody, private transports, authentication/cryptography and protected publishing | Stripe/tracker/vendor adapters and entitlement storage; never replacements that weaken core isolation/authorization |

A candidate layout is feature-organized MVC rather than another set of global
catch-all directories:

```text
core repository
  app/                       startup and dependency composition
  features/
    tasks/                   model + controller + view + application logic
    workflows/
    agents/
    projects/                one-project baseline
    rooms/
    artifacts/
    repositories/
    storage/
    secrets/
    history/
  infrastructure/            worker, provider, persistence and crypto adapters
  extensions/                narrow exported add-on contracts/registration
  cmd/                       executable entry points

private commercial repository
  addons/
    teams/                   paid model/controller/view
    organizations/
    projects/                multi-project extension
    worker_fleet/
    task_integrations/
    billing/
    consensus/
    enhanced_memory/
    attribution/
  app/                       compose a versioned core with approved add-ons
```

These are conceptual module responsibilities; actual Go package paths/exported
interfaces must be designed during extraction. An external private Go module
cannot simply import the core's internal packages or unexported cmd/server types.
A stable core import/module version and explicit exported contracts are required.

## Functionality

Models express invariants and state without depending on HTTP or DOM. Controllers
handle request/response and delegate workflow orchestration; views collect input
and display state without granting permissions. Long-running task/application
logic is not placed inside a new giant HTTP controller. Infrastructure adapters
satisfy explicit contracts and retain mocked test seams.

| Existing code/type | Responsibility and intended treatment |
| --- | --- |
| server.taskDispatch | MVC request controller currently mixed with use-case orchestration; extract TaskExecution application logic behind it. |
| room.Task / TaskExecution / acceptance records | Domain models; separate definitions, attempts and human decisions. |
| server.load / appendBatch / eventDatabase | Projection/application support and persistence adapter; not all domain logic belongs in the DB model. |
| controllerCaptureSession / controllerObjectReadSession | Authenticated infrastructure control sessions, not MVC controllers despite their names. Keep operation grants distinct. |
| captureWorkerService / objectReadWorkerService | Infrastructure ownership/cancellation/joining; share mechanical lifecycle code where behavior is equivalent. |
| workspace/custody/quota implementations | Mandatory execution/storage infrastructure, not a paid feature or an MVC view concern. |
| provider adapters / RunCodex | Infrastructure execution adapters called by core application logic. |
| global render wrappers | View composition debt; migrate functionality to explicit view ownership incrementally. |
| managedOrchestration / protectedAuditArchive HTTP501 placeholders | Unapproved entitlement demonstrations; remove/replace after reviewing their exact usage, not promote them to product domains. |

Commercial add-ons register approved routes/views and implement paid use cases;
they do not monkey-patch core globals or duplicate the base application. Add-on
code is trusted installed software, not a model-selectable runtime plugin. The core
must build/run/test without the private repository. Add-ons pin a compatible core
version and preserve normal authentication, audit and permission enforcement.

Start extraction around the complete isolated task-to-result-to-approved-push
workflow. Keep behavior-preserving moves separate from changes that open production
admission. A file move by itself is not architectural completion.

## Properties and state changes

This is a proposed responsibility map; no application modules have been moved.
Paid/free extraction must preserve existing prototype identities, scope identifiers,
audit history and task records. Historical company IDs must not be deleted merely
because company administration becomes commercial. Capabilities and entitlements
remain distinct from mandatory safety controls and ordinary feature flags.
