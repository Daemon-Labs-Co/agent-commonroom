'use strict';
const fs = require('node:fs');
const cp = require('node:child_process');
const path = require('node:path');
const {randomUUID} = require('node:crypto');

// Bootstrap only inside the mandatory container. The reviewed bundle is mounted
// read-only and cloned into the membership's private work volume. No host Git
// configuration, hooks, credentials or working tree are imported. A fork CAN
// mount host directories here, BUT would invalidate the containment contract.
function prepareWorkspace(env, options = {}) {
  const digest = env.COMMONROOM_DEVELOPMENT_DIGEST;
  if (!digest) return null;
  const branch = env.COMMONROOM_DEVELOPMENT_BRANCH;
  if (!/^[a-f0-9]{64}$/.test(digest) || !/^feature\/commonroom-session-[a-f0-9]{32}$/.test(branch || '')) throw new Error('Invalid development bootstrap policy');
  const files = options.fs || fs, processes = options.cp || cp;
  const root = options.root || '/workspace';
  const repository = path.join(root, 'repository');
  const marker = path.join(root, '.development-snapshot');
  const gitEnv = {...env, GIT_CONFIG_NOSYSTEM: '1', GIT_CONFIG_GLOBAL: '/dev/null', GIT_TERMINAL_PROMPT: '0'};
  const execute = (command, args, cwd = root) => {
    const result = processes.spawnSync(command, args, {cwd, env: gitEnv, encoding: 'utf8', timeout: 20000, maxBuffer: 65536, shell: false});
    if (result.error || result.status !== 0) throw new Error(`Development readiness failed: ${command}`);
    return String(result.stdout || '').trim();
  };
  const git = (args, cwd) => execute('git', ['-c', 'core.hooksPath=/dev/null', ...args], cwd);
  execute('sh', ['-c', 'exit 0']);
  execute('git', ['--version']);
  execute('go', ['version']);
  execute('node', ['--version']);
  execute('make', ['--version']);
  if (!files.existsSync(marker)) {
    if (files.existsSync(repository)) throw new Error('Incomplete workspace setup; operator recovery required');
    git(['clone', '--quiet', '--', '/run/repository.bundle', repository]);
    git(['remote', 'remove', 'origin'], repository);
    git(['switch', '-c', branch], repository);
    files.writeFileSync(marker, digest, {flag: 'wx', mode: 0o600});
  }
  if (files.readFileSync(marker, 'utf8') !== digest) throw new Error('Workspace snapshot changed; a fresh workspace is required');
  if (git(['branch', '--show-current'], repository) !== branch) throw new Error('Development requires the assigned feature branch');
  git(['rev-parse', '--verify', 'HEAD'], repository);
  // Verify both ordinary edits and .git metadata are writable before a prompt is
  // sent. Clean up probes on all paths. No commit or provider turn is involved.
  for (const directory of [repository, path.join(repository, '.git')]) {
    const probe = path.join(directory, '.commonroom-write-probe-' + randomUUID());
    let created = false;
    try { files.writeFileSync(probe, 'ready', {flag: 'wx', mode: 0o600}); created = true; }
    finally { if (created) files.unlinkSync(probe); }
  }
  files.mkdirSync(path.join(root, '.tmp'), {recursive: true, mode: 0o700});
  return {cwd: repository, env: {...gitEnv, GOCACHE: '/home/agent/.cache/go-build', GOMODCACHE: '/opt/commonroom/go-modules', GOTOOLCHAIN: 'local', GOPROXY: 'off', GOSUMDB: 'off', GOTMPDIR: path.join(root, '.tmp')}};
}
module.exports = {prepareWorkspace};
