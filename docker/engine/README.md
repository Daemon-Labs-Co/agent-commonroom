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
disable every client customization or lifecycle hook. The eventual engine must
have a fresh controlled home/settings, no development repository mount, no
untrusted plugin/config discovery and only the supervisor-controlled broker link.
The broker must authorize and durably audit each action before a credential-free
tool container performs it. Pair cancellation, quotas, receipts and absence proof
remain mandatory.

Before enabling managed development, verify the actual pinned Linux binary's
callback behavior and native-route refusal, bind its artifact digest/provenance to
supervisor admission, and finish paired runtime/audit integration. Never accept a
stock binary merely because it reports the same version. OAuth compatibility and
end-to-end execution have not been verified. No runtime is deployed by preparation.

## Tests and flags

`make check` runs offline Python tests for source/policy/lockfile refusal, exact
local-only normalization and preservation of rejected input. The upstream Rust tests verify callback-only construction,
broader/native/deferred-list rejection and preservation of a stricter host ceiling.
They require the pinned upstream source and compiler, so are separate from the
fast CommonRoom suite. No DB or credential fixtures are used.

On 14 October 2026 the pinned Linux/amd64 factory compiled and all three selected
library tests passed. The successful retry took 3m 45s including container startup,
with Cargo reporting 3m 39s compilation and under 0.01s test execution. Dependencies
were cached from preceding failed/cancelled attempts; this is not a clean-build
benchmark. Removed native registrations leave unused-code warnings in upstream
code; no automatic `cargo fix` was applied. This result verifies router construction,
not OAuth or a production broker callback round trip.

The unoptimized verification CLI subsequently built with `cargo build -p codex-cli
--bin codex --locked` (7m 52s including startup; Cargo reported 7m 49s). Its
SHA-256 is `7f3faeeb218b04340d972ba3df3bad7f188e9203fa4e61f4710d961536076f88`.
Network-disabled, read-only containers with temporary empty homes successfully
reported `codex-cli 0.162.1` and generated the experimental app-server schemas.
These are metadata commands, not provider turns. The artifact is a development
verification build, not a release image; no production configuration was changed.
Release packaging and a real broker callback round trip still need verification.

There is no product toggle for credential isolation. Existing optional development
flags cannot enable the legacy credential-bearing native-tool path. A source fork
**CAN** restore native tools or bypass pins, **BUT** doing so lets tool execution
read engine-held provider credentials and invalidates the security contract.
