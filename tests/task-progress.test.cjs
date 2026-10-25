'use strict';
const {test}=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
function harness({owner=true,enabled=true,view='Tasks'}={}){
 let html='',result={},save,buttons=[];const calls=[],nodes={};
 const esc=s=>String(s??'').replaceAll('&','&amp;').replaceAll('<','&lt;').replaceAll('"','&quot;');
 const parse=text=>{for(const m of text.matchAll(/id="([^"]+)"/g))if(!nodes['#'+m[1]])nodes['#'+m[1]]={};buttons=[...text.matchAll(/data-task-progress-action="([^"]+)" data-task-id="([^"]+)" data-task-revision="([^"]+)"/g)].map(m=>({dataset:{taskProgressAction:m[1],taskId:m[2],taskRevision:m[3]}}))};
 const panel={set innerHTML(text){html=text;parse(text)},get innerHTML(){return html},set textContent(text){html=String(text)}};nodes['#task-progress-panel']=panel;
 const main={insertAdjacentHTML(_,text){html+=text;parse(text)}};
 const ctx=vm.createContext({state:{currentActorId:owner?'local-owner':'member',featureFlags:enabled?{taskProgress:true}:{},tasks:[{id:'task',roomId:'engineering',revision:4}],rooms:[{id:'engineering'}]},roomId:'engineering',view,render(){},esc,$:s=>s==='#main'?main:nodes[s],document:{querySelectorAll:()=>buttons},form(t,body,fn){html=body;save=fn},api:async(...args)=>{calls.push(args);return result},commit:async(...args)=>calls.push(args),taskForm:t=>calls.push(['edit',t.id]),dispatchTask:async(t,m)=>calls.push(['run',t.id,m]),showTaskAcceptance:async t=>calls.push(['review',t.id]),showTaskExecutionHistory:async t=>calls.push(['history',t.id]),note:m=>calls.push(['notice',m])});
 vm.runInContext(fs.readFileSync('web/task-progress.js','utf8'),ctx);
 return{ctx,nodes,calls,setResult:r=>result=r,html:()=>html,buttons:()=>buttons,submit:d=>save(d)};
}
function fixture(phase='criterion-review-required'){return{tasks:[{taskId:'task',title:'Build <web>',taskRevision:4,taskStatus:'Accepted',definitionRevision:2,phase,detail:'Recorded <detail>',executionId:'attempt',executionMode:'work',executionStatus:'validated',workExecutionId:'attempt',returnedValueCount:1,inputState:'current',dependencies:[{inputName:'source',sourceTaskId:'<source>',sourceTaskRevision:3,executionId:'source-attempt',outputName:'result',current:false}]}],pageCounts:{[phase]:1},totalTasks:27,nextCursor:'task cursor',structuredDispatchEnabled:true,criterionReviewEnabled:true,observations:'controller-recorded-state-only'}}
test('progress labels controller evidence, escapes metadata and reuses authorized action paths',async()=>{
 const h=harness();h.setResult(fixture());await h.ctx.loadTaskProgress();assert(h.html().includes('Controller-recorded state only'));assert(h.html().includes('This page: 1 of 27'));assert(h.html().includes('Criterion review required'));assert(h.html().includes('Saved task state: Accepted'));assert(h.html().includes('Build &lt;web>'));assert(h.html().includes('&lt;source>'));assert(!h.html().includes('valueJson'));
 const review=h.buttons().find(b=>b.dataset.taskProgressAction==='review');await review.onclick();assert(h.calls.some(c=>c[0]==='review'&&c[1]==='task'));h.nodes['#task-progress-next'].onclick();await Promise.resolve();assert(h.calls.some(c=>c[0].includes('after=task%20cursor')));
});
test('members inspect history but get no run, edit or owner-review controls',async()=>{
 const h=harness({owner:false});h.setResult(fixture('retry-required'));await h.ctx.loadTaskProgress();assert.deepEqual(h.buttons().map(b=>b.dataset.taskProgressAction),['history']);await h.buttons()[0].onclick();assert(h.calls.some(c=>c[0]==='history'));assert(!h.calls.some(c=>c[0]==='run'||c[0]==='edit'||c[0]==='review'));
});
test('pending or disabled work cannot advertise a run action and stale revisions refuse actions',async()=>{
 const h=harness(),data=fixture('inputs-unusable');data.tasks[0].dispatchStatus='pending';h.setResult(data);await h.ctx.loadTaskProgress();assert(!h.buttons().some(b=>b.dataset.taskProgressAction==='run'));
 data.tasks[0].dispatchStatus='failed';data.structuredDispatchEnabled=false;await h.ctx.loadTaskProgress();assert(!h.buttons().some(b=>b.dataset.taskProgressAction==='run'));
 data.structuredDispatchEnabled=true;await h.ctx.loadTaskProgress();const run=h.buttons().find(b=>b.dataset.taskProgressAction==='run');h.ctx.state.tasks[0].revision=5;await run.onclick();assert(!h.calls.some(c=>c[0]==='run'));assert(h.calls.some(c=>c[0]==='notice'&&c[1].includes('Task changed')));
});
test('late progress responses and old buttons cannot act after room or generation changes',async()=>{
 const h=harness();let resolve;h.ctx.api=()=>new Promise(r=>resolve=r);const pending=h.ctx.loadTaskProgress();h.ctx.roomId='elsewhere';resolve(fixture());await pending;assert.equal(h.html(),'');
 const next=harness();next.setResult(fixture());await next.ctx.loadTaskProgress();const old=next.buttons()[0];await next.ctx.loadTaskProgress();await old.onclick();assert(next.calls.some(c=>c[0]==='notice'&&c[1].includes('Progress changed')));assert(!next.calls.some(c=>c[0]==='history'));
});
test('progress defaults off, excludes personal rooms and settings remain owner-only',async()=>{
 const off=harness({enabled:false});off.ctx.drawTaskProgress();assert.equal(off.calls.length,0);
 const personal=harness();personal.ctx.state.rooms[0].personalActorId='member';personal.ctx.drawTaskProgress();assert.equal(personal.calls.length,0);
 const settings=harness({view:'Settings',enabled:false});settings.ctx.drawTaskProgress();assert(settings.html().includes('Task progress · Off'));settings.nodes['#task-progress-setting'].onclick();await settings.submit(new Map([['enabled','on']]));assert.equal(settings.calls[0][1].name,'taskProgress');const member=harness({view:'Settings',owner:false});member.ctx.drawTaskProgress();assert.equal(member.html(),'');
 const page=fs.readFileSync('web/index.html','utf8');assert(page.indexOf('src="task-progress.js"')>page.indexOf('src="task-acceptance.js"'));assert(page.indexOf('src="task-progress.js"')<page.indexOf('src="feature-flag-tabs.js"'));
});
