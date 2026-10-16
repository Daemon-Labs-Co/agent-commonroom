'use strict';
const {test} = require('node:test');
const assert = require('node:assert/strict');
const {EventEmitter} = require('node:events');
const {run} = require('../docker/engine/entrypoint.cjs');

function fixture(extra = Buffer.alloc(0)) {
  const raw = Buffer.from(JSON.stringify({auth_mode: 'chatgpt', tokens: {access_token: 'synthetic-only'}}));
  const proc = new EventEmitter();
  proc.platform = 'linux'; proc.getuid = () => 1000; proc.argv = ['node', '/opt/commonroom/engine-entrypoint.cjs'];
  proc.env = {OPENAI_API_KEY: 'synthetic-not-inherited', NODE_OPTIONS: '--require synthetic'};
  const child = new EventEmitter(), server = new EventEmitter();
  const calls = [], writes = [], directories = [], logs = [], signals = [], listeners = [];
  child.kill = signal => signals.push(signal);
  server.close = () => calls.push(['server-close']);
  server.listen = (port, host, ready) => { listeners.push([port, host]); ready(); };
  const length = Buffer.alloc(4); length.writeUInt32BE(raw.length);
  const framed = Buffer.concat([Buffer.from('COMMONROOM-AUTH-V1\n'), length, raw, extra]);
  let offset = 0;
  return {proc, child, server, calls, writes, directories, logs, signals, listeners, raw, framed, readOffset: () => offset,
    files: {existsSync: () => false,
      readSync: (fd, buffer, start, size) => {
        assert.equal(fd, 0);
        const n = Math.min(size, 3, framed.length - offset);
        framed.copy(buffer, start, offset, offset + n); offset += n; return n;
      },
      mkdirSync: (name, options) => directories.push([name, options]),
      writeFileSync: (name, data, options) => writes.push([name, data, options])},
    network: {createServer: () => server},
    processes: {spawn: (...args) => { calls.push(args); return child; }},
    stderr: {write: text => logs.push(text)}};
}
test('fixed subscription-only app-server in an exclusive controlled home', async () => {
  const f = fixture(), result = run(f);
  const [command, args, options] = f.calls[0];
  assert.equal(command, '/opt/commonroom/bin/codex');
  assert.deepEqual(args.slice(-3), ['app-server', '--strict-config', '--stdio']);
  assert.ok(args.includes('forced_login_method="chatgpt"'));
  assert.ok(args.includes('hooks')); assert.ok(args.includes('skip_host_skill_discovery'));
  assert.equal(options.shell, false); assert.deepEqual(options.stdio, ['inherit', 'inherit', 'ignore']);
  assert.equal(options.cwd, '/tmp/commonroom-engine');
  assert.equal(options.env.OPENAI_API_KEY, undefined); assert.equal(options.env.NODE_OPTIONS, undefined);
  assert.equal(options.env.CODEX_HOME, '/tmp/commonroom-engine/.codex');
  assert.equal(options.env.HTTPS_PROXY, 'http://127.0.0.1:18080');
  assert.deepEqual(f.listeners, [[18080, '127.0.0.1']]);
  assert.ok(f.directories.every(([, o]) => o.mode === 0o700 && o.recursive === undefined));
  assert.ok(f.writes.every(([, , o]) => o.flag === 'wx' && o.mode === 0o600));
  f.child.emit('close', 0, null);
  assert.equal(await result, 0); assert.equal(f.proc.listenerCount('SIGTERM'), 0);
});
test('unsafe startup state fails before gateway or process launch without raw errors', async () => {
  for (const mode of ['platform', 'arguments', 'configuration', 'bad frame', 'oversize auth', 'API login', 'API fallback', 'old home', 'file error', 'uid', 'truncated frame']) {
    const f = fixture();
    if (mode === 'platform') f.proc.platform = 'darwin';
 if (mode === 'uid') f.proc.getuid = () => 0;
 if (mode === 'truncated frame') f.files.readSync = () => 0;
    if (mode === 'arguments') f.proc.argv.push('--unsafe');
    if (mode === 'configuration') f.files.existsSync = name => name === '/etc/codex';
    if (mode === 'bad frame') f.framed[0] = 0;
    if (mode === 'oversize auth') f.framed.writeUInt32BE(65537, Buffer.byteLength('COMMONROOM-AUTH-V1\n'));
    if (mode === 'API login' || mode === 'API fallback') {
      const raw = Buffer.from(JSON.stringify({auth_mode: mode === 'API login' ? 'apikey' : 'chatgpt', OPENAI_API_KEY: 'synthetic-private', tokens: {access_token: 'synthetic'}}));
      const header = Buffer.alloc(4); header.writeUInt32BE(raw.length);
      const framed = Buffer.concat([Buffer.from('COMMONROOM-AUTH-V1\n'), header, raw]);
      let offset = 0;
      f.files.readSync = (fd, buffer, start, size) => { const n = Math.min(size, framed.length-offset); framed.copy(buffer,start,offset,offset+n); offset+=n; return n; };
    }
    if (mode === 'old home') f.files.mkdirSync = () => { throw new Error('synthetic-private'); };
    if (mode === 'file error') f.files.writeFileSync = () => { throw new Error('synthetic-private'); };
    assert.equal(await run(f), 70, mode);
    assert.equal(f.calls.length, 0, mode); assert.equal(f.listeners.length, 0, mode);
    assert.deepEqual(f.logs, ['Reviewed engine startup refused\n']);
  }
});
test('gateway and spawn errors fail closed, signals forward, resources close', async () => {
  for (const mode of ['gateway', 'spawn', 'signal', 'child error']) {
    const f = fixture();
    if (mode === 'spawn') f.processes.spawn = () => { throw new Error('synthetic-private'); };
    if (mode === 'gateway') f.server.listen = () => { f.server.emit('error', new Error('synthetic-private')); };
    const result = run(f);
    if (mode === 'signal') { f.proc.emit('SIGTERM'); f.child.emit('close', null, 'SIGTERM'); }
    if (mode === 'child error') f.child.emit('error', new Error('synthetic-private'));
    assert.equal(await result, 70);
    if (mode === 'signal') assert.deepEqual(f.signals, ['SIGTERM']);
    assert.equal(f.proc.listenerCount('SIGTERM'), 0);
    assert.ok(f.logs.every(line => !line.includes('synthetic-private')));
  }
});

test('engine image requires reviewed immutable inputs and a restricted build context', () => {
  const fs = require('node:fs');
  const recipe = fs.readFileSync(require('node:path').join(__dirname, '../docker/engine/Dockerfile'), 'utf8');
  const context = fs.readFileSync(require('node:path').join(__dirname, '../docker/engine/Dockerfile.dockerignore'), 'utf8');
  assert.ok(recipe.includes('FROM ${ENGINE_SOURCE_IMAGE} AS reviewed_engine'));
  assert.ok(recipe.includes('FROM ${NODE_IMAGE}'));
  assert.ok(recipe.includes('CODEX_BINARY_SHA256'));
  assert.ok(recipe.includes('sha256sum --check --strict -'));
  assert.ok(recipe.includes('USER 1000:1000'));
  assert.ok(recipe.includes('/out/LICENSE'));
  assert.ok(context.split('\n').includes('**'));
  assert.deepEqual(context.split('\n').filter(line => line.startsWith('!')), [
    '!docker/', '!docker/engine/', '!docker/engine/Dockerfile',
    '!docker/engine/Dockerfile.dockerignore', '!docker/engine/entrypoint.cjs'
  ]);
  assert.ok(!/npm install|curl |wget |COPY \.|ADD /u.test(recipe));
});

test('gateway connections target only the mounted Unix socket and close with the child', async () => {
  const f = fixture();
  let accept;
  f.network.createServer = handler => { accept = handler; return f.server; };
  function socket() {
    const value = new EventEmitter(); value.destroyed = false; value.targets = [];
    value.pipe = other => { value.targets.push(other); return other; };
    value.destroy = () => { if (!value.destroyed) { value.destroyed = true; value.emit('close'); } };
    return value;
  }
  const client = socket(), remote = socket(), paths = [];
  f.network.connect = name => { paths.push(name); return remote; };
  const result = run(f);
  accept(client);
  assert.deepEqual(paths, ['/run/provider.sock']);
  assert.deepEqual(client.targets, [remote]); assert.deepEqual(remote.targets, [client]);
  f.child.emit('close', 0, null);
  assert.equal(await result, 0);
  assert.equal(client.destroyed, true); assert.equal(remote.destroyed, true);
});


test('credential framing consumes no following protocol bytes', async () => {
  const protocol = Buffer.from('{"id":1,"method":"initialize"}\n');
  const f = fixture(protocol), result = run(f);
  const consumed = f.readOffset();
  assert.equal(consumed, Buffer.byteLength('COMMONROOM-AUTH-V1\n') + 4 + f.raw.length);
  assert.deepEqual(f.framed.subarray(consumed), protocol);
  f.child.emit('close', 0, null);
  assert.equal(await result, 0);
});

test('credential frame short payload and invalid sizes fail without process launch', async () => {
  for (const mode of ['zero', 'short payload', 'read error']) {
    const f = fixture();
    if (mode === 'zero') f.framed.writeUInt32BE(0, Buffer.byteLength('COMMONROOM-AUTH-V1\n'));
    if (mode === 'short payload') f.framed.writeUInt32BE(f.raw.length + 1, Buffer.byteLength('COMMONROOM-AUTH-V1\n'));
    if (mode === 'read error') f.files.readSync = () => { throw new Error('synthetic-private'); };
    assert.equal(await run(f), 70, mode);
    assert.equal(f.calls.length, 0);
    assert.deepEqual(f.logs, ['Reviewed engine startup refused\n']);
  }
});
