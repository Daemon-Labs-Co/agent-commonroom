const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const context=vm.createContext({render(){},Date,esc:s=>String(s).replaceAll('&','&amp;').replaceAll('<','&lt;').replaceAll('>','&gt;')});
vm.runInContext(fs.readFileSync('web/connection-testing.js','utf8'),context);
const result=vm.runInContext(`connectionTestMarkup({status:'needs-sign-in',checkedAt:'2026-10-11T00:00:00Z',checks:[{name:'<script>',status:'required',detail:'<img src=x>'}]})`,context);
assert.ok(result.includes('Sign-in needed'));assert.ok(result.includes('&lt;script&gt;'));assert.ok(!result.includes('<img'));
const ready=vm.runInContext(`connectionTestMarkup({status:'working',checkedAt:'2026-10-11T00:00:00Z',checks:[]})`,context);assert.ok(ready.includes('Connection working'));
console.log('Connection readiness labels and escaped diagnostic rendering passed.');

const failed=vm.runInContext(`connectionTestMarkup({status:'failed',checkedAt:'2026-10-11T00:00:00Z',checks:[]})`,context);assert.ok(failed.includes('Connection test failed'));
const unverified=vm.runInContext(`connectionTestMarkup({status:'ready',checkedAt:'2026-10-11T00:00:00Z',checks:[]})`,context);assert.ok(!unverified.includes('Connection working'));assert.ok(unverified.includes('provider unverified'));
assert.ok(unverified.includes('Sign in and test again'));assert.ok(result.includes('View log'));
assert.ok(result.indexOf('Sign-in needed')<result.indexOf('connection-check-list'));
assert.ok(result.indexOf('connection-check-list')<result.indexOf('View log'));
for(const status of ['needs-registration','blocked','unsupported','needs-sign-in']){
 const markup=context.connectionTestMarkup({status,checkedAt:'2026-10-11T00:00:00Z',checks:[]});assert.ok(!markup.includes('Connection working'));
}
console.log('Compact outcomes put next steps before statuses and optional escaped diagnostics.');
const logResult={status:'failed',checkedAt:'2026-10-11T00:00:00Z',checks:[{name:'<script>',status:'failed',detail:'<img src=x>\nExplanation'}]};
assert.ok(!context.connectionTestMarkup(logResult).includes('Explanation'));
const logMarkup=context.connectionTestLogMarkup(logResult);assert.ok(logMarkup.includes('&lt;img'));assert.ok(!logMarkup.includes('<script>'));assert.ok(logMarkup.includes('Explanation'));
let appended=0,removed=0,parentListener,dialog;
const parent={addEventListener(type,f){assert.equal(type,'close');parentListener=f},removeEventListener(type,f){assert.equal(f,parentListener);parentListener=null}};
context.state={currentActorId:'local-owner',featureFlags:{connectionTesting:true}};
context.document={body:{append(){appended++}},createElement(){const closeButton={};dialog={innerHTML:'',setAttribute(){},addEventListener(type,f){this.cleanup=f},querySelector(){return closeButton},showModal(){this.shown=true},close(){this.cleanup()},remove(){removed++},closeButton};return dialog}};
context.openConnectionTestLog(logResult,parent);assert.equal(appended,1);assert.equal(dialog.shown,true);assert.ok(dialog.innerHTML.includes('Close connection test log'));assert.ok(dialog.innerHTML.includes('&lt;img'));
dialog.closeButton.onclick();assert.equal(removed,1);assert.equal(parentListener,null);
context.openConnectionTestLog(logResult,parent);parentListener();assert.equal(removed,2);
context.state.featureFlags.connectionTesting=false;context.openConnectionTestLog(logResult,parent);assert.equal(appended,2);
context.state.featureFlags.connectionTesting=true;context.state.currentActorId='developer';context.openConnectionTestLog(logResult,parent);assert.equal(appended,2);
const logButton={};const output={querySelector(){return logButton}};context.state.currentActorId='local-owner';context.renderConnectionTestResult(output,logResult,parent);assert.ok(output.innerHTML.includes('View log'));logButton.onclick();assert.equal(appended,3);dialog.close();
console.log('Separate log overlay renders escaped diagnostics, routes the button, cleans up and enforces owner/flag guards.');

(async()=>{
 context.TextDecoder=TextDecoder;
 const encode=new TextEncoder();
 function response(text){let i=0;return {body:{getReader(){return {async read(){if(i>=text.length)return {done:true};return {done:false,value:text.slice(i,i+=3)}},async cancel(){},releaseLock(){}}}}}}
 const messages=[{type:'stage',check:{name:'Worker runtime',status:'running',detail:'Checking…'}},{type:'stage',check:{name:'Worker runtime',status:'passed',detail:'Ready'}},{type:'result',result:{status:'working'}}].map(x=>JSON.stringify(x)).join('\n')+'\n';
 const progress=[];const result=await context.readConnectionTestStream(response(encode.encode(messages)),c=>progress.push(c.status));assert.equal(result.status,'working');assert.deepEqual(progress,['running','passed']);
 await assert.rejects(()=>context.readConnectionTestStream(response(encode.encode('{}\n')),()=>{}),/interrupted/);
 await assert.rejects(()=>context.readConnectionTestStream(response(encode.encode('{"type":"error","message":"Audit unavailable"}\n')),()=>{}),/Audit unavailable/);
 const markup=context.connectionProgressMarkup([{name:'Worker runtime',status:'running',detail:'<unsafe>'}]);assert.ok(markup.includes('waiting'));assert.ok(markup.includes('running'));assert.ok(!markup.includes('<unsafe>'));
 console.log('Real progress stream handles chunked UTF-8, ordering, incomplete responses, errors and escaped stages.');
})().catch(e=>{console.error(e);process.exitCode=1});

for(const [status,tone] of [['passed','passed'],['failed','failed'],['blocked','failed'],['required','attention'],['running','running'],['waiting','neutral']]){assert.ok(context.connectionStatusMarkup(status).includes('connection-check-status '+tone));}
assert.ok(!context.connectionStatusMarkup('<script>').includes('<script>'));
for(const render of [context.connectionTestMarkup,context.connectionTestLogMarkup]){const html=render({status:'failed',checkedAt:'2026-10-11T00:00:00Z',checks:[{name:'Worker runtime',status:'failed',detail:'Stopped'}]});assert.ok(html.includes('connection-check-status failed'));}
assert.ok(context.connectionProgressMarkup([{name:'Worker runtime',status:'passed',detail:'Ready'}]).includes('connection-check-status passed'));
console.log('Passed, failed/blocked, required and waiting tones are consistent and escaped.');
