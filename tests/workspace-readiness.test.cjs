'use strict';
const {test} = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs'), os = require('node:os'), path = require('node:path');
const {prepareWorkspace} = require('../docker/agent/workspace.cjs');
const {providerInvocation, run} = require('../docker/agent/entrypoint.cjs');
const digest = 'a'.repeat(64), branch = 'feature/commonroom-session-' + 'b'.repeat(32);
const env = {COMMONROOM_DEVELOPMENT_DIGEST: digest, COMMONROOM_DEVELOPMENT_BRANCH: branch};
function fixture(t, failCommand) {
 const root = fs.mkdtempSync(path.join(os.tmpdir(), 'commonroom-workspace-'));
 t.after(() => fs.rmSync(root, {recursive:true, force:true}));
 const calls = [];
 const cp = {spawnSync(command, args, options) {
  calls.push({command,args,options});
  assert.equal(options.shell, false);
  assert.equal(options.env.GIT_CONFIG_GLOBAL, '/dev/null');
  if (failCommand === command) return {status:1, stdout:''};
  if (args.includes('clone')) fs.mkdirSync(path.join(root,'repository','.git'),{recursive:true});
  return {status:0, stdout: args.includes('--show-current') ? branch : 'fixture-head'};
 }};
 return {root,cp,calls};
}
test('disabled workspaces do not execute commands or touch files', () => {
 assert.equal(prepareWorkspace({}, {fs:{},cp:{}}), null);
});
test('new workspace gets a private feature branch and writable code/git metadata', t => {
 const options = fixture(t);
 const result = prepareWorkspace(env, options);
 assert.equal(result.cwd, path.join(options.root,'repository'));
 assert.ok(options.calls.some(c => c.args.includes('clone')));
 assert.ok(options.calls.some(c => c.args.includes('switch') && c.args.includes(branch)));
 assert.ok(options.calls.some(c => c.args.includes('remove') && c.args.includes('origin')));
 assert.equal(fs.readFileSync(path.join(options.root,'.development-snapshot'),'utf8'),digest);
 assert.equal(result.env.GOPROXY,'off');
 assert.equal(result.env.GOTOOLCHAIN,'local');
 options.calls.length=0;
 fs.writeFileSync(path.join(result.cwd,'work.txt'),'preserve edits');
 prepareWorkspace(env, options);
 assert.ok(!options.calls.some(c=>c.args.includes('clone') || c.args.includes('switch')));
 assert.equal(fs.readFileSync(path.join(result.cwd,'work.txt'),'utf8'),'preserve edits');
});
for (const failed of ['sh','git','go','node','make']) test(`unavailable ${failed} blocks development before provider launch`, t=> {
 assert.throws(()=>prepareWorkspace(env,fixture(t,failed)), /readiness failed/);
});
test('snapshot change, wrong branch and incomplete setup fail closed', t => {
 const options=fixture(t);
 prepareWorkspace(env, options);
 assert.throws(()=>prepareWorkspace({...env, COMMONROOM_DEVELOPMENT_DIGEST:'c'.repeat(64)},options),/snapshot changed/);
 const wrong={...options,cp:{spawnSync(){return {status:0,stdout:'main'}}}};
 assert.throws(()=>prepareWorkspace(env,wrong),/feature branch/);
 fs.unlinkSync(path.join(options.root,'.development-snapshot'));
 assert.throws(()=>prepareWorkspace(env,options),/Incomplete/);
 assert.throws(()=>prepareWorkspace({...env,COMMONROOM_DEVELOPMENT_BRANCH:'main'},options),/bootstrap policy/);
});
test('read-only code or git metadata fails and removes write probes', t=> {
 for(const denyGit of [false,true]) {
  const options=fixture(t);
  const files={...fs,writeFileSync(file,...args){
   if(file.includes('.commonroom-write-probe-') && (file.includes('/.git/')===denyGit)) throw new Error('EACCES');
   return fs.writeFileSync(file,...args);
  }};
  assert.throws(()=>prepareWorkspace(env,{...options,fs:files}),/EACCES/);
  assert.equal(fs.existsSync(path.join(options.root,'repository','.commonroom-write-probe')),false);
 }
});
test('development Codex uses outer container backend without changing chat or login policies',()=> {
 const development=providerInvocation('codex',['exec','-s','danger-full-access'],{});
 assert.ok(development.args.includes('danger-full-access'));
 assert.ok(!development.args.includes('--dangerously-bypass-approvals-and-sandbox'));
 const chat=providerInvocation('codex',['exec','-s','read-only'],{});
 assert.ok(chat.args.includes('read-only'));
});

test('failed workspace preflight never copies credentials or launches a provider', () => {
 let copied=false, launched=false;
 const files={copyFileSync(){copied=true}};
 const cp={spawnSync(){return {status:1}},spawn(){launched=true;throw new Error('unexpected provider')}};
 assert.throws(()=>run('codex',['exec'],{env,fs:files,cp,net:{},exit(){}}),/readiness failed/);
 assert.equal(copied,false); assert.equal(launched,false);
});
