# Agent activity, stops and recovery

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


[Documentation home](documentation-index.md) · [Product guide](concepts.md)

Understand what an agent is doing, request attention and recover without repeating uncertain work.

## Emergency agent cutoff

The owner can choose **Stop all agents** in the header or **Stop room agents** in a room. Confirm the scope and optionally record a reason. New agent dispatch and session creation/invitation in the stopped scope are blocked; managed turns receive cancellation requests. Human access, saved messages, memory, evidence and workflows are preserved. Per-agent **Remove** remains available and also requests cancellation.

Use **Resume all agents** or **Resume room agents** explicitly to allow new requests. Resuming application dispatch does not clear separately stopped rooms, and resuming a room cannot bypass an application stop or integrity incident. Cancelled work is not restarted. A cancelled turn’s result is discarded even if dispatch has since resumed. Completed effects and already recorded messages cannot be undone.

A recorded stop survives restart. Failed stop/resume persistence also attempts
 to latch the private integrity marker and cancels managed contexts; unsafe or
 unwritable marker storage needs owner investigation. Do not restart to clear a
 failed stop. Application stop can act before a successful state read; room stop
 needs room state. Use the [incident recovery procedure](operations.md#integrity-and-history).

Cancellation is a request until the worker proves the matching container absent.
Managed jobs use durable journals and explicit reconciliation. A killed SSH
process or provider EOF alone is not shutdown proof. Earlier effects are not
undone and provider tokens are not automatically revoked. Inspect uncertain jobs
before replacement work; never remove the isolation or audit guard to continue.

`COMMONROOM_AGENT_SAFETY_UI` defaults to `true` and sets the initial visibility. The owner can override it in **Settings → Features → Security → Emergency agent controls**. The saved setting survives restarts; changing it reloads the page. **Backend stop enforcement and the protected API cannot be disabled by this UI flag**; an active stop stays active, and the emergency banner remains visible. Owner permissions and CSRF/Host/Origin protection remain mandatory. See the operations/API guides for access when controls are hidden.

## Safe session recovery

The owner opens an agent's **three-dot actions menu** and chooses **Recover session**, then uses:

- **Check worker**: read-only verification of the approved worker. It must report
  no managed containers and no starting/running/unknown job records. A busy shared
  worker does not prove this particular agent is running; wait or inspect Background
  work. An unavailable worker never counts as stopped.
- **Resume session**: after another check at action time, return the agent to Idle
  while preserving its saved isolated provider session. A missing, legacy or
  workspace-mismatched session must use Start fresh instead.
- **Start fresh session**: return to Idle and clear the provider session binding.
  Room history, profile, credentials and private workspace files remain. This does
  not undo partial changes or erase conversation records.

No option automatically invokes the model or replays an old task. After recovery,
the composer selects the agent and preserves your unsent draft; inspect partial
work and explicitly send the next request. Workflow runs remain separate and
require their own explicit recovery. Integrity incidents, emergency stops and any
controller-active model turn block session recovery.

The owner-controlled Security flag **Agent session recovery** (`agentRecovery`)
defaults **on**. Off rejects recovery server-side and hides controls/gates the
script. Stored sessions and history remain. Reload when enabling optional scripts;
no controller restart is needed for flag changes. Recovery is single-controller,
remote-worker-only, and deliberately requires the entire shared worker to be idle.
Unknown/stale nonterminal records require administrator reconciliation; there is no
force override. Other controllers must not share this worker concurrently.

Install the updated `commonroom-worker-bin` before using recovery. Its new
`idle-v1` protocol is a bounded read-only probe; older workers reject it, and the
app reports that verification is unavailable. Follow the existing worker binary
installation in [connection sign-in](connection-sign-in.md). No image rebuild or
credential migration is required for this protocol change.

## Background work

Open an agent's **three-dot actions menu** and choose **Background work** for its expandable
operational timeline. Remote workers are polled through the approved SSH transport
every five seconds; unchanged states are recorded at most every fifteen seconds.
The view shows worker startup/running/terminal observations, controller completion
and timeout, and Codex-reported command/tool start and finish. Claude currently
provides worker lifecycle only, because its adapter returns a final JSON result.
This is not a complete OS process trace or the model's private reasoning. Command
contents, outputs, credentials and arbitrary provider text are intentionally omitted.

The timeline updates while it is open; older reading positions are preserved.
Its independently scrollable feed shows the latest 64 observations; earlier audit
records remain in History. Each turn records at most 64 observations, reserving
space for shutdown and completion. A heartbeat or running job does not prove useful
progress and never extends the existing three-minute hard turn maximum.

The owner can select **Stop all agents** inside the view. This immediately requests
the existing application-wide emergency stop without relying on the cosmetic safety
button flag. “Cancellation requested” does not mean “shutdown confirmed”: confirmed
shutdown appears only after the remote worker acknowledges the matching job's
container is absent and its record terminal. Failed confirmation is displayed.
Agents cannot stop/resume the application. Resuming permits new requests and does
not replay cancelled work. No monitor automatically changes permissions or retries.

Settings → Features → Cosmetic includes **Agent background work** (`agentBackgroundWork`),
default **on**, owner-controlled. Off blocks new telemetry collection and omits
timelines server-side, gates the script, and preserves existing audit history.
Existing monitoring loops may finish their bounded polling until their current turn
ends. No restart is required for the flag; reload to fetch optional scripts when
enabling it. Monitoring is single-controller and does not reconcile jobs after restart.

## Agent status explanations

**Click a squad status** (or press Enter/Space when focused) to open an agent
diagnosis panel. It shows the last activity, a safe explanation and the next step.
The owner can check whether credentials are saved, explicitly test a fixed provider
reply, or open History. Saved credentials do not prove an OAuth session is valid.
Connection tests may consume provider quota and send no room history. If sign-in
is required, use Connections to complete official provider authentication.

**Compose smaller request** selects that agent and focuses the message box while
preserving an existing unsent draft. It sends nothing. Inspect partial file changes
before sending; retries can repeat side effects. Emergency stops and integrity
incidents disable this action until owner recovery. There is no one-click reset
that erases the failure or automatically replays a failed workflow. Workflow recovery
uses its existing explicit Resume control. Close X and the help icon appear at the
top right. These actions reuse existing server-authorized APIs and their respective
flags; `agentStatusDetails` off removes this panel as well as the status explanations.

Hover over an agent's squad status, focus it with Tab, or tap it to see **Doing**,
**Why** and **Next**. Escape, pointer exit, loss of focus, scrolling or navigation
closes the tooltip. Room chat turns record “Replying to a room message”; workflow
turns record their workflow/step title. Historical turns without activity or error
records explicitly show that information as unavailable. Recognized failures have
safe explanations for time limits, sandbox namespaces, read-only workspaces,
worker/provider connectivity, sign-in, review format and workflow limits. Unknown
failures direct the owner to History rather than exposing raw stderr.

Chat turns currently have a three-minute hard limit, including worker startup;
workflow steps have an explicit limit of at most 180 seconds. A deadline is
reported as a time-limit failure even when stopping the process produces
`signal: killed`. Older killed-process records remain unchanged: they do not
prove a timeout, because resource exhaustion or another termination can also
kill a process. Inspect partial work before retrying, and split long tasks into
smaller steps. Expired provider sessions require connection sign-in and a
connection test. Background work provides bounded observations, not complete process monitoring.

New turns record a controller dispatch timestamp. With `agentStatusDetails` on,
Working turns at least 120 seconds old display **Working · taking longer** and
elapsed seconds in their status help. This is an age warning, not a heartbeat,
proof of model progress, or automatic timeout extension. Legacy events without a
timestamp do not invent elapsed time. Disabling the flag omits these fields and
warnings server-side; it never changes execution budgets or containment.

Managed chat and workflow turns share mandatory single-controller concurrency
limits: **eight active model turns overall, four per room**. An oversized chat
broadcast returns HTTP 429 before recording its message or launching any member
of that broadcast. Workflow capacity failures preserve the failed run for
explicit resume; there is no automatic queue or retry. Wait for current turns to
finish or request a smaller group. These resource safeguards have no off switch;
presentation flags do not bypass them. A separate four-container physical ceiling and durable journal checks also apply.
Spend/storage enforcement and multi-controller ownership remain unfinished.

The owner-controlled Cosmetic flag `agentStatusDetails` defaults **on** under
Settings → Features. It changes immediately after save/reload without restart.
Off removes details from server state and gates `/agent-status.js`; it never changes
execution, stop enforcement or recorded audit events. Existing room access controls
also apply to these explanations. The tooltip omits raw credentials, paths, prompts
and provider output. An output report of being blocked is not itself a recorded
execution failure; this feature reports application status and recorded errors.

## Messages needing human attention

Agent messages requiring approval, a decision, clarification or missing information
can start with `[HUMAN ACTION REQUIRED]`. New managed **chat** prompts request this
convention when **Human attention requests** is enabled. Explicit requests receive
a **Needs your attention** badge, a cobalt accent and pale background. The protocol marker is hidden from the displayed message body while its original text remains in history. A room banner
counts outstanding requests and **View request** moves to the oldest one. Sidebar
room counts expose outstanding requests only for rooms the signed-in human can
access. Counts persist across reloads because they derive from retained events.

A small set of direct request phrases in older agent messages produces **May need
your attention** instead. This is a suggestion, not proof: quoted phrases may be
false positives, and indirect requests may be missed. A question mark alone does
not trigger an alert. Agents should use the explicit first-line convention for
reliable detection. Human-authored messages cannot impersonate agent requests.

**Acknowledge** records who saw the request and clears its outstanding alert; it
**does not** accept work, approve a workflow, grant permissions or resolve the
underlying question. Any authenticated human with room access may acknowledge;
agent accounts may not. The original message remains immutable. An acknowledged
badge stays visible, and later messages can create new requests independently.

The owner-controlled Cosmetic flag `humanAttention` defaults **on**, changes with
save/reload and requires no restart to toggle. Off hides highlighting/counts, omits
attention metadata from API state, gates `/human-attention.js`, blocks acknowledgment
and stops adding the convention to new chat prompts. Prior message/acknowledgment
records remain; already-running prompts are not recalled. Re-enabling restores the
projection, including previous acknowledgments. Workflow approval nodes keep their
separate execution contract; this feature does not pause or drive workflow nodes.
No sounds, desktop notifications or external transmissions are added.

Manual attention UI verification uses the 60-second in-memory fixture:
`COMMONROOM_BROWSER_ATTENTION_TEST=1 go test ./cmd/server -run
'^TestWorkflowBrowserHarness$' -count=1 -timeout 2m -v`. It contains synthetic explicit,
suggested and routine agent messages. Acknowledgment affects only the mock store.

## Squad action placement

Background work and Recover session live in each agent's three-dot menu, before
the destructive Remove action. They do not add rows beneath the agent's status.
Opening an action closes the menu before showing its overlay. Background work
uses the existing `agentBackgroundWork` flag; recovery uses `agentRecovery`, is
owner-only and is omitted while the agent is Working. Server flags/permissions
and worker checks remain enforced. Turning off `squadMenuLayout` restores native
menu positioning but keeps these actions within the menu. Repeated rendering
does not duplicate items. Automated UI regressions cover placement, dismissal,
permissions, both feature states, active agents and the native-menu fallback.

## Free durable job safety

See [worker lifecycle, shutdown verification and diagnostics](worker-jobs.md) for mandatory protocol-v2 execution, durable intent, crash reconciliation, owner diagnostics and the `jobDiagnostics` flag (default on). Shutdown uncertainty never counts as success; no task is replayed automatically.
