# Read-only physical project-quota observer

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


`internal/projectquota` observes an independently provisioned XFS project quota
through an already opened directory descriptor. It checks actual kernel state,
not Docker labels, directory-size samples or account reservation counters. It
changes no quota, creates no volume and grants no read/write/export authority.
It is not yet wired into production storage admission or reader witnesses.

## Supported boundary

The native adapter is restricted to Linux x86-64 with `quotactl_fd` support
(Linux5.14+). Other platforms refuse with `ErrUnavailable`; there is no device-path,
command-line, privilege-escalation or host-execution fallback. XFS is the only
supported filesystem. No syscall runs at package initialization.

Trusted provisioning supplies a directory opened under actual creator/custody
ownership and an independent `Expected` binding: device, inode, nonzero project
ID, reserved byte bound and finite inode bound. Never derive that expected binding
from model output or simply adopt the first directory observed. Byte/inode bounds
are not substitutes for account admission. Current account limits, complete
retained/staged/uncertain inventory and durable reservations remain mandatory.

`Inspect` duplicates and pins the descriptor with close-on-exec before querying.
It never closes the caller's original directory. It rejects non-directory roots,
wrong filesystem/identity/project, missing project inheritance or realtime flags.
It reads project status and exact project limits twice, rechecks root identity,
and refuses changed observations or lost enforcement. Both project accounting
and enforcement must be active. Missing quota identity, unlimited or excessive
hard byte/inode limits, overflow, over-limit usage and realtime allocation refuse.
XFS counters are converted from512-byte units with checked arithmetic. A full
quota is still enforced, but this observation does not admit another allocation.

All queries are fixed read-only operations. No setter, enable/disable, discovery,
volume deletion, storage refund, file export or command runner exists. Results
are a point-in-time observation, not a long-lived capability or a promise that
an administrator cannot change quotas afterward. Revalidate under the actual
writer/creator fences immediately around every operation needing this evidence.

## Provisioning and containment requirements

An administrator must separately provision project IDs, finite hard limits and
inheritance and bind every relevant file/volume to the approved project. This
observer neither provisions nor proves complete descendants/account inventory.
Root-directory inheritance alone is not proof that older files have the right
project. Writers and lifecycle operations must remain fenced during observation.

All managed model/tool processes must remain in the reviewed rootless container
boundary, outside the host's initial user namespace. The kernel restricts changes
to project identity/inheritance in noninitial user namespaces, but this module
does not attest container namespaces or grant that property from metadata. Host
administrator changes and source forks replacing these observations invalidate
the wider guarantee. Actual daemon/volume identity, complete inventory, creator
ownership, current permissions and retention/export proofs remain independent.

Kernel permission refusal, unavailable enforcement or unsupported ABI returns a
sanitized `ErrUnavailable`. Do not grant agents extra host capabilities or launch
a privileged container to hide a refusal. Narrowly reviewed observation access
and deployment integration remain required. Nothing here automatically installs
a privileged helper. Missing evidence must continue blocking new allocations.

## Cancellation and ownership

Queries run synchronously. Context cancellation is checked before and after the
fixed calls and before returning evidence; canceled or uncertain results are
withheld. A kernel call may block in the kernel and cannot be detached by this
context API. Owners must join it while retaining their actual creator lease;
cancellation is never proof that a query or writer has finished. The pinned
inspection descriptor closes on success and refusal; a close error withholds
observations. No timeout may turn a still-active inspection into released custody.

This is mandatory security infrastructure, with no optional feature flag or paid/
owner exception. It cannot weaken authentication, isolation, accounting or audit.
No new application settings, HTTP route or user download flow is introduced.

## Verification and remaining integration

Tests inject synthetic kernel readers and verify actual decoded observations,
query order, repeated root/limit/enforcement checks, cancellation, overflow,
unlimited/over-limit byte/inode refusal, realtime refusal and joined descriptor
ownership, including close failures. Unsupported-platform tests refuse actual
host substitution. Every automated kernel/DB/Docker/provider boundary is mocked;
no test enables quotas, mounts a filesystem or requires elevated privileges.

Linux x86-64 native source is cross-compiled separately from macOS validation:

```sh
GOOS=linux GOARCH=amd64 go test -c -o /tmp/commonroom-projectquota-linux.test ./internal/projectquota
```

Compilation is not kernel execution or live enforcement evidence. Still required:
live XFS/native ABI and privilege validation, reviewed namespace/volume mapping,
pre-provisioning and complete inventory, durable account/storage integration,
physical witness assembly and end-to-end storage exhaustion/recovery acceptance.
No P0 storage or public MVP gate is certified by these synthetic tests.

## ABI references

The adapter follows Linux UAPI layouts and uses only query operations:

- [XFS quota structures and enforcement flags](https://github.com/torvalds/linux/blob/master/include/uapi/linux/dqblk_xfs.h)
- [Filesystem project attributes](https://github.com/torvalds/linux/blob/master/include/uapi/linux/fs.h)
- [x86-64 syscall numbers](https://github.com/torvalds/linux/blob/master/arch/x86/entry/syscalls/syscall_64.tbl)
- [Descriptor-based quota queries](https://man7.org/linux/man-pages/man2/quotactl.2.html)
- [Project-attribute namespace restrictions](https://github.com/torvalds/linux/blob/master/fs/file_attr.c)

No third-party implementation or comments were copied into the Go adapter;
constants/layouts express the operating-system ABI. Live verification must pin
the actual deployment kernel and candidate binary, not these moving reference URLs.

Operators can run the [worker storage preflight](worker-storage-preflight.md) against private independently approved bindings. It performs read-only observations before database startup and grants no allocation or live acceptance.

## Bounded aggregate observation

`InspectSet(ctx, roots, budget)` observes1–32 independently approved roots under a
finite aggregate reserved-byte budget. Each `Root` supplies the already-open
`Directory` and independent `Expected` binding. All identities and bounds validate
before any kernel reader opens. Duplicate device/inode or device/project pairs
refuse, preventing a shared quota from being counted twice. The full expected
reservation amounts, including unused capacity, must fit the aggregate budget
with checked subtraction; low observed usage never erases reserved capacity.

The set owns copies of its expected bindings and pins every descriptor before
inspecting any root. It observes each project, then repeats in reverse order
while all descriptors remain pinned. Changed identity, hard limits, usage,
enforcement, cancellation, partial opening or any close error withholds the
entire result. Every acquired reader closes even if a later open/inspection fails.
The original caller descriptors remain caller-owned. Native operations retain
the same Linux x86-64 restrictions and synchronous cancellation contract.

This is a reusable aggregate observation primitive, not a complete inventory
source or a new admission route. The caller must independently establish account
ownership, the COMPLETE root/project list and writer/creator exclusion across all
workers, retained artifacts and staged bundles. Equal repeated observations do
not prove an atomic snapshot or exclude intervening changes. The supplied budget
is not trusted merely because it is finite: production admission must derive it
from actual durable account/reservation state. The operator CLI uses this aggregate observer for each configured company,
while retaining all configured directory descriptors through the complete check.
It remains a diagnostic; it does not prove the inventory is complete.
No optional flag, quota mutation, source disposal, refund or execution grant exists.

Synthetic tests cover aggregate boundary/overflow, duplicate quota identities,
all-roots pinning, repeated observations, partial opening (including an acquired
reader returned with an error), changed usage/limits/enforcement, cancellation
and joined cleanup. No real quotas or provider/database/Docker interactions run.
