'use strict';
const {test} = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs'), os = require('node:os'), path = require('node:path');
const {EventEmitter} = require('node:events');
const {providerInvocation, run} = require('../docker/agent/entrypoint.cjs');

test('Codex login and resumed turns share the worker export location and file store', () => {
  const env = {HOME: '/home/node', CODEX_HOME: '/wrong/store', HTTPS_PROXY: 'http://127.0.0.1:18080'};
  const login = providerInvocation('codex', ['auth-login'], env);
  const turn = providerInvocation('codex', ['exec', 'resume', 'fixture-session', 'hello'], env);
  assert.equal(login.credential, '/home/agent/.codex/auth.json');
  assert.equal(login.env.CODEX_HOME, '/home/agent/.codex');
  assert.equal(login.env.HOME, '/home/agent');
  assert.equal(login.env.HTTPS_PROXY, env.HTTPS_PROXY);
  assert.equal(turn.env.CODEX_HOME, login.env.CODEX_HOME);
  assert.ok(login.args.includes('cli_auth_credentials_store="file"'));
  assert.deepEqual(login.args.slice(-2), ['login', '--device-auth']);
  assert.deepEqual(turn.args.slice(-4), ['exec', 'resume', 'fixture-session', 'hello']);
  assert.equal(login.args[login.args.indexOf('respect_system_proxy')-1], '--disable');
  assert.equal(login.args[login.args.indexOf('system_proxy_fallback')-1], '--disable');
  assert.equal(env.CODEX_HOME, '/wrong/store');
});

test('Claude login and turns use the same isolated configuration directory', () => {
  const login = providerInvocation('claude', ['auth-login'], {CLAUDE_CONFIG_DIR: '/wrong/store'});
  const turn = providerInvocation('claude', ['-p', 'fixture'], {});
  assert.equal(login.env.CLAUDE_CONFIG_DIR, '/home/agent/.claude');
  assert.equal(turn.env.CLAUDE_CONFIG_DIR, login.env.CLAUDE_CONFIG_DIR);
  assert.equal(login.credential, '/home/agent/.claude/.credentials.json');
  assert.deepEqual(login.args, ['auth', 'login', '--claudeai']);
  assert.throws(() => providerInvocation('shell', [], {}));
});

// A mocked CLI writes a synthetic file using the actual launch environment. This
// exercises login-to-export behavior without a provider, Docker or database.
for (const kind of ['valid', 'missing', 'symlink', 'oversized', 'provider-failure']) {
  test(`login handoff: ${kind}`, t => {
    const home = fs.mkdtempSync(path.join(os.tmpdir(), 'commonroom-login-test-'));
    t.after(() => fs.rmSync(home, {recursive: true, force: true}));
    const child = new EventEmitter(), server = new EventEmitter();
    server.listen = (port, host, ready) => { assert.equal(host, '127.0.0.1'); ready(); };
    server.close = () => {};
    let exit;
    const plan = run('codex', ['auth-login'], {home, env: {CODEX_HOME: '/wrong'}, fs,
      net: {createServer: () => server}, exit: code => { exit = code; },
      cp: {spawn: (provider, args, options) => {
        assert.equal(options.shell, false);
        const file = path.join(options.env.CODEX_HOME, 'auth.json');
        if (kind === 'valid') fs.writeFileSync(file, '{"fixture":true}', {mode: 0o644});
        if (kind === 'oversized') fs.writeFileSync(file, 'x'.repeat(65537));
        if (kind === 'symlink') { const target = path.join(home, 'other'); fs.writeFileSync(target, '{"fixture":true}'); fs.symlinkSync(target, file); }
        return child;
      }}});
    child.emit('exit', kind === 'provider-failure' ? 3 : 0);
    assert.equal(exit, kind === 'valid' ? 0 : kind === 'provider-failure' ? 3 : 78);
    if (kind === 'valid') assert.equal(fs.statSync(plan.credential).mode & 0o777, 0o600);
  });
}
