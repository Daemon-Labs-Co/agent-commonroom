# Application, container runtime and agent separation

Status: proposed design. No code has moved and no behavior described here as
future exists yet. Current dispatch remains blocked by the credential boundary in
[execution](execution.md).

## Input

Three independent concepts, each owning one responsibility:

- Application (controller): rooms, tasks, permissions, audit and the browser.
  It runs on the host serving the web site, holds no provider credentials and
  never invokes a container engine.
- Container runtime: a separately maintained project in its own repository and
  directory. It owns the container engine, images, volumes, quotas, repository
  checkouts and cleanup proof. It exposes only the versioned agent-session protocol.
- Agent: a model instantiated inside runtime-created containers. It has no
  container-engine access and reaches the application only through scoped,
  audited callbacks.

An agent session request, sent by the application to the runtime, contains:

| Field | Meaning |
| --- | --- |
| Binding | Initiating human, room, participant, connection, job and deadline. |
| Repository | Provider URL and exact ref or revision to check out. |
| Model and adapter | Registered adapter identifier and model name. |
| Images | Workspace image and engine image, each by immutable digest. Defaults are the project's reviewed images. |
| Limits | CPU, memory, process, duration and storage reservation, within account quotas. |

Authority comes from the existing owner permissions, stop state and integrity
checks. Model output supplies no field of the request.

## Output

Protocol operations and their results:

- `start`: accepted session identity, or a refusal naming the failed admission check.
- `status`: bounded lifecycle state and evidence, never raw container output.
- `cancel`: acknowledged request. Cancellation does not certify termination.
- `read-artifact`: verified bytes of one declared member, or refusal.
- `prove-absence`: confirmation that the exact immutable container IDs no longer
  exist, or explicit uncertainty with storage and capacity retained.

Results leaving a session are a patch or branch bundle plus declared artifacts.
They are held for human review. They are not applied to a remote repository.

## Functionality

### Separation rules

The application imports a protocol package and nothing that touches a container
engine. The runtime accepts only the fixed protocol: no arbitrary executable,
shell, image reference by tag, or URL. The two projects share the protocol
definition and its version; each is built, tested and released independently. A
runtime may run on the application host or another approved host. Remote transport
alone never authorizes host execution.

### Container space

A container space is the durable place where agent sessions run: one runtime
process, one private state directory and one container-engine endpoint, under an
OS account that is not the application's. It is addressed only by its protocol
endpoint (a local unix socket or an approved SSH destination) and a pinned
identity, never by a filesystem path, so it can run on the application host or any
other approved host.

- Created and started by an operator, not by the application. The application
  registers an existing space and cannot provision, restart or reconfigure it.
- Long-lived. The space owns the lifecycle journals, quota accounting and exact-ID
  cleanup proof that must outlive any one session. Sessions inside it are
  short-lived and each ends with cleanup proof or retained uncertainty.
- Stopping a space cancels its sessions and keeps its state. A restarted space
  reconciles recorded sessions by exact ID before admitting new work.
- The first deployment is one space on the application host under a dedicated
  account with rootless Docker. Multiple spaces are a later extension of the same
  protocol.
- First transport: the existing [worker protocol](worker-protocol.md) over SSH to
  the space's account, running a separately built `commonroom-worker` executable
  with a forced command. This needs no daemon. A long-lived socket service is a
  possible later transport carrying the same actions, not a prerequisite.

### One isolated environment per agent

Each agent session receives a fresh pair of containers and its own checkout:

- Workspace container: the repository checkout and its toolchain. It holds no
  provider credentials.
- Engine container: instantiates the model through the selected adapter. Provider
  credentials reach only this container, over a private channel from the runtime.

Two agents on the same repository never share a working tree. Each produces its own
result. This preserves the separation that blocks development today until it is
verified; it is not relaxed to allow a single container to hold both roles.

### Container contract

A workspace image must provide an entrypoint that implements the tool protocol
(receive a bounded tool call, run it in the checkout, return a bounded result,
stop on signal) and must not require credentials or container-engine access.
An engine image must provide an entrypoint that instantiates a model from
credentials supplied once on the private channel and speaks the engine protocol.
The project's reviewed images are reference implementations of this contract.

Users may name their own images, so a repository can bring its own toolchain.
Any image, reviewed or not, is admitted only by digest and receives identical
containment, quotas, pair separation and audit. Admission is never based on tags
or on trust in the image author. An image that fails the contract is refused at
start.

### Repository access

The runtime performs the checkout using a read-only repository credential it holds.
That credential is never passed to either container. Publishing a result to a
remote repository is a separate operation using credentials the model cannot
reach, human authority and exact-revision approval, as in [publishing](publishing.md).
Neither is implemented in this proposal.

### Migration order

1. Extract the existing actions into a protocol package and complete the
   [worker protocol](worker-protocol.md) inventory.
2. Build a `commonroom-worker` executable in the runtime project that answers
   `ready-v1` and refuses every other action; prove a controller-to-worker round
   trip across the account boundary with a contract test and a live check.
3. Move the remaining actions one at a time without behavior changes, keeping
   `make check` green at each step.
4. Replace controller-side engine calls with protocol client calls.
5. Connect production dispatch through the verified pair, then a live provider
   acceptance on a real worker.

## Properties and state changes

This page changes no runtime behavior, default, flag or permission. Mandatory
container isolation is not optional: there is no off switch, no per-image
exemption and no uncontained fallback. Naming custom images is default
functionality of the runtime and has no feature flag, by owner decision; it is an
exception to the feature-flag requirement in CONTRIBUTING.md. Image admission is
still enforced by the runtime: digest-only references, the container contract,
quotas and pair separation apply to every image, and tests cover accepted and
refused images. Any other option introduced to implement this design requires a
documented default, server-side enforcement and tests for both states before it
ships. No option can bypass authentication, authorization, audit integrity or
retention safeguards.

A container space persists its journals, quota accounting and session records
across restarts; the application persists only the registered endpoint, pinned
identity and session audit events. Application, runtime and agent identities stay
distinct. Existing audit history,
job journals and storage reservations are preserved across any extraction. Until
the later steps are implemented and accepted live, passing synthetic tests are
not evidence that agents can develop this repository.
