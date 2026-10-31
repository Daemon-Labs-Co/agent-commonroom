# Reviewed Codex engine source preparation

This directory prepares a dedicated **bridge-only** Codex client. It retains the
CLI account/subscription approach; it does not switch customers to API-key billing.
It is build infrastructure, not an available connection type or completed runtime.

## Pinned inputs

- Upstream: `https://github.com/openai/codex`, release `rust-v0.162.1`.
- Commit: `092d3acd6bec3e3a14bdc7e7a2810ab628ab759d`.
- Rust: `1.95.0`; Linux/amd64 verification builder:
  `rust@sha256:4c2fd73ef19c5ef9d54bee03b06b2839a392604fbfcd578ed948b71b37c1d7fb`.
- `lock-codex.py` pins both the original factory file and the replacement Rust
  fragment by SHA-256. Changes require source review and updated verification.
- The upstream release lockfile records 161 workspace packages as `0.0.0`, while
  release manifests use `0.162.1`. Preparation changes only those local package
  versions, checks the complete result hash and preserves external dependencies.
  Original lock SHA-256: `49fe417bcafe3f69f1f6636e23274f90d37e75e388127a76888c39001a4f8117`;
  normalized: `765110795bf3385d734bf1acfcae77660d29a6c678d5b77015d7ffe49f26e69f`.
  Never resolve fresh dependencies or remove `--locked` to bypass a mismatch.

From the CommonRoom repository, prepare a separate trusted upstream checkout:

```sh
python3 docker/engine/lock-codex.py /absolute/path/to/reviewed-codex-checkout
```

The tool refuses changed source, changed policy/lockfile, already-patched files and
symlink input files. It accepts only a trusted build checkout, not model-controlled paths.
The pins are review drift checks, not signatures or an independent supply-chain
attestation: editing this script can change the pins. Preserve upstream Apache-2.0
license and applicable notices when distributing a custom client.

Compile inside a dedicated rootless Linux build container with no provider
credentials, live repository, worker socket or Docker socket mounted. Use private
build storage, resource limits and the pinned compiler. Dependency downloads are
trusted build traffic; this does not authorize network access for development
tools.

When transferring source from macOS, omit AppleDouble `._*` metadata files and
extended attributes (for example, set `COPYFILE_DISABLE=1` when creating an archive).
These sidecars can be mistaken for SQL migration sources during compilation.
Only transfer the reviewed public source; exclude `.git` and private build caches.
Within the upstream `codex-rs` directory:

```sh
cargo test --lib -p codex-core commonroom_bridge_tests --locked
cargo build -p codex-cli --bin codex --release --locked
```

The filter selects synthetic router unit tests; it must not launch a real provider
or reuse personal sessions. Compilation alone is insufficient for deployment.

## Policy and remaining acceptance gates

The factory accepts exactly one non-deferred dynamic function named
`commonroom_tool`. It constructs a direct router containing that handler only;
native filesystem/shell/patch, MCP/app, hosted tools, collaboration tools and Code
Mode routes are not registered. An existing host tool ceiling can narrow access,
never broaden it. Missing or broader restored tool lists fail; there is no fallback.

This removes model tool routes in the reviewed factory, but does not independently
disable every client customization or lifecycle hook. The engine bootstrap enforces a fresh controlled home/settings, no development
repository mount, no untrusted plugin/config discovery and only the
supervisor-controlled broker link.
The broker must authorize and durably audit each action before a credential-free
tool container performs it. Pair cancellation, quotas, receipts and absence proof
remain mandatory.

Before enabling managed development, verify the actual pinned Linux binary's
callback behavior and native-route refusal, bind its artifact digest/provenance to
supervisor admission, and complete public dispatch and live runtime/audit
acceptance. Never accept a
stock binary merely because it reports the same version. OAuth compatibility and
end-to-end execution have not been verified. No runtime is deployed by preparation.

## Tests and flags

`make check` runs offline Python tests for source/policy/lockfile refusal, exact
local-only normalization and preservation of rejected input. The upstream Rust tests verify callback-only construction,
broader/native/deferred-list rejection and preservation of a stricter host ceiling.
They require the pinned upstream source and compiler, so are separate from the
fast CommonRoom suite. No DB or credential fixtures are used.

The pinned Linux/amd64 factory and three selected router tests were verified on
14 October 2026. An unoptimized development CLI also built and ran version/schema
metadata commands in a network-disabled, read-only container with an empty home.
Those checks establish source/router and metadata compatibility, not OAuth or a
production callback turn. They do not identify an approved release image.

There is no product toggle for credential isolation. Existing optional development
flags cannot enable the legacy credential-bearing native-tool path. A source fork
**CAN** restore native tools or bypass pins, **BUT** doing so lets tool execution
read engine-held provider credentials and invalidates the security contract.

The [engine bootstrap contract](../../docs/technical/archive/codex-engine-bootstrap.md) documents
the new runtime recipe and fixed entrypoint. No release image or live acceptance
is claimed. The trusted source-image exporter must include `/out/codex`,
`/out/LICENSE` and `/out/notices/` with all required notices.

## Reproducible confined build launcher

On the dedicated Linux rootless worker, export the exact public upstream commit
above into a fresh private directory (without `.git`, credentials or caches).
Transfer only `build-reviewed.py`, `lock-codex.py` and `codex-bridge-router.rs`
from this reviewed repository. Resolve the export's absolute path first:

```sh
python3 build-reviewed.py --source /var/tmp/private-build/source \
  --name commonroom-engine-build-unique-run
```

The launcher requires full Docker daemon evidence with rootless mode and explicit
boolean support for memory, swap, PID and CPU CFS quotas. Missing, false or
malformed capability fields refuse before source preparation or container start.
It also refuses noncanonical mount paths, preparation
parent symlinks, common private configuration roots and invalid container names.
It verifies source/policy/lockfile pins before creating the build container. The
pinned Rust image compiles the synthetic router tests and then the release CLI
with `--locked`, one compiler job/four CPUs, 8 GiB memory with no additional
swap allowance, 512 PIDs, all capabilities dropped and
no-new-privileges. Only the trusted public source export is bind-mounted; provider
environment variables and Docker endpoint overrides are not inherited. Network
access inside this trusted build is for locked build dependencies, not agent
workspace execution. The root filesystem/container build cache is writable for
the compiler; this is a build policy, not the stricter engine execution policy.
A different rootless deployment context must be configured as Docker's default
before running; this launcher does not accept a caller-supplied Docker endpoint.

The named container is deliberately retained for diagnostics and artifact
extraction. The binary is `/build-cache/target/release/codex` inside it. Record
its digest, compiler version, successful router tests and image provenance before
exporting it with upstream license/notices. On error retain the build log, repair
the build prerequisite and use a fresh source export/name: partial preparation is
not silently reused. Remove only the exact operator-owned build container after
saving required evidence/artifacts; no blanket prune or source disposal is done.
This tool neither publishes nor approves an engine image and never launches a
provider. It has no product feature flag because it is build infrastructure.
`make check` runs mocked launcher tests; they do not download images or run Docker.

An initial four-job release compilation exceeded the 8-GiB build bound after
all three synthetic bridge-router tests passed. The launcher now uses one
compiler job while retaining the same memory/isolation limits. A compile failure
is not an approved binary; record the actual final build result separately.
