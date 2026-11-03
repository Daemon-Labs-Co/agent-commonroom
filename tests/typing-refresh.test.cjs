const test=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
function fixture(){
 let resolve,requests=0,renders=0;
 const source=fs.readFileSync('web/app.js','utf8');
 const start=source.indexOf('function backgroundRefreshBlocked');
 const snippet=source.slice(start<0?source.indexOf('async function refresh'):start,source.indexOf('async function commit'));
 const field={matches:()=>true,closest:()=>null,isContentEditable:false};
 const main={contains:node=>node===field};
 const ctx={state:{version:1},view:'Room',roomId:'engineering',document:{activeElement:null,querySelector:s=>s==='#main'?main:null},api:()=>{requests++;return new Promise(r=>resolve=r)},render(){renders++}};
 vm.createContext(ctx);vm.runInContext(snippet,ctx);
 return {ctx,field,requests:()=>requests,renders:()=>renders,reply:()=>resolve({version:2})};
}
test('background refresh leaves a focused editor intact',async()=>{
 const f=fixture();f.ctx.document.activeElement=f.field;
 const pending=f.ctx.refresh({background:true});if(f.requests())f.reply();await pending;
 assert.equal(f.renders(),0);assert.equal(f.requests(),0);
});
test('typing begun during an in-flight background request is not interrupted',async()=>{
 const f=fixture();const pending=f.ctx.refresh({background:true});f.ctx.document.activeElement=f.field;f.reply();await pending;
 assert.equal(f.renders(),0);assert.equal(f.ctx.state.version,1);
 f.ctx.document.activeElement=null;const next=f.ctx.refresh({background:true});f.reply();await next;assert.equal(f.renders(),1);
});
test('background requests do not overlap or apply results after room navigation',async()=>{
 const f=fixture();const first=f.ctx.refresh({background:true});const second=f.ctx.refresh({background:true});
 assert.equal(f.requests(),1);f.ctx.roomId='qa';f.reply();await Promise.all([first,second]);assert.equal(f.renders(),0);
});
test('explicit user actions may refresh even when the old editor is focused',async()=>{
 const f=fixture();f.ctx.document.activeElement=f.field;const pending=f.ctx.refresh();f.reply();await pending;assert.equal(f.renders(),1);
});

test('editable content and open forms block background painting',async()=>{
 for(const modalOpen of [false,true]){
  const f=fixture();f.field.isContentEditable=true;f.field.matches=()=>false;
  if(modalOpen){const old=f.ctx.document.querySelector;f.ctx.document.querySelector=s=>s==='#modal'?{open:true,querySelector:()=>({})}:old(s)}else f.ctx.document.activeElement=f.field;
  await f.ctx.refresh({background:true});assert.equal(f.requests(),0);
 }
});
test('a failed background request releases the polling slot',async()=>{
 const f=fixture();f.ctx.api=async()=>{throw Error('Synthetic unavailable')};await assert.rejects(f.ctx.refresh({background:true}),/Synthetic unavailable/);
 f.ctx.api=async()=>({version:3});await f.ctx.refresh({background:true});assert.equal(f.renders(),1);
});
test('the actual status timer uses protected background refresh',async()=>{
 let tick,options;
 const ctx={state:{participants:[{status:'Working'}]},view:'Room',document:{querySelector:()=>({open:false})},setInterval(fn){tick=fn},refresh:async value=>{options=value},note(){}};
 vm.createContext(ctx);vm.runInContext(fs.readFileSync('web/app.js','utf8').split('\n').find(line=>line.startsWith('setInterval(')),ctx);
 await tick();assert.equal(options.background,true);
});
