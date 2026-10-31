# Tool-only live acceptance evidence

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


Recorded 18 October 2026 on a dedicated Debian 13 Linux/amd64 worker. These are
manual checks using disposable synthetic source and a credential-free tool image.
They are separate from automated tests and do not establish production dispatch,
provider compatibility, complete image security review or full P0 acceptance.

## Runtime preparation

The worker originally had only an inaccessible system Docker context. A separate
user-level Docker 29.9.1 daemon was installed and its local rootless context selected.
The system daemon was not changed. The user service is enabled for future logins;
the account already had persistent user services and subordinate UID/GID ranges.
No provider login, credential files or application database were used.

Installation followed [Docker's rootless instructions](https://docs.docker.com/engine/security/rootless/).
The downloaded installer was compared with the upstream source at commit
`ada7a567e59b3935bd601bb21e422da50e80e376`, allowing only the three documented
release-template substitutions. Installer SHA-256:
`dfa409e66addb813cfbfc10e9cfb866529477c0965c330b181411772365b9f32`.

Daemon observations include rootless mode, seccomp, cgroup v2, the systemd driver,
and reported CPU CFS quota, memory, swap and PID limit support. Actual container
kernel limit checks below supplement those reported capabilities.

## Immutable image inputs

The tool recipe and its minimal helper sources came from commit
`af171950bb37d647e07856117a63e7f9854327d4`. Only the nine files needed for the tool
build context were transferred; no checkout, Git directory, credentials, tests,
provider bootstrap or private application data were copied.

| Input | Immutable reference |
| --- | --- |
| Go Linux/amd64 base | `golang@sha256:bcef992b77b1e2031aaaa51da75cebc32da8e9c182e12e633ea79023c79d9eff` |
| Node Linux/amd64 base | `node@sha256:7f7f17f773027e7f0475fa829fc0d793d75918fda85b9e517c9cc5aa46200406` |
| Built local tool image | `sha256:98a2bba5bddfca7eefa228729e082f0c28a28bff719d644898d475185db2c739` |

BuildKit built `docker/tools/Dockerfile` with the explicit immutable base arguments.
Its helper compilation steps used network-none mode. The Go SDK was copied into
the resulting image. Required digest arguments intentionally have no defaults;
the builder's empty-default warnings do not authorize mutable image tags.

This image is a local acceptance artifact, not a published or approved release
image. Retain normal provenance, licensing and security-review requirements.

## Observed behavior

A fresh named volume and fresh stopped container were created for each check,
with unique synthetic identities. The container was inspected before start:
non-root UID/GID 1000, network-none, read-only root, no host bind mounts, 512 MiB
memory/swap, half a CPU and a 64-process ceiling. Files were written through the
fixed tool helper and their returned content digests checked.

A tiny dependency-free Go module tested addition, read-only-root write refusal
and actual cgroup values. `go test ./...` completed successfully with the fixed
offline command environment and workspace cache/temp paths:

| Observation | Value |
| --- | --- |
| Cold command duration on this worker | 107.48 seconds |
| Observed Docker client exit code | 0 |
| Combined output size | 22 bytes |
| Output SHA-256 | `b8eb1a0996c2911ae81adb8a05585079e612067211f15c894a551c99aec6` |
| Kernel `memory.max` | `536870912` |
| Kernel `pids.max` | `64` |
| Kernel `cpu.max` | `50000 100000` |

The cold check was close to the 120-second command bound. This single observation
does not establish timing for larger repositories. A fresh workspace has a cold
Go cache; dependency availability, compilation cost and resource bounds can
prevent a real repository's checks from completing.

A separate fresh container confirmed a detached background sleep process through
both a PID-inclusive Docker process listing and in-container `/proc` observation.
The exact container was then removed and its absence verified. Its exclusively
owned synthetic volume was removed after checking for references, and exact-name
absence was verified. The Go-check container and its volume were also removed.
No old agent container was restarted or adopted.

The Go diagnostic identity was `045980b699454be3a4541f9d9015737d`; the final
background-process diagnostic identity was `153c89e2ea414f1d965f0c58be52bf61`.
These identify disposable Docker diagnostics, not authenticated controller jobs.

## Limits of this evidence

These checks did not run a model or the controller's durable audit protocol.
They do not establish public admission, provider credential renewal, engine/tool
pair fencing, Git-bundle bootstrap, revision-bound artifact attestation or storage
capacity release. Separate automated tests exercise those implemented controller
and broker boundaries with synthetic dependencies.

Reported/kernel resource limits were checked, but memory/PID exhaustion, daemon
crash, transport interruption and recovery stress cases remain. The current plain
local volumes have no verified hard byte quota. The worker's ext4 filesystem has
no configured project quota, and provisioning requires administrator access.
Storage reservations alone cannot substitute for physical enforcement.

The full [live acceptance plan](p0-acceptance-tests.md#live-acceptance-is-separate-and-still-pending)
and all corresponding release gates remain open. See [isolated commands](isolated-commands.md),
[storage accounting](storage-accounting.md), and [release readiness](release-readiness.md).
