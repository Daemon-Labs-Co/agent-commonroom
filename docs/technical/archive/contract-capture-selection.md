# Capture selection from task contracts

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


`workspacecapture.ContractRequest` derives a finite helper request from the exact
versioned task definition. This internal metadata validation closes the gap between
the task's declared file outputs and the source members selected for capture. It
opens no root, starts no process and grants no execution or delivery authority.
It does not enable required file-task dispatch.

## Usage and output selection

Trusted worker/controller assembly loads the current task definition independently
and supplies its execution/source binding and fresh capture challenge:

```go
request, err := workspacecapture.ContractRequest(definition, binding, challenge, []string{"optional-report"})
```

Every required file or bundle output is included automatically. The optional list
may name only optional file/bundle outputs from that definition. An empty list means
required files only. Required values are handled by the existing value-output path;
they are never turned into file paths. Unknown/duplicate optional names, required
names repeated as optional, and selected value outputs are refused.

File names, output names and member byte maxima come from the definition. Bundle
members are expanded, and the complete request is sorted by output then member name.
The execution definition digest must match exactly. Existing snapshot constraints
also apply:32 members,16MiB aggregate declared member maxima, safe relative names,
no Git metadata, no shared source path across outputs, valid source identity and
fresh challenge. Unsupported selections fail before capture; no required member
is silently dropped to fit a limit. A task contract valid for metadata/storage may
still exceed this capture implementation's bounded selection limits.

## Current-authority validation

The trusted parent guard must call `ValidateContractRequest` against its own loaded
current definition. It requires the exact canonical member set, limits, identities,
definition digest and all required file outputs. It detects omissions, substitutions,
duplicate members, reordered/noncanonical requests and stale definition bindings.
A helper/model-supplied definition must never be treated as authoritative.

These functions prove only metadata correspondence. They do not prove the definition
is current, the user has permission, the workspace is frozen, bytes are stored, a
quota is physically enforced or a report is useful. Existing current human/task/
export guards, mandatory isolation and durable custody journals remain required.
There is no product toggle for bypassing this boundary; existing task/deliverable
flags and permissions remain server-enforced by their own execution paths.

## Actual byte validation and retained failures

After successful capture, sealed receipts may be passed to `artifactmanifest.Resolve`
for the corresponding definition/output and actual quota-backed custody store.
That operation rereads/hashes stored bytes under current read authorization and
durable acknowledgment, then applies registered format checks. A capture report
alone is not a validated manifest, public delivery or human acceptance.

Tests connect contract selection to actual synthetic source capture, private custody,
manifest resolution/revalidation and corrupt-object rejection. They also prove that
validly sealed non-JSON bytes cannot satisfy an application/json declaration and
that denied validation retains objects, receipts and charges. All kernel/authority/
quota/journal boundaries are explicitly mocked; no real provider, database or
container is used.

Production task dispatch, current-definition/authority composition, physical quotas,
reviewed image/live acceptance and authenticated file delivery remain incomplete.
