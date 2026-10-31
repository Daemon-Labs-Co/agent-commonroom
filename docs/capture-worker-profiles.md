# Private capture worker profiles and owned service

This internal deployment layer selects a fixed capture-control endpoint and
mutually pinned certificate identity for an approved company, worker and role.
It owns dialing, listener setup, serving calls and client shutdown. It does not
provision physical storage witnesses, start model workers or enable public
file/patch delivery. Authentication and mandatory isolation have no bypass flag.

## Operator configuration

Provision a separate capture policy outside source repositories, model tools and
container mounts. Use absolute clean paths and operator-controlled parent
directories. Policy and referenced certificate/key/CA files must satisfy the
[private-file boundary](private-file-security.md): current-process ownership,
owner-only permissions, regular single-link files and bounded sizes. Never derive
paths, endpoints, keys or peer pins from task/model metadata or a received frame.

The policy accepts ordinary JSON formatting and ordering, with version1 and the
explicit protocol `commonroom-capture-control-v1`. A reader policy, wrong protocol,
missing protocol, unknown fields, duplicate keys (including case variants), more
than32 entries, empty policies and duplicate company/worker/role entries refuse.
The whole policy is bounded to64KiB. Every identity in the file must validate before
ANY entry is usable, including unrelated entries. There is no partial loading.

Example shape using documentation addresses and placeholders:

```json
{
  "protocol": "commonroom-capture-control-v1",
  "version": 1,
  "profiles": [
    {
      "companyId": "example-company",
      "worker": "worker-1",
      "role": "worker",
      "address": "192.0.2.10:443",
      "identity": {
        "certificate": "/srv/commonroom/capture/worker-cert.pem",
        "privateKey": "/srv/commonroom/capture/worker-key.pem",
        "authorities": "/srv/commonroom/capture/controller-ca.pem",
        "peerDNSName": "controller.example.test",
        "peerSHA256": "REPLACE_WITH_APPROVED_CONTROLLER_LEAF_SHA256"
      }
    }
  ]
}
```

Install only the appropriate roles on each host. `worker` supplies client identity;
`controller` supplies server identity and requires the worker's CA, DNS name and
leaf pin, with its own controller certificate/key. Missing roles refuse rather
than borrowing another entry. Provision an independently verified64-character
lowercase hexadecimal peer leaf pin. Certificate files need explicit PEM markers.
The [identity provisioning contract](reader-authority-bridge.md) applies to the
shared cryptographic material loader. TLS1.2 and1.3 remain supported.

Addresses must be approved literal unicast IP:port values, with nonzero ports and
no hostnames, URLs, scopes or mapped aliases. Certificate DNS identity is checked
independently. The operator must enforce firewall and routing policy; local JSON
is transport configuration, not room/task authority or a physical quota proof.

## Offline preflight

Run the built application with an explicit absolute private policy path:

```sh
./bin/commonroom --check-capture-profiles /absolute/private/capture-profiles.json
```

This command runs before controller leases, configuration/database startup or
worker launch. It only reads the policy and referenced identity files: no database,
listener, dial, Docker, provider or model is used. Exit0 means every private profile
and identity passed local validation; it does not test endpoint reachability,
physical storage, authorization or live acceptance. Success reports the entry
count and explicitly leaves those observations unverified.

Exit1 means validation or success-output writing failed. A markerless certificate
gets a specific BEGIN CERTIFICATE/END CERTIFICATE correction; other refusals give a
sanitized configuration diagnostic without echoing paths, operands or credential
contents. Exit2 means missing/extra operands. Unknown/typo command names refuse
before startup instead of silently launching services. `--help` lists the command.
The reader preflight cannot validate capture policy, and vice versa.

## Trusted service use

Application/bootstrap code calls `newCaptureWorkerService(ctx, absolutePolicyPath)`.
No service-activation environment variable, startup flag, public route or automatic
listener is installed by this layer. The offline preflight below only validates
configuration. The loaded policy is one immutable service snapshot;
rotation requires closing it and constructing a fresh service under approved
configuration. Changing a file does not mutate an active session's trust.

For a worker, call `Connect(ctx, worker, retainedRecord, finiteIntent, localStop,
netDialer)`. Selection uses the finite intent's company and exact worker/client
role. The endpoint pins the original retained record and copied capture intent.
It dials only that endpoint once. Pass the returned authenticated client to the
[owned capture runtime](capture-supervision.md) or independently authorized exact-ID
cleanup recovery. The client outlives setup but follows its original context and
the service lifetime. Call its Stop method when finished.

For a controller with the ORIGINAL authenticated human/worker session, call
`ListenAndServe(ctx, originalSession, netListenConfig)`. It creates only the approved
TCP endpoint with a five-second setup bound. Alternatively pass an already-owned
listener to `Serve`; its reported address must exactly match the selected policy
address before Accept. Both paths consume the listener and use the authenticated
one-shot protocol. A source-injected listener lying about its address invalidates
network confinement. No peer-selected operation or alternate endpoint is accepted.

Always call `Close` on the service. It prevents new admissions, cancels active
setup and retained clients, and joins dial/setup calls, serving calls and client
read pumps. Early policy/address refusals also finish listener closure and
controller integrity revocation before tracked serving work ends. Dependencies must honor context cancellation and listener Close must
unblock Accept. Stop callbacks must be nonblocking, local, and must not reenter
service Close or join their own pump. A failed/late listener or dial result is
closed and refused; no automatic retry occurs.

## Permissions, retention and limits

TLS/session binding cannot replace current original-principal/task permissions,
server feature flags, exclusive creator ownership, complete inventory, quota,
source fencing or custody evidence. Service shutdown leaves unfinished local and
controller history and uncertain capacity retained. It is not proof of helper
absence, artifact persistence or safe refunds. Renewed cleanup uses independent
current authority and an already journaled exact ID.

Tests use temporary generated identities, synthetic listeners/dialers, in-memory
TLS pipes and mocked database/evidence. They cover lifecycle/audit binding,
all-or-nothing policy validation, company/worker isolation, protocol separation,
wrong listener addresses, setup failure/cancellation and joined shutdown while
dialing, accepting and idle. Other capture tests retain enabled/disabled permission
coverage. No real Docker/provider or database is used.

Deployment bootstrap, actual PKI/firewall provisioning, independent complete
inventory and physical storage/custody/export enforcement remain incomplete.
Public task-output attachment/review and live Linux/provider acceptance remain
open. Synthetic service tests do not certify those release gates. Source forks
substituting policy loaders, transports or evidence adapters can bypass contracts.
