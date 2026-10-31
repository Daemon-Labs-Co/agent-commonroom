# Development workspaces: current boundary and preparation

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


[Documentation index](documentation-index.md) · [Provider/tool separation](provider-tool-separation.md)

**Public development dispatch is blocked.** The old single-container development
path exposed provider credentials to model-controlled tools and is no longer an
acceptable execution path. Its flags cannot re-enable it. Chat, sign-in and
connection tests are separate capabilities.

The replacement Codex engine/tool pipeline exists internally and has synthetic
coverage. Production admission and live provider/container acceptance remain open.
Do not treat a healthy worker or successful chat reply as permission to run
repository commands.

## What the replacement must provide

- A fresh provider engine and a separate tool container for every execution.
- No provider credentials in the tool container or writable host checkout mounts.
- A reviewed repository snapshot, private feature branch and exact job/repository scope.
- Controller-authorized, durably audited tool intent before action and outcome before
  output is returned to the model.
- Independent cleanup proof for both container roles; uncertainty retains ownership.
- Preserved source until an explicit, authorized disposal workflow permits removal.

See [paired containers](worker-pairs.md), [workspace bootstrap](workspace-bootstrap.md)
and [worker pair assembly](worker-pair-assembly.md) for the implemented contracts.
Bounded build/test commands, trusted artifacts and approved disposal remain P0 work.
Agents must not create commits or push; the owner reviews and signs changes.

## Configuration is preparation, not activation

The controller's `COMMONROOM_DEVELOPMENT_WORKSPACES` flag defaults to **false**.
`COMMONROOM_DEVELOPMENT_REPOSITORIES` maps room IDs to approved bundle SHA-256 values.
An alternative private `.data/development-workspaces.json` file contains:

```json
{"enabled":true,"repositories":{"engineering":"REVIEWED_64_HEX_BUNDLE_SHA256"}}
```

Use actual room IDs and replace the digest placeholder. The file must be a private,
owner-owned, single-link regular file, mode 0600. Explicit environment values
override file configuration. Invalid enabled configuration prevents startup.

The worker independently validates its `developmentEnabled` and
`developmentRepositories` configuration. Neither controller nor worker configuration
overrides the current credential-separation barrier. Room access, owner permissions,
emergency stops, audit integrity and mandatory isolation always apply.

Stop active work before configuration changes. Disabling development admission
preserves edits and history; it does not prove already-running containers stopped.
Changing a snapshot never authorizes adopting or resetting an old workspace.

## Prepare a reviewable source snapshot

1. Review the selected committed revision and its reachable history for secrets and
   malicious build instructions. A bundle can include historical tracked secrets.
2. Create a bundle outside the working tree and record its SHA-256:

   ```sh
   git bundle create /tmp/commonroom-repository.bundle HEAD
   shasum -a 256 /tmp/commonroom-repository.bundle
   ```

3. Transfer it through the approved administrative channel into a private worker
   directory. Keep parent directories worker-controlled and the bundle non-writable
   by group/others. Do not include credentials in filenames or configuration.
4. Configure the approved absolute path and matching digest under the intended room.
   This prepares a snapshot only; it does not enable public development execution.

Bundles contain committed reachable history, not uncommitted/ignored files or
working-tree credentials. Submodules and Git LFS downloads are not provisioned.
Do not grant general network access to work around missing dependencies; required
toolchains and dependencies belong in reviewed images.

## Per-turn verified snapshot staging

The legacy snapshot reader opens a no-follow, nonblocking descriptor, validates a
regular nonempty file and caps reads at 256 MiB plus one detection byte. It hashes a
private copy rather than allowing Docker to reopen an unchecked source path.
Special files, final symlinks and group/world-writable sources are refused.

The replacement bootstrap consumes a bounded checksummed stream into a fresh
workspace and enforces its own checkout limits. See [workspace bootstrap](workspace-bootstrap.md)
for the authoritative replacement contract. The old bind-staging implementation
is not permission to use the retired provider/tool co-location path.

## Verification and known gaps

Automated tests use temporary synthetic files and mocked Docker/database boundaries.
They cover snapshot bounds, permissions, initialization, paired lifecycle and audited
callbacks. No real provider, database or Docker workload is launched by those tests.
Historical single-container deployment checks do not validate the replacement design.

Follow the [remaining P0 acceptance plan](p0-acceptance-tests.md) for production
admission, provider adapters, commands, artifacts, quotas and disposal. Workspace
synchronization, automatic merging and host-checkout access are not provided.
