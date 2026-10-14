'use strict';
const {test} = require('node:test');
const assert = require('node:assert/strict');
const {EventEmitter} = require('node:events');
const {run} = require('../docker/tools/tool-call.cjs');

function fixture() {
  const proc = new EventEmitter();
  proc.platform = 'linux'; proc.argv = ['node', '/opt/commonroom/tool-call.cjs'];
  proc.env = {SYNTHETIC_SECRET: 'do-not-inherit', NODE_OPTIONS: '--require synthetic', HTTP_PROXY: 'synthetic'};
  const child = new EventEmitter(), calls = [], signals = [], messages = [];
  child.kill = signal => { signals.push(signal); };
  return {proc, child, calls, signals, messages,
    cp: {spawn(...args) { calls.push(args); return child; }},
    stderr: {write(message) { messages.push(message); }}
  };
}

test('fixed executable and isolated environment, with no shell or dynamic arguments', async () => {
  const f = fixture(), result = run(f);
  assert.equal(f.calls.length, 1);
  const [command, args, options] = f.calls[0];
  assert.equal(command, '/opt/commonroom/tool-helper');
  assert.deepEqual(args, []);
  assert.equal(options.shell, false);
  assert.equal(options.cwd, '/workspace/repository');
  assert.equal(options.stdio, 'inherit');
  assert.deepEqual(options.env, {
    HOME: '/tmp/tool-home', PATH: '/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin',
    HTTP_PROXY: '', HTTPS_PROXY: '', ALL_PROXY: '', NO_PROXY: '',
    NODE_OPTIONS: '', NODE_USE_ENV_PROXY: '0', NODE_VERSION: '', YARN_VERSION: ''
  });
  f.child.emit('close', 0, null);
  assert.equal(await result, 0);
  assert.equal(f.proc.listenerCount('SIGTERM'), 0);
});

test('unsupported host or caller arguments refused before process launch', async () => {
  for (const mode of ['platform', 'argv']) {
    const f = fixture();
    if (mode === 'platform') f.proc.platform = 'darwin';
    else f.proc.argv.push('--root=/synthetic');
    assert.equal(await run(f), 2);
    assert.deepEqual(f.calls, []);
  }
});

test('helper failure and ambiguous write codes preserved, unknown codes refused', async () => {
  for (const [code, signal, expected] of [[2,null,2],[3,null,3],[99,null,2],[0,'SIGTERM',2],[null,null,2]]) {
    const f = fixture(), result = run(f);
    f.child.emit('close', code, signal);
    assert.equal(await result, expected);
  }
});

test('launch errors do not leak dependency diagnostics or retain signal handlers', async () => {
  for (const synchronous of [true, false]) {
    const f = fixture();
    if (synchronous) f.cp.spawn = () => { throw new Error('synthetic secret detail'); };
    const result = run(f);
    if (!synchronous) f.child.emit('error', new Error('synthetic secret detail'));
    assert.equal(await result, 2);
    assert.deepEqual(f.messages, ['Tool helper launch refused\n']);
    assert.equal(f.proc.listenerCount('SIGINT'), 0);
    assert.equal(f.proc.listenerCount('SIGTERM'), 0);
  }
});

test('signals forwarded to helper but completion requires child close', async () => {
  const f = fixture(), result = run(f);
  let finished = false; result.then(() => { finished = true; });
  f.proc.emit('SIGINT'); f.proc.emit('SIGTERM');
  await Promise.resolve();
  assert.equal(finished, false);
  assert.deepEqual(f.signals, ['SIGINT','SIGTERM']);
  f.child.emit('close', null, 'SIGTERM');
  assert.equal(await result, 2);
  assert.equal(f.proc.listenerCount('SIGTERM'), 0);
});
