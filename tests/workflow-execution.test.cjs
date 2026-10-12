const test=require('node:test');
const assert=require('node:assert/strict');
const vm=require('node:vm');
const fs=require('node:fs');

function harness(){
 const calls=[],forms=[],notes=[];let refreshes=0,poll;
 const context={crypto:require('node:crypto'),structuredClone,render:()=>{},window:{addEventListener:()=>{}},setInterval:fn=>{poll=fn},
  state:{currentActorId:'local-owner',featureFlags:{workflowExecution:true},participants:[],workflowRuns:[]},view:'Workflows',
  form:(title,fields,save)=>forms.push({title,fields,save}),note:text=>notes.push(text),
  api:async(path,body)=>{calls.push({path,body});return {}},refresh:async()=>{refreshes++},
  document:{querySelector:()=>null},$:()=>null};
 vm.createContext(context);vm.runInContext(fs.readFileSync('web/workflows.js','utf8'),context);
 vm.runInContext('workflow={id:"saved",name:"Feature",revision:7};wfDirty=false',context);
 return {context,calls,forms,notes,poll:()=>poll(),refreshes:()=>refreshes};
}
test('Run requires a saved clean graph, owner and enabled flag before creating a task prompt',async()=>{
 for(const change of ['wfDirty=true','workflow.id=""','state.currentActorId="developer"','state.featureFlags.workflowExecution=false']){
  const h=harness();vm.runInContext(change,h.context);h.context.wfStartRun();assert.equal(h.forms.length,0);assert.equal(h.calls.length,0);assert.equal(h.notes.length,1);
 }
 const h=harness();h.context.wfStartRun();assert.match(h.forms[0].fields,/isolated containers/);
 await h.forms[0].save(new Map([['input','Owner-approved feature']]));
 assert.equal(h.calls[0].path,'workflow-runs');assert.equal(h.calls[0].body.revision,7);assert.equal(h.calls[0].body.input,'Owner-approved feature');assert.equal(h.refreshes(),1);
});
test('human acceptance, yes/no and rework carry pinned revision and rationale',async()=>{
 for(const [port,action] of [['next','advance'],['yes','advance'],['no','advance'],['changes','reject']]){
  const h=harness();h.context.wfGate({id:'run',revision:9},port);await h.forms[0].save(new Map([['evidence','Inspected actual diff and checks']]));
  assert.equal(h.calls[0].body.action,action);assert.equal(h.calls[0].body.runId,'run');assert.equal(h.calls[0].body.revision,9);assert.equal(h.calls[0].body.port,port);assert.match(h.calls[0].body.evidence,/actual diff/);
 }
});
test('failed controls show server error without claiming success',async()=>{
 const h=harness();h.context.api=async()=>{throw Error('Run changed; reload')};await h.context.wfRunAction({id:'run',revision:1},'cancel');assert.deepEqual(h.notes,['Run changed; reload']);assert.equal(h.refreshes(),0);
});
test('polling reconciles running runs and leaves open forms and dirty drafts alone',async()=>{
 const h=harness();h.context.state.workflowRuns=[{status:'running'}];await h.poll();assert.equal(h.calls[0].path,'workflow-runs');assert.equal(h.refreshes(),1);
 vm.runInContext('wfDirty=true',h.context);await h.poll();assert.equal(h.calls.length,1);
 vm.runInContext('wfDirty=false',h.context);h.context.document.querySelector=()=>({open:true});await h.poll();assert.equal(h.calls.length,1);
 h.context.document.querySelector=()=>null;h.context.view='Room';await h.poll();assert.equal(h.calls.length,1);
});
