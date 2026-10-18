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
  return {exportFactory: () => ({requested: () => false, complete: () => false, close() {}}), proc, child, server, calls, writes, directories, logs, signals, listeners, raw, framed, readOffset: () => offset,
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

function exportFixture(mode='success') {
 const {credentialExporter}=require('../docker/engine/entrypoint.cjs');
 const raw=Buffer.from(JSON.stringify({auth_mode:mode==='API credential'?'apikey':'chatgpt',tokens:{access_token:'synthetic-refreshed'}}));
 if(mode==='invalid JSON')raw.fill(120);
 const server=new EventEmitter(),child=new EventEmitter(),client=new EventEmitter();
 let accept,offset=0,failures=0,closed=0,output;
 const signals=[];
 server.listen=path=>assert.equal(path,'/tmp/commonroom-engine/credential-export.sock');
 server.close=()=>{closed++};
 child.kill=signal=>{signals.push(signal);return mode!=='kill failure'};
 client.setTimeout=(ms,handler)=>{assert.equal(ms,10000);client.timeout=handler};
 client.destroy=()=>{client.destroyed=true};
 client.end=frame=>{output=frame};
 const files={constants:{O_RDONLY:0,O_NOFOLLOW:1,O_NONBLOCK:2},
  openSync:(path,flags)=>{assert.equal(path,'/tmp/commonroom-engine/.codex/auth.json');assert.equal(flags,3);return 7},
  fstatSync:()=>({isFile:()=>mode!=='nonregular',uid:mode==='owner'?0:1000,nlink:mode==='hardlink'?2:1,mode:mode==='permissions'?0o644:0o600,size:mode==='oversize'?65537:raw.length}),
  readSync:(fd,buffer,start,size)=>{assert.equal(fd,7);const n=Math.min(size,3,raw.length-offset);raw.copy(buffer,start,offset,offset+n);offset+=n;return n},closeSync:fd=>assert.equal(fd,7)};
 const exporter=credentialExporter({files,network:{createServer:(options,handler)=>{assert.equal(options.allowHalfOpen,true);accept=handler;return server}},child,fail:()=>{failures++}});
 accept(client);
 return {exporter,child,client,raw,signals,files,server,accept,get output(){return output},get failures(){return failures},get closed(){return closed}};
}
test('private export stops the child before reading final private credentials',()=>{
 const f=exportFixture(),magic=Buffer.from('COMMONROOM-CREDENTIAL-EXPORT-V1\n');
 f.client.emit('data',magic.subarray(0,5));f.client.emit('data',magic.subarray(5));f.client.emit('end');
 assert.deepEqual(f.signals,['SIGTERM']);assert.equal(f.output,undefined);assert.equal(f.exporter.requested(),true);
 f.child.emit('close',null,'SIGTERM');
 assert.equal(f.failures,0);assert.equal(f.exporter.complete(),true);
 assert.deepEqual(f.output.subarray(0,magic.length),magic);
 assert.equal(f.output.readUInt32BE(magic.length),f.raw.length);
 assert.deepEqual(f.output.subarray(magic.length+4),f.raw);
 const second=new EventEmitter();second.destroy=()=>{second.destroyed=true};f.accept(second);assert.equal(second.destroyed,true);
 f.exporter.close();assert.equal(f.closed,1);
});
test('unsafe credential inode, invalid request, child refusal and timeout cannot export',()=>{
 for(const mode of ['nonregular','owner','hardlink','permissions','oversize','kill failure','bad request','timeout','invalid JSON','API credential']){
  const f=exportFixture(mode);
  f.client.emit('data',Buffer.from(mode==='bad request'?'wrong request':'COMMONROOM-CREDENTIAL-EXPORT-V1\n'));
  f.client.emit('end');
  if(mode==='timeout')f.client.timeout();else if(mode!=='kill failure'&&mode!=='bad request')f.child.emit('close',null,'SIGTERM');
  assert.ok(f.failures>0,mode);assert.equal(f.output,undefined,mode);f.exporter.close();
 }
});
test('provider exit alone does not export credentials',()=>{
 const f=exportFixture();f.child.emit('close',0,null);assert.equal(f.output,undefined);assert.equal(f.exporter.requested(),false);f.exporter.close();
});

test('bootstrap stays alive after private child export until supervised termination',async()=>{
 const f=fixture();delete f.exportFactory;
 const exportServer=new EventEmitter();let accept;
 exportServer.listen=path=>assert.equal(path,'/tmp/commonroom-engine/credential-export.sock');exportServer.close=()=>{};
 f.network.createServer=(options,handler)=>{if(typeof options==='function')return f.server;accept=handler;return exportServer};
 f.child.kill=signal=>{f.signals.push(signal);return true};
 const readAuth=f.files.readSync;let offset=0;
 f.files.constants={O_RDONLY:0,O_NOFOLLOW:1,O_NONBLOCK:2};
 f.files.openSync=()=>7;f.files.fstatSync=()=>({isFile:()=>true,uid:1000,nlink:1,mode:0o600,size:f.raw.length});
 f.files.closeSync=()=>{};
 f.files.readSync=(fd,buffer,start,size)=>{if(fd===0)return readAuth(fd,buffer,start,size);const n=Math.min(size,f.raw.length-offset);f.raw.copy(buffer,start,offset,offset+n);offset+=n;return n};
 let resolved=false;const result=run(f).then(code=>{resolved=true;return code});
 const client=new EventEmitter();client.setTimeout=()=>{};client.destroy=()=>{};client.end=frame=>{client.output=frame};accept(client);
 client.emit('data',Buffer.from('COMMONROOM-CREDENTIAL-EXPORT-V1\n'));client.emit('end');f.child.emit('close',null,'SIGTERM');
 await Promise.resolve();assert.equal(resolved,false);assert.ok(client.output.includes(f.raw));assert.equal(f.logs.length,0);
 f.proc.emit('SIGTERM');assert.equal(await result,0);
});

test('private worker helper rejects truncated, oversized and replayed export frames',()=>{
 const vm=require('node:vm'),fs=require('node:fs');
 const source=fs.readFileSync(require('node:path').join(__dirname,'../internal/toolbroker/engine_credentials.go'),'utf8');
 const script=source.match(/const engineCredentialExportScript = `([\s\S]*?)`/u)[1];
 for(const mode of ['success','wrong magic','short','trailing','oversize','timeout','disconnect']){
  const socket=new EventEmitter(),stdout=new EventEmitter(),chunks=[];let timeout;
  socket.end=request=>assert.equal(request.toString(),'COMMONROOM-CREDENTIAL-EXPORT-V1\n');
  socket.destroy=()=>socket.emit('close');stdout.write=(bytes,cb)=>{chunks.push(Buffer.from(bytes));cb()};
  const proc={stdout};
  vm.runInNewContext(script,{Buffer,process:proc,require:name=>{assert.equal(name,'node:net');return {createConnection:path=>{assert.equal(path,'/tmp/commonroom-engine/credential-export.sock');return socket}}},setTimeout:handler=>{timeout=handler;return 1},clearTimeout:()=>{}});
  socket.emit('connect');
  const raw=Buffer.from('{"synthetic":"credential"}'),size=Buffer.alloc(4);size.writeUInt32BE(raw.length);
  let frame=Buffer.concat([Buffer.from('COMMONROOM-CREDENTIAL-EXPORT-V1\n'),size,raw]);
  if(mode==='wrong magic')frame[0]=0;
  if(mode==='short')frame=frame.subarray(0,frame.length-1);
  if(mode==='trailing')frame=Buffer.concat([frame,Buffer.from('x')]);
  if(mode==='oversize')frame=Buffer.alloc(70000);
  if(mode==='timeout')timeout();else if(mode==='disconnect')socket.emit('close');else{socket.emit('data',frame);socket.emit('end')}
  if(mode==='success'){assert.deepEqual(chunks,[raw]);assert.equal(proc.exitCode,undefined)}else{assert.equal(chunks.length,0,mode);assert.equal(proc.exitCode,78,mode)}
 }
});
