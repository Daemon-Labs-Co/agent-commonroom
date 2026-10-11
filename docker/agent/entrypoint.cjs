'use strict';
const net = require('node:net');
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const {prepareWorkspace} = require('./workspace.cjs');

// Provider storage must match the worker's export paths. An image's passwd home,
// inherited CODEX_HOME, or OS keyring must never redirect a connection's login.
function providerInvocation(provider, inputArgs, inheritedEnv, home = '/home/agent') {
  if (!['codex', 'claude'].includes(provider)) throw new Error('Unsupported provider');
  const login = inputArgs[0] === 'auth-login';
  const dir = path.join(home, provider === 'codex' ? '.codex' : '.claude');
  const env = {...inheritedEnv, HOME: home};
  let args;
  if (provider === 'codex') {
    env.CODEX_HOME = dir;
    // Use the fixed gateway proxy rather than platform proxy discovery. The
    // container remains network-disabled; these flags cannot bypass isolation.
    args = ['--disable', 'respect_system_proxy', '--disable', 'system_proxy_fallback',
      '-c', 'cli_auth_credentials_store="file"',
      ...(login ? ['login', '--device-auth'] : inputArgs)];
  } else {
    env.CLAUDE_CONFIG_DIR = dir;
    args = login ? ['auth', 'login', '--claudeai'] : [...inputArgs];
  }
  return {provider, args, env, dir, login,
    credential: path.join(dir, provider === 'codex' ? 'auth.json' : '.credentials.json')};
}

// The loopback bridge exposes only the per-turn, provider-allowlisted Unix
// gateway. It is not a Docker socket. A fork CAN remove this boundary, BUT would
// expose host/network authority and invalidate the supported isolation contract.
function run(provider, args, options = {}) {
  const files = options.fs || fs, network = options.net || net, processes = options.cp || cp;
  const exit = options.exit || (code => process.exit(code));
  const plan = providerInvocation(provider, args, options.env || process.env, options.home);
  // Readiness is mandatory for configured development turns and runs before
  // credentials are copied or the provider is started. Login remains separate.
  const workspace = !plan.login ? prepareWorkspace(plan.env, {fs: files, cp: processes}) : null;
  if (workspace) plan.env = workspace.env;
  files.mkdirSync(plan.dir, {recursive: true, mode: 0o700});
  if (!plan.login) {
    files.copyFileSync('/run/provider-auth.json', plan.credential);
    files.chmodSync(plan.credential, 0o600);
  }
  const server = network.createServer(client => {
    const remote = network.connect('/run/provider.sock');
    client.pipe(remote); remote.pipe(client);
    client.on('error', () => remote.destroy()); remote.on('error', () => client.destroy());
    client.on('close', () => remote.destroy());
  });
  server.on('error', () => exit(70));
  server.listen(18080, '127.0.0.1', () => {
    const child = processes.spawn(plan.provider, plan.args, {stdio: 'inherit', shell: false, env: plan.env, ...(workspace ? {cwd: workspace.cwd} : {})});
    child.on('error', () => { server.close(); exit(70); });
    child.on('exit', code => {
      server.close();
      if (code === 0 && plan.login) {
        // Report failure before the worker attempts export if the provider chose
        // an unsupported store. Never print the credential or its contents.
        try {
          const info = files.lstatSync(plan.credential);
          if (!info.isFile() || info.size <= 2 || info.size > 65536) return exit(78);
          files.chmodSync(plan.credential, 0o600);
        } catch { return exit(78); }
      }
      exit(code ?? 70);
    });
    if (!options.exit) for (const signal of ['SIGTERM', 'SIGINT']) process.on(signal, () => child.kill(signal));
  });
  return plan;
}
if (require.main === module) {
  try { run(process.argv[2], process.argv.slice(3)); }
  catch (error) { console.error("Isolated agent startup failed:", error.message); process.exit(64); }
}
module.exports = {providerInvocation, run};
