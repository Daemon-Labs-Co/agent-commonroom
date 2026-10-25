# Internal finite artifact manifests

`internal/artifactmanifest` resolves a file or bundle output contract against
actual sealed bytes in `internal/objectcustody`. It has no product API, UI,
managed dispatch hook or production adapter. File/bundle task dispatch still
refuses. No feature flag can bypass custody, audit, permissions or quota.
This internal foundation does not establish a production release gate.

## Finite contract

`Resolve` requires a valid complete task definition, its exact digest and the
company, room, task, execution, job, repository and revision identity. Each claim
must carry a sealed custody receipt for that exact execution, output name and
logical member name. A logical filename is metadata, never a host path. Custody
uses its independently generated object ID under its pinned private store.

A file has exactly its declared filename. A bundle has every declared member
exactly once; there are no optional or undeclared members, reused objects or
implicit directory/archive expansion. Preflight checks the entire claim set,
registered formats, exact reservation caps, individual and aggregate byte limits
before attempting any object read. Bundles support 1–32 members; task contracts
and custody impose a maximum of 16 MiB. The reservation is for the declared
maximum, not merely the actual smaller content.

Every read passes through the real custody store's sealed receipt, current
journal authorization, inode, hash, size and durable release acknowledgment
checks. A failed member returns no partial manifest and never deletes bytes or
refunds capacity. Earlier read acknowledgments record authorization to release
bytes within this process; they do not imply a complete bundle or remote receipt.

The result contains sorted members, actual sealed sizes and hashes, canonical
media types, structural-check versions, definition/execution identities and a
contract digest. Its own SHA-256 digest covers the complete manifest. `Revalidate`
requires exact canonical metadata and repeats current-authority and actual-byte
verification; recomputing a digest cannot authorize forged or stale metadata.
A trusted production journal must recheck actual source revision and authority;
matching caller-supplied strings alone is never proof of origin.

## Registered structural formats

| Media type | Structural check |
| --- | --- |
| `application/octet-stream` | Bounded opaque bytes, including empty content |
| `text/plain`, `text/markdown` | Valid UTF-8; control characters refused except tab, CR and LF |
| `application/json` | Strict bounded JSON, including duplicate-key, Unicode and nesting validation |

Text and JSON optionally accept only `charset=utf-8`, case-normalized. Opaque
bytes accept no parameters. Unsupported formats or parameters refuse before
reads. Empty text/opaque data is valid; empty JSON is invalid. There are no
arbitrary plugins, shell commands, archive extraction or image/document decoders.
Opaque acceptance is not a secret scan, malware scan or validation of semantics.
JSON structure is not proof that a human criterion or independent check passed.

## Required integration and verification

Production still requires a trusted isolated worker export transport, real
source/volume/revision identity, durable accounting and custody journal, complete
retained-data inventory, enforced physical quota, retention/disposal workflow,
and authenticated audited public delivery. No synthetic adapter may activate
managed output delivery. Export cancellation must be transport-aware. Independent
observed checks and owner acceptance remain separate from format validation.

Tests use private temporary files and a mocked in-memory journal, with the real
custody store. They cover canonical ordering, exact members and scopes, full-set
preflight refusal, strict actual content, tampering, revoked authority, missing
seals and release-audit failure. They do not exercise live workers, provider
accounts, a database, kernel quota, crash recovery or public download behavior.
See [custody](object-custody.md), [task definitions](task-definitions.md),
[architecture](architecture.md) and [release readiness](release-readiness.md).
