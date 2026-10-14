# Ready-to-work development agents

[Documentation index](README.md)

Development agents need a repository they can edit, Git metadata they can update, and commands that can execute **before their first development turn**. A reachable Docker daemon or a successful conversational reply does not prove those capabilities.

## What changes

This first development-workspace implementation requires an approved SSH connection to the Linux worker, including when the controller also runs on Linux. Direct local connection endpoints remain chat-only; mapping them to a development room fails closed.

Configured development rooms use a reviewed Git bundle as their repository snapshot. Each room participant/connection gets a separate writable clone at `/workspace/repository`, a private `feature/commonroom-session-…` branch, and persistent private work storage. Other rooms remain chat-only. Your host repository is never mounted writable. Agents must not commit or push; the owner reviews and signs changes manually.

Every development turn checks the snapshot, shell, Git/Go/Node/make tools, assigned feature branch, and write access to both ordinary files and `.git` metadata before the provider starts. Missing tools, changed snapshots, incomplete initialization or write failures stop the turn. There is no retry outside isolation. An interrupted initialization requires operator recovery of that specific private volume; it is never silently erased.

Codex development turns use `danger-full-access` **inside the mandatory rootless Docker container** so that its inner bubblewrap launcher does not require nested user namespaces. This is an explicit container-backed execution policy, not an automatic sandbox fallback. Claude development turns enable a fixed set of built-in code tools and noninteractive permissions inside the same boundary; customizations and MCP remain disabled. Chat and sign-in policies remain unchanged.

The container still has no general network, no host root filesystem, no Docker socket, no extra capabilities, a read-only operating-system filesystem, and CPU/memory/process limits. Provider HTTPS goes through the existing allowlisted gateway. Source-level modifications **CAN** remove the outer isolation, **BUT** using these provider policies on a host would give model-generated commands the host account's authority. Never introduce that launcher or use privileged containers to fix a namespace error.

## Feature flag and permission model

`COMMONROOM_DEVELOPMENT_WORKSPACES` defaults to **false** on the CommonRoom controller. It requires a deployment administrator to set it to `true` and configure `COMMONROOM_DEVELOPMENT_REPOSITORIES` as a JSON map of room ID to bundle SHA-256. Invalid enabled configuration prevents server startup. Alternatively persist these two values in the dedicated owner-only `.data/development-workspaces.json` file as `{"enabled":true,"repositories":{"engineering":"BUNDLE_SHA256"}}` (mode `0600`). Explicit environment values override the file. This configuration contains no provider or database credentials and is read during normal startup.

This is currently an **operator-configured flag**, not a Settings toggle. Ordinary users and agents cannot change it through room messages or event APIs.

The worker independently requires `developmentEnabled: true` in its owner-only configuration and an approved bundle for the requested room and digest. Both sides must authorize development. An authenticated controller cannot supply an arbitrary host path. Existing room membership, authentication, dispatch permissions, emergency stops and audit protections still apply.

Disable the controller flag and restart to return rooms to chat-only execution; the worker also rejects new development requests if its flag is false. Private edits and audit history are preserved. A mode or snapshot change starts a new provider session instead of resuming a session with different execution authority. Conversation history still appears in the room prompt. Stop active agents before changing configuration: flags are admission controls and do not cancel already-running containers.

## Installation

1. Prepare the normal rootless worker and sign in each connection using [connection sign-in](connection-sign-in.md).
2. Build `docker/agent/Dockerfile.development` with the repository root as its context, supplying reviewed registry digests for both the existing agent image and a compatible Go image:

   ```sh
   docker build -f docker/agent/Dockerfile.development \
     --build-arg AGENT_IMAGE='YOUR_REGISTRY/commonroom-agent@sha256:REVIEWED_DIGEST' \
     --build-arg GO_IMAGE='golang@sha256:REVIEWED_DIGEST' \
     -t YOUR_REGISTRY/commonroom-development:reviewed .
   docker image inspect YOUR_REGISTRY/commonroom-development:reviewed --format '{{.Id}}'
   ```

   The image includes Git, make, GCC, Go, the existing providers and Node. It preloads the dependencies in `go.mod`/`go.sum` during the trusted build. Development commands use an offline module cache, local Go toolchain and a writable work-volume temporary directory. Other languages/dependencies require a reviewed image extension. Do not grant arbitrary package-registry access to agents as a workaround.
3. Scan the repository and its history for secrets, review the intended committed revision, and create a bundle on the controller:

   ```sh
   git bundle create /tmp/commonroom-repository.bundle HEAD
   shasum -a 256 /tmp/commonroom-repository.bundle
   ssh commonroom-worker 'mkdir -p ~/.local/share/commonroom-worker/repositories; chmod 700 ~/.local/share/commonroom-worker/repositories'
   scp /tmp/commonroom-repository.bundle commonroom-worker:.local/share/commonroom-worker/repositories/project.bundle
   ssh commonroom-worker 'chmod 644 ~/.local/share/commonroom-worker/repositories/project.bundle'
   ```

   A bundle includes the selected commit and reachable history, not uncommitted edits, ignored `.data` files, working-tree credentials, Git configuration or hooks. **Tracked secrets, historical secrets, source instructions and committed malicious code still require human review.** Keep the containing directory private. The bundle must be a canonical absolute path without symlinks, a nonempty regular file no larger than 256 MiB, and not writable by group/others. Submodules and Git LFS payload downloads are not provisioned.
4. Install the updated worker binary **before** changing its configuration. Preserve `authDir`; do not copy or replace credentials. Extend `~/.config/commonroom-worker/config.json` (mode `0600`):

   ```json
   {
     "image": "sha256:YOUR_DEVELOPMENT_IMAGE_ID",
     "authDir": "/home/commonroom-worker/.local/share/commonroom-worker/auth",
     "developmentEnabled": true,
     "developmentRepositories": {
       "engineering": {
         "bundle": "/home/commonroom-worker/.local/share/commonroom-worker/repositories/project.bundle",
         "sha256": "YOUR_64_CHARACTER_BUNDLE_SHA256"
       }
     }
   }
   ```
5. Configure the controller, then restart it using the updated binary. Use the persistent file described above, or export:

   ```sh
   export COMMONROOM_DEVELOPMENT_WORKSPACES=true
   export COMMONROOM_DEVELOPMENT_REPOSITORIES='{"engineering":"YOUR_64_CHARACTER_BUNDLE_SHA256"}'
   ```

   Use the actual room IDs, not titles. Only explicitly mapped rooms receive development access. The worker mapping must match exactly. Switching to a new bundle checksum creates fresh private storage rather than overwriting work.
6. Run the existing connection test to verify provider authentication, then send a small development request in the configured room. A development turn must pass the workspace preflight before the model receives work. Connection Test verifies chat/provider readiness; it does **not** claim that every room has a provisioned development workspace.

## Technical contract and limitations

The versioned worker envelope has an optional `repository` SHA-256. Development arguments are admitted only when this field is present and are compared against the exact provider policy; arbitrary CLI options, environment values and mount paths remain forbidden. The worker binds this digest to the room component of the server-generated execution key, verifies the approved file checksum, and mounts only that file read-only. A private volume identity includes the snapshot digest.

`prepareWorkspace` in `docker/agent/workspace.cjs` initializes a clone without inherited Git configuration or active hooks, removes its remote, creates the assigned feature branch without a commit, and probes writes. It never resets a previously initialized workspace or switches its branch after initialization. Provider output retains the existing bounds and cancellation remains container-wide. Changing the mode replaces provider session identity; arbitrary client-supplied session provenance remains forbidden.

Snapshots and private workspaces are not synchronized between agents. Automatic patch export, merging, shared worktrees, task-specific branches remain separate work. Optional sequential orchestration is described in [workflow execution](workflow-execution.md). Operators can inspect the named volume with a separate hardened maintenance container and collect `git diff` plus untracked files for human review; do not commit or push on behalf of the owner. Existing transcripts are conversation context, not a filesystem checkout.

Automated tests use synthetic temporary files and mocked process launches, never a database, provider, or real repository fixture. They cover both flags, room authorization, checksum/permission failures, exact provider policies, preserved Docker restrictions, initialization/resume, failed commands, wrong branches and read-only metadata. Live image/provider verification is recorded separately from the automated suite.

## Deployment verification — 11 October 2026

The dedicated Debian rootless worker was updated and a secret-scanned committed repository bundle was configured for Engineering and Quality & Security. Both Codex and Claude performed real shell commands under the unchanged network/capability/root-filesystem limits, confirmed `/workspace/repository` and their separate feature branches, wrote/read/deleted a temporary probe, and reported Git 2.39.5, Go 1.26.9, Node 24.21.0 and GNU Make 4.3. Go reported `GOPROXY=off`, `GOMODCACHE=/opt/commonroom/go-modules` and `GOTOOLCHAIN=local`. Codex no longer encountered the nested namespace error in that development turn. A nonfatal provider-side MCP connection warning was observed separately; it did not prevent command execution.

These were manual deployment checks, not automated test fixtures; they used no database, live room messages or project edits. They do not prove every future provider version, language toolchain, task, or malicious workload is supported. A new image digest must be reviewed and reverified. The loaded snapshot is the selected committed revision; uncommitted fixes are not silently imported. Automatic export/merge remains unavailable. Sequential execution can be enabled separately; it does not transfer private workspace files between agents.

## Per-turn verified snapshot staging

Before a development container launches, the worker opens the approved bundle
with no-follow/nonblocking flags and validates that opened inode. Bundle reads
are capped at 256 MiB plus one detection byte, including if a file grows after
its initial size check. Special files, final-component symlinks, empty/oversized
files and group/world-writable sources are refused.

The launcher then copies and hashes one bounded source descriptor into the
private per-turn directory. It mounts that read-only checked copy rather than
reopening the original host path through Docker. A source replacement after
staging cannot change the approved bytes mounted by that turn; a replacement
before staging must still match the approved digest. Failed copy/hash/close
validation prevents launch and removes the partial copy. The per-turn directory
is removed by existing cleanup. Provision up to 256 MiB of additional temporary
space per active development turn; restart/crash orphan cleanup remains an
operator concern. This is mandatory snapshot enforcement, not an optional flag;
the development capability remains off by default and owner-configured as above.

Synthetic file fixtures verify substitution isolation, unsafe-type/permission/
size refusal and cleanup without a database, Docker or provider. Live bind-mount
behavior and disk-outage handling were not tested. Trusted worker-account/parent
directory control remains required. This does not separate credentials from the
provider's development tools; that finding remains open.
