# Task repository references

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


Repository references attach a small human-maintained list of GitHub issue, branch,
commit and pull-request URLs to an existing task in a shared room. They help people
find related work. A saved URL is not verified evidence, a deliverable, acceptance,
repository access, execution authority or permission to commit/push code.

This initial feature is **API-only**: there is no dedicated task dialog or Settings
switch yet. Administrators use the authenticated API described below. No GitHub API
call, clone, fetch, provider launch, webhook, credential exchange or automatic
synchronization occurs when saving or reading references.

## Configuration and permissions

`taskRepositoryLinks` defaults **off** (a missing flag is off). Only the company
owner can change it through the normal authenticated event endpoint. No restart,
migration or third-party setup is required. Send:

```http
POST /api/events
Content-Type: application/json
X-CSRF-Token: <current session CSRF token>
```

```json
{"kind":"feature.flag.updated","payload":{"name":"taskRepositoryLinks","enabled":true}}
```

Use `false` to disable. The server immediately refuses both reference reads and
writes with 403 while disabled. Existing references and append-only audit history
remain retained; enabling again restores reads. Flag changes do not alter tasks,
running work, repository permissions, retention or isolation safeguards. Non-owner
flag changes are denied by the server; hiding a UI is not the enforcement mechanism.

Enabled reads require a signed-in enabled **human** with room access. Only the owner
can replace or clear the list. Agents, disabled humans, anonymous requests and
humans without room access cannot use the endpoint. Personal rooms are excluded.
Writes also require valid CSRF, an operation key, current revisions and healthy
integrity state. No flag can override these protections.

## Read the current record

```http
GET /api/task-repository-links?room=engineering&taskId=task
```

Use an existing task ID and its actual shared room. A successful JSON response has:

- `record`: version 1, company/room/task identity, link-list `revision`, `links`, and
  last-change metadata when saved. A new list has revision 0 and empty links.
- `taskRevision`: the task's **current** revision to use for the next write.
- `title`: current task title.
- `canEdit`: true only for the owner. Other authorized humans can read the list.

The record's own `taskRevision` describes the last reference update; it may differ
from the current top-level `taskRevision`. Always refresh before editing. Reference
records are read through this endpoint, not the general state JSON.

## Replace or clear the list

POST replaces the complete list; it does not append one item. Obtain both revisions
from GET and send a fresh stable operation key for this logical edit:

```http
POST /api/task-repository-links?room=engineering
Content-Type: application/json
X-CSRF-Token: <current session CSRF token>
Idempotency-Key: <unique operation key>
```

```json
{
  "taskId":"task",
  "taskRevision":2,
  "revision":0,
  "links":[
    {"kind":"issue","url":"https://github.com/example/project/issues/12"},
    {"kind":"pull-request","url":"https://github.com/example/project/pull/13"},
    {"kind":"branch","url":"https://github.com/example/project/tree/feature/example"},
    {"kind":"commit","url":"https://github.com/example/project/commit/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
  ]
}
```

The example revisions are illustrative: use the observed task/list revisions, not
these constants. If supplied in the query, `taskId` must match the body. The server
chooses company, author, timestamp and next list revision; clients cannot supply
those fields. Success returns `committed:true` and the updated `record`. Task revision,
status, dispatch and acceptance remain unchanged.

To clear references, send `links:[]` with current revisions. `null`/missing links are
invalid. Retry an uncertain save with the **same operation key and identical body**;
the recorded response is returned without a second update. JSON whitespace/order
may differ on replay. Reusing a key for another payload or using stale revisions
returns 409. Reload and review the current list before preparing a new edit.

## URL and request limits

At most 16 distinct URLs are accepted. URLs must use the exact `https://github.com/`
host, with no credentials, query, fragment, percent escapes, backslashes or whitespace.
Only the four kinds shown above are accepted. Issue/PR numbers are positive decimal
integers of at most 10 digits; commit IDs are lowercase 40- or 64-hex. Branch paths use
a limited ASCII safe subset and at most 200 characters. Empty, dot and parent path
segments are rejected. URLs are at most 1024 bytes. Duplicate checking is conservatively
case-insensitive, including branch paths: case-distinct branches cannot both be saved.

The request body is limited to 24576 bytes and rejects unknown/duplicate JSON fields.
This is a bounded URL syntax check, not full Git reference validation, proof of
repository ownership, URL accessibility or existence. A safe saved URL may still
point to unavailable or untrusted content. Do not place credentials in links.

## Responses and audit

| Status | Meaning |
|---|---|
| 200 | Successful read, committed update or matching operation replay. |
| 400 | Invalid/oversized JSON, link syntax, duplicates, missing operation key or fields. |
| 403 | Feature disabled, non-owner write/flag change, or failed CSRF/authorization. |
| 404 | Task/shared room unavailable or room access denied. |
| 405 | Method other than GET/POST. |
| 409 | Stale task/list revision, operation conflict or unhealthy integrity. |
| 503 | State unavailable; storage failures may also prevent a successful save. |

Authentication middleware may return its own authentication error before this
handler. Do not infer resource existence from refusal responses.

Updates append `task.repository_links.updated` with versioned identity/revisions,
links and owner change metadata under the normal audited operation transaction.
Direct event injection is refused; use the dedicated endpoint. Replay validates
historical flag state, shared-room/task identity, owner authority and revision order.
Failed persistence trips the integrity stop; invalid/read-only requests do not
append updates. Clearing links does not erase older audit entries. No external
repository authority or task outcome is inferred from history.

## Verification and limitations

Tests use mocked database interactions and synthetic task/actor/URL fixtures only.
They cover default/off/on behavior, retention across disabling, owner/member/agent/
room permissions, CSRF, idempotency and conflicts, clear/replay, task revision binding,
URL/body limits, malformed JSON, integrity and persistence failures. No external
GitHub request, real database or provider is contacted. UI editing, GitLab/other
hosts and automatic repository operations are not implemented by this feature.
