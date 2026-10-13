# Paid feature access

Commercial access is separate from an owner's feature preference. A supported
paid feature needs both a currently valid issuer-signed entitlement and an
owner-enabled flag. Client-supplied `paid` values, feature flags alone and an
owner login cannot grant an entitlement.

## Current scope

The paid-access foundation is implemented. Payment checkout, issuer service,
subscription webhooks, pricing, archive exports, usage metering/overage billing
and managed model orchestration are **not** implemented. Both integration
boundaries return HTTP 501 after valid entitlement and enabled flag checks; no
records leave the application and no planning model is invoked. Settings
explicitly shows integration availability rather than claiming protection.

Supported default-off flags:

| Flag | Settings category | Proposed capability |
| --- | --- | --- |
| `protectedAuditArchive` | Security | Independently protected immutable archive, checkpoints and rollback detection |
| `managedOrchestration` | Workflows | Internal planning, task organization and proposed assignments, independent of room membership |

Flags are saved through owner-authenticated, CSRF-protected
`feature.flag.updated`. Enabling without the feature's entitlement returns 402;
non-owners are denied. Disabling remains possible after expiry. Flag changes
need no restart. License file changes are verified on every access; expiry,
removal, wrong installation or invalid signature immediately removes effective
access. The saved preference/history remains; renewing a grant restores that
preference. These are preparation preferences until the adapters are available.

Basic append-only records, database privilege restrictions, authentication,
authorization, integrity shutdown, mandatory container isolation and retention
remain required and cannot be sold as bypassable controls or disabled by these
flags. Future archive disable/expiry must stop new exports without unlocking
retained objects, deleting history or erasing evidence. Billing never overrides
retention, including the 30-day minimum and human-controlled purge.

## Commercial model

Billing periods will be monthly, quarterly and yearly; base prices are not yet
selected. Protected Audit Archive includes **100,000 archived records per month**.
Additional blocks provide **100,000 records per month each**, at a separately
chosen price **Y per block**. Quarterly/yearly payment does not change the monthly
allowance. Discounts, billing-cycle boundaries, proration, rollover, overage
behavior and managed-vs-customer storage ownership remain decisions for checkout
and metering implementation. No automatic purchases or overage charges exist.

The proposed billable unit is one archived audit record, not an S3 storage object
or upload attempt. Batching and retries must not change billable volume. Future
metering must identify records uniquely and charge only verified durable exports,
with bounded record sizes. The present license can attest purchased extra blocks;
it does not measure usage or enforce consumption limits because exports do not
exist yet. Block counts outside 0–1,000 are rejected as unsupported; this is a
validation bound, not a published product plan.

Managed orchestration needs a separate model-call/token budget and pricing
policy. An archive record allowance must never authorize unrestricted model
spend. Bedrock is one possible model hosting adapter, not a particular model or
proof of quality. Evaluate a selected model with synthetic task-planning examples,
schema validation, failure cases, cost limits and explicit context-disclosure
rules before dispatch. An internal model may propose plans; the application must
retain permissions, workflow transitions, safety boundaries and human acceptance.
Any managed model integration must preserve the mandatory worker isolation
contract. No direct host execution fallback is introduced by commercial access.

## Issuer and deployment setup

A deployment operator configures the commercial issuer's **public** Ed25519 key:

```sh
export COMMONROOM_PREMIUM_ISSUER_KEY='BASE64_ISSUER_PUBLIC_KEY'
export COMMONROOM_PREMIUM_LICENSE_FILE='/private/config/premium-license.json'
```

These are placeholders, not usable credentials. With no issuer key configured,
all paid features remain locked. An invalid configured key stops startup with a
sanitized error. Never install the issuer's private signing key on CommonRoom.
It belongs to a separately administered licensing service that issues grants
only after verified payment or a deliberate authorized trial. No issuer, trial
shortcut or default/test license ships with the application.

On first configured startup, the app creates `premium-installation-id` beside
`OWNER_KEY_FILE` (private regular file, mode 0600). The operator supplies that ID
to the issuer; a grant for another installation is rejected. Preserve this file
with private installation backups. Loss/replacement needs a reissued grant, not
an owner-editable `paid` switch. Both identity/license files reject symlinks,
non-private permissions and oversized files (8 KiB maximum). Issuer key changes
require restart; replace license files atomically and retain mode 0600.

The license envelope is JSON with `payload` and `signature`, both standard base64.
The signature covers the exact decoded payload bytes. The payload contract is:

```json
{
  "version": 1,
  "product": "agent-commonroom",
  "installationId": "ISSUER_RECEIVED_INSTALLATION_ID",
  "features": ["protectedAuditArchive", "managedOrchestration"],
  "licenseId": "commercial-grant-id",
  "notBefore": "2026-10-13T00:00:00Z",
  "expiresAt": "2026-11-13T00:00:00Z",
  "archiveAdditionalBlocks": 2
}
```

This unsigned example does not unlock access. Issuer time bounds must be valid;
the app rejects unknown/duplicate features, wrong product/installation, bad
signature, expired/future grants and unsupported extra blocks. Archive blocks
require an archive grant. Enabling a flag records the authenticated owner through
the existing immutable audit path. The signed commercial grant is kept outside
room state; APIs return a safe status projection, never signatures or signing
material. File replacement/issuance auditing and online revocation belong to the
future licensing service; local file removal is the current revocation method.
Offline expiry relies on the deployment's clock: use protected time configuration.

## API and enforcement

- `GET /api/premium-features`: owner-only status for each supported feature,
  including entitlement, requested preference, effective flag, availability,
  expiry and archive allowance metadata. No checkout or arbitrary paid setter.
- `GET /api/protected-audit-archive` and `GET /api/managed-orchestration`:
  authentication and owner authorization, current entitlement (402), flag (403),
  then explicit unavailable-integration response (501). No network/model/storage
  side effects. Mutating methods are rejected.
- `GET /api/state`: owner receives safe `premiumFeatures`; other humans do not
  receive licensing details. Effective paid flags are false without entitlement.

All future action routes and adapters must call `requirePremiumFeature`; UI
visibility alone is insufficient. Commercial grants do not replace room access
or other authorization. Security and pricing enforcement must remain independent.

## Limits and tests

This gate protects against ordinary client requests; it is not tamper-proof DRM
against someone controlling the source, binary, environment and local filesystem.
Such an operator **CAN** replace the issuer key or fork verification. Managed
services and independently controlled integrations provide the commercial trust
boundary; an open-source local feature flag cannot guarantee payment. Protect
issuer administration and signing keys separately from customer installations.

Tests generate synthetic Ed25519 keys in memory, use temporary file fixtures and
mock every database interaction. They cover paid/unpaid and on/off access,
feature scope, installation binding, signature forgery, validity/expiry, unsafe
files, owner/member permissions, saved-preference removal after expiry, signed
allowances and truthful unavailable adapters. UI tests cover locked controls,
failed requests, duplicate clicks and non-owner visibility. No real payment,
archive, cloud model or database interaction occurs in automated tests.
