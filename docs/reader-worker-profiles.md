# Private reader worker profiles and service ownership

This internal deployment layer maps an approved company and worker to a fixed
reader control address and private certificate files. It connects private-file
identity loading, one-shot TCP setup, mutual TLS, exact retained-job binding and
controller authorization. It does not enable file-task dispatch, publish downloads
or supply physical storage evidence. No live release gate is certified here.

## Operator configuration

Keep the profile JSON and every referenced certificate/key/CA file outside source
repositories, model tool access and agent/container mounts. Parent directories
must remain operator-controlled. Use absolute clean paths, current-process file
ownership, owner-only permissions (0600 or0400), regular files and a single hard
link. Symlinks, oversized files and unsafe private files refuse through the
existing [private-file boundary](private-file-security.md).

The policy is version1, bounded to64KiB and1–32 entries. Ordinary JSON whitespace
and property ordering are accepted. Unknown fields and duplicate keys (including
case variants) refuse. Each `(companyId, worker, role)` must be unique. Roles are
`worker` for client-side identity or `controller` for server-side identity.
Each address must be a literal unicast IP with a nonzero port, never a URL or
hostname; unsupported zones and mapped aliases refuse. Certificate DNS identity
is checked separately from the literal address.

Example shape (documentation IP and placeholders; replace before validation):

```json
{
  "version": 1,
  "profiles": [
    {
      "companyId": "example-company",
      "worker": "worker-1",
      "role": "worker",
      "address": "192.0.2.10:443",
      "identity": {
        "certificate": "/srv/commonroom/reader/worker-cert.pem",
        "privateKey": "/srv/commonroom/reader/worker-key.pem",
        "authorities": "/srv/commonroom/reader/controller-ca.pem",
        "peerDNSName": "controller.example.test",
        "peerSHA256": "REPLACE_WITH_APPROVED_CONTROLLER_LEAF_SHA256"
      }
    },
    {
      "companyId": "example-company",
      "worker": "worker-1",
      "role": "controller",
      "address": "192.0.2.10:443",
      "identity": {
        "certificate": "/srv/commonroom/reader/controller-cert.pem",
        "privateKey": "/srv/commonroom/reader/controller-key.pem",
        "authorities": "/srv/commonroom/reader/worker-ca.pem",
        "peerDNSName": "worker-1.example.test",
        "peerSHA256": "REPLACE_WITH_APPROVED_WORKER_LEAF_SHA256"
      }
    }
  ]
}
```

For separated deployments, install only the appropriate endpoint roles on each
host; a client does not need the server's private key. A role missing from that
host's snapshot refuses the corresponding operation. Never derive a profile,
certificate path or address from a task, room message or model suggestion.
Fingerprints must be independently verified64-character lowercase hex, not
copied from an unauthenticated connection. Certificate/key/trust format, validity
and role rules are in the [reader authority guide](reader-authority-bridge.md).

Loading is all-or-nothing: every entry and referenced identity must validate
before any profile becomes usable. An invalid unrelated entry blocks the whole
snapshot. Selection never falls back to a different company, worker or role.
Configuration provides transport policy only, not room access or export authority.

## Offline preflight

After building the server binary, run:

```sh
./bin/commonroom --check-reader-profiles /absolute/private/profiles.json
```

This exits before controller lease acquisition, database startup or worker/provider
launch. It opens no listener and performs no network/Docker operation. Exit0 means
private policy and identity checks passed; exit1 means refusal; exit2 means wrong
arguments. Success reports the entry count and explicitly says network/storage/
live acceptance are unverified. Failure never prints supplied paths, certificate
contents or key bytes. Missing certificate PEM markers keep their actionable
BEGIN CERTIFICATE/END CERTIFICATE diagnostic. Other refusal messages remain
sanitized. Do not use this command's success as evidence of a live P0 gate.

## Service assembly and ownership

Trusted setup calls `newObjectReadWorkerService(ctx, privateProfilePath)` to load
an immutable snapshot under the original service lifetime. `Connect` selects the
worker role by the exact company from the bound receipt and the approved worker
alias, then supplies the original retained record and finite intent to the
one-shot connection setup. No network call occurs for an unknown company/worker.

On the controller, `ListenAndServe` selects the controller role from the original
human/worker-pinned session and passes exactly `tcp` and the fixed profile address
to an injected deployment-owned `net.ListenConfig`. Bind is bounded to five
seconds and canceled by service/session lifetime loss. A late/failed bind that
returns a listener closes it. `Serve` also accepts an independently provisioned
listener, but verifies its actual numeric address/port matches the profile before
Accept. The one-shot listener closes before TLS; uncertain creation/closure
refuses rather than retrying or replacing a session.

Use an approved `net.Dialer` on the worker. Network interfaces, firewall policy,
certificate issuance/renewal and peer mapping are deployment responsibilities.
No remote address is discovered and no shell, SSH or plaintext fallback exists.
The service does not create physical quota/custody/creator/export witnesses;
those remain mandatory independent dependencies of actual reader execution.

`Close` permanently refuses new work, cancels active setup and authority, and
joins serving calls, dial/listen setup and worker read pumps. Concurrent Close
callers also wait. Stop callbacks must remain nonblocking and must not reenter
service Close. Injected dial/listen implementations must honor cancellation;
listener Close must unblock Accept. Callers must join Close before releasing
actual creator ownership. Cancellation never proves cleanup or releases storage.

Rotation requires independently approved replacement profiles/certificates,
closing and joining the old service and loading a fresh snapshot. Editing files
does not change active trust. There is no automatic reload, room/account HTTP
configuration API, provider credential store or key-generation command here.
Missing configuration refuses reader setup, not authentication or isolation.
This is mandatory security infrastructure: no optional flag, owner exemption or
paid bypass can disable the checks. Source forks replacing loaders/factories or
physical evidence invalidate the supported security boundary.

## Verification and remaining work

Tests cover strict/private policy loading, duplicate/unknown/version/role/address
refusal, all-or-nothing identity validation, company/worker separation, listener
address binding, current controller permission and durable audit, shutdown while
idle/waiting/dialing/listening, late/failed listener cleanup and refusal after
shutdown. The full reader runtime matrix now uses profile/service setup before
mutual TLS, binding, journals and isolated read/recovery checks. The offline
command tests safe success/usage/error output. All DB, network, Docker and provider
boundaries are mocked; identities and files are generated temporary fixtures.

Still required: actual deployment configuration and firewall/PKI operation,
production task-dispatch assembly, real physical quota/inventory/custody/export
and creator evidence, reviewed images and live rootless acceptance, retained
object ownership after source disposal, incident reconciliation and public
file delivery/human review. This layer supplies usable internal service assembly,
not an automatically running production artifact service or public endpoint.
