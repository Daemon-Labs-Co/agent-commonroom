# Reviewed Codex engine bootstrap

The engine bootstrap and Docker recipe prepare a credential-bearing app-server
process for the reviewed callback driver. This is internal mandatory isolation
infrastructure, with no bypass flag or public connection option. It is not a
public production provider path or an approved release image. Internal creation,
attachment and callback execution exist; live acceptance remains outstanding.

## Image construction

Build with trusted immutable `ENGINE_SOURCE_IMAGE` and `NODE_IMAGE` references,
both containing `@sha256:` plus a lowercase 64-character digest, and the reviewed
`CODEX_BINARY_SHA256`. The source image must publish the reviewed bridge-only
binary as `/out/codex`, upstream `/out/LICENSE`, and `/out/notices/` containing all
required upstream/dependency notices. The image copies those artifacts and checks
the binary digest. There is no mutable-tag default, stock CLI installation,
package download or silently resolved newer version.

The Dockerfile-specific context is deny-all except the engine recipe, ignore file
and entrypoint source. Source and Node base image contents remain trusted build
inputs: digest spelling and copying a binary do not attest provenance, router
restriction, dependencies, absence of secrets or correctness of inherited image
configuration. Review/sign/scan the resulting image before admission. Pinned
source preparation remains documented in `docker/engine/README.md`.

Example shape, using operator-reviewed values rather than invented release pins:

```sh
docker build -f docker/engine/Dockerfile \
  --build-arg ENGINE_SOURCE_IMAGE="$REVIEWED_ENGINE_SOURCE_IMAGE" \
  --build-arg NODE_IMAGE="$REVIEWED_ENGINE_NODE_IMAGE" \
  --build-arg CODEX_BINARY_SHA256="$REVIEWED_CODEX_BINARY_SHA256" .
```

This command has not been run or verified here. Build in a dedicated rootless
builder without provider credentials or private repository mounts. A missing
artifact, notice directory, digest or compatible runtime must fail the build.
The compiled binary's shared-library/runtime compatibility still needs testing.

## Fixed process startup

The entrypoint refuses non-Linux execution, UID other than 1000, extra arguments,
pre-existing discovery roots, an existing engine home and unsafe credential input.
It exclusively creates `/tmp/commonroom-engine` and its `.codex` directory as
0700, writes authentication and a fixed minimal configuration with exclusive
creation and mode 0600, and never adopts, resets or deletes an old home.

Credentials arrive only on private stdin before any app-server request: the
literal `COMMONROOM-AUTH-V1\n`, an unsigned four-byte big-endian length, and
2–65536 JSON bytes. The bootstrap reads exactly each segment, handles short reads,
and refuses invalid magic, bounds or truncated input before opening the gateway
or launching the provider. It does not read ahead into the callback protocol.
The supervisor deadline and fenced cleanup bound a bootstrap waiting for input;
the blocking read itself supplies no independent timer or shutdown proof.

The worker holds the connection credential lease while provisioning this
one-shot frame. The engine creates its own UID-1000 mode-0600 auth file in its
private home, avoiding host credential bind mounts and mapped-host-UID file
ownership. JSON must use `auth_mode: chatgpt` with a nonempty access token and no
non-null `OPENAI_API_KEY`. Credentials stay in the engine and trusted worker,
never the controller, tool image, source volume, model callback, audit or logs.
No API-key billing fallback exists. A partial frame is never retried.

Deploy only a reviewed immutable image supporting this framing. There is no
credential-file mount fallback or permission-relaxation path.

The fixed executable is `/opt/commonroom/bin/codex`. Its arguments force file
credential storage and ChatGPT login, disable hooks/plugins/recommended plugins,
browser integration and system proxy discovery/fallback, enable skipping host
skill discovery, and select strict app-server configuration over stdio. These
options were checked against the local pinned rust-v0.162.1 source. A newer client
requires reviewed compatibility and a new immutable image; no in-place upgrade
occurs in this bootstrap.

The child receives an explicit minimal environment, fixed private HOME/CODEX_HOME,
XDG paths and loopback HTTP(S) proxy. No ambient API key, NODE_OPTIONS, custom CA,
plugin path or other provider setting is inherited. Its cwd is the fresh engine
home, with no development repository. The loopback proxy accepts only on
127.0.0.1:18080 and connects only to `/run/provider.sock`. The worker must still
provide and verify the network-disabled container and provider-allowlisted Unix
gateway; the JavaScript alone is not network isolation or endpoint authorization.

## Shutdown, custody and acceptance gaps

Startup failures print only a generic error on stderr; stdout belongs to the
app-server protocol. Provider stderr is discarded rather than copied to controller
logs or model output. Signals are forwarded, gateway connections close with the
child, and listeners are removed on exit. Neither signal forwarding nor stream
closure proves descendant or container absence. The paired supervisor must run
its fenced shutdown and retain unresolved evidence after any ambiguity.

The internal launcher, immutable-image inspection, private stdin provisioning,
owned gateway and combined callback operation have synthetic tests. Public
production dispatch and live image/transport acceptance remain incomplete.
Refreshed-token custody must also be integrated before claiming a reliable
subscription lifecycle: credentials modified in the temporary engine home
are not currently synchronized back to the worker's official credential store.
Never export them through model output or silently substitute an API key after
an authentication failure. Concurrent refresh handling needs separate binding. The worker now durably marks
credential custody uncertain before provisioning the frame. That marker survives
process exit and both-role fencing, blocks status/execution reuse and requires
explicit sign-in recovery. See [worker credential recovery](worker-credential-recovery.md).
It is a safety gate, not an implementation of refreshed-token export.

Tests inject synthetic filesystem, gateway and child-process adapters. They cover
private exclusive home creation, fixed arguments/environment, unsafe or API
credentials, old configuration/home refusal, process/gateway errors, signal
forwarding and connection cleanup. Static recipe tests check required pins and
the context allowlist. No actual image build, provider, credential, container or
network gateway is launched; these tests do not establish live compatibility.
Production development admission remains blocked pending this and all other P0
acceptance requirements.

See [reviewed engine-container runtime](engine-container-runtime.md) for
fresh creation, reported policy verification and remaining production integration.
