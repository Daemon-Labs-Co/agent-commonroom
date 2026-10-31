# Worker physical storage preflight

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


Operators can check independently provisioned storage before attempting managed
work. Run the server binary on the approved Linux x86-64 worker:

```sh
./bin/commonroom --check-worker-storage /absolute/private/storage.json
```

This read-only command runs before controller/database startup. It opens only
configured directories and performs the fixed native queries described in the
[physical project-quota observer](physical-project-quota-observer.md). It does not
start Docker, providers, listeners or model turns, and never changes quotas or
creates directories. No database or network connection is made. Running on an
unsupported platform refuses; there is no alternate host/SSH execution path.

## Private provisioning policy

Use an owner-only regular configuration file, owned by the current operator,
with no symlink leaf or additional hard links, under operator-controlled parents.
The file must be an absolute clean path and at most64KiB. Version1 accepts1–32
roots. Unknown/duplicate properties (including case variants), invalid encoding,
trailing JSON, zero/missing bounds and duplicate roots refuse the whole policy.

The following is illustrative synthetic configuration; these are not discovered
or deployable identities:

```json
{
  "version": 1,
  "roots": [
    {
      "companyId": "example-company",
      "worker": "worker-a",
      "path": "/srv/commonroom/approved-project-a",
      "device": 17,
      "inode": 23,
      "project": 7,
      "maxBytes": 5000000000,
      "maxInodes": 100000
    }
  ]
}
```

Supply device, inode, nonzero project ID and finite byte/inode reservation bounds
from independently reviewed provisioning records. Never copy a first observed
identity into the policy to bypass a refusal. Company and worker are operator
labels, not inferred authorization. Byte amounts are exact bytes, not GB strings;
XFS hard limits must be expressible in512-byte blocks and no greater than these
bounds. Usage may equal the enforced limit; that does not authorize new storage.

Each root path must be absolute, clean and not the filesystem root. The command
opens a directory with no-follow on the leaf and pins its descriptor for the
inspection. Parent directories must remain operator-controlled; this is not a
sandbox for arbitrary caller paths. The actual observed device/inode/project
must match the independent binding. Duplicate paths, device/inode pairs or
project IDs on the same device refuse, even across different company labels:
a shared project quota cannot be counted as separate reservations.

## Aggregate budget and descriptor ownership

Before any root opens, the complete policy's reserved byte bounds are grouped by
company. Each company must fit the free edition's mandatory5,000,000,000-byte
ceiling, including reservations on different configured workers. Exact boundary
is allowed; one additional reserved byte refuses. Independent company groups are
checked independently. All paths are local to the host running this command;
worker labels do not contact remote workers or discover their storage. Separate
worker runs cannot be added together as a complete account inventory without
independent cross-worker fencing and durable admission. Operator labels alone
are not proof of actual account
ownership or complete inventory. There is no paid/owner bypass in this edition.

All configured directory descriptors remain pinned through the entire check.
For each company, the native `InspectSet` observer pins its own descriptors,
observes all projects and repeats in reverse order before releasing any reader.
Changed identity, limits, usage or enforcement, partial observations and any
cleanup failure refuse the whole run. Low actual usage never erases reserved
capacity. Repeated matching reads are not an atomic snapshot: actual operations
still need independent complete inventory and writer/creator fences.

## Results and diagnosis

Exit0 means all configured roots returned acceptable point-in-time observations.
Exit1 means refusal or an output failure. Exit2 means incorrect arguments. No
partial success is printed. Refusal messages never echo paths, identities or raw
kernel errors. Inspect private provisioning, filesystem/enforcement support,
finite hard limits, permissions and ownership through approved administration.
Do not grant agents elevated host capabilities to make this check pass.

Cancellation withholds success and joins synchronous kernel queries. Descriptors
close on success and refusal; cleanup uncertainty withholds evidence. An in-flight
kernel query cannot be forcibly interrupted by a context deadline. No timeout
should be treated as proof that creator ownership or writer activity has ended.

Success does not prove complete descendant inventory, quota provisioning,
rootless namespace/volume mapping, creator/custody fences, account capacity,
retention, export permission, durable admission or live provider acceptance.
Writers may change state afterward. The command grants no capability and is not
connected to production task dispatch. Production operations still need their
own independently fenced fresh observations and complete admission checks.

## Security, flags and verification

This is mandatory operator security infrastructure with no optional product flag,
paid exception or weaker fallback. It cannot bypass authentication, authorization,
isolation, audit integrity or retention. A source fork replacing the observer
invalidates evidence, even though the command additionally validates returned
identity and bounds. No HTTP setting, user-supplied path or model tool is exposed.

Tests mock all native quota queries and use synthetic temporary directory/policy
fixtures. They cover strict/private policies, scope duplication, symlink/file/
missing roots, descriptor ownership, changed or excessive evidence, cancellation,
all-or-nothing results and sanitized errors. Compilation and mock tests do not
certify actual XFS privileges, deployment mapping or any P0 live release gate.
