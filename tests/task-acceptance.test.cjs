'use strict';
const {test}=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
function harness({owner=true,enabled=true,view='Tasks',realRequest=false,retryToken=false}={}){
 let html='',title='',save,result={};const nodes={},calls=[];
 const esc=s=>String(s??'').replaceAll('&','&amp;').replaceAll('<','&lt;').replaceAll('"','&quot;');
 const insert=(_,s)=>{html+=s;for(const m of s.matchAll(/id="([^"]+)"/g))nodes['#'+m[1]]={};};
 const ctx=vm.createContext({state:{currentActorId:owner?'local-owner':'member',featureFlags:enabled?{taskDefinitions:true,taskAcceptanceReview:true}:{},tasks:[]},roomId:'engineering',view,render(){},taskCard(){return '<article>Task</article>'},wireTasks(){},tdSummary:d=>'<p>'+esc(d.task)+'</p>',esc,$:s=>s==='#main'?{insertAdjacentHTML:insert}:nodes[s],document:{querySelectorAll:()=>[]},selectField:(label,name,options,value)=>'<select name="'+name+'">'+options.map(o=>'<option '+(o===value?'selected':'')+'>'+esc(o)+'</option>').join('')+'</select>',form(t,body,fn){title=t;html=body;save=fn;insert('',body)},api:async(...args)=>{calls.push(args);return result},commit:async(...args)=>calls.push(args),refresh:async()=>{},note:m=>calls.push(['notice',m])});
 if(realRequest){
  let failed=false;
  ctx.crypto={randomUUID:()=> 'synthetic-review-operation'};
  ctx.fetch=async(url,options)=>{
   calls.push([url,options]);
   if(url==='/api/session')return{ok:true,json:async()=>({csrf:'renewed-token'})};
   if(options?.method==='POST'&&retryToken&&!failed){failed=true;return{ok:false,status:403,text:async()=> 'Invalid request token'}};
   return{ok:true,json:async()=>result};
  };
  const app=fs.readFileSync('web/app.js','utf8');
  vm.runInContext("let csrf='initial-token';\n"+app.slice(app.indexOf('async function api('),app.indexOf('function note(')),ctx);
 }
 vm.runInContext(fs.readFileSync('web/task-acceptance.js','utf8'),ctx);
 return{ctx,nodes,calls,html:()=>html,title:()=>title,setResult:r=>result=r,submit:d=>save(d)};
}
function fixture(){return{enabled:true,ready:true,status:'Needs review',snapshot:'snapshot',review:{revision:2},task:{id:'task',title:'Build <web>',revision:7,status:'Submitted',definition:{task:'Return <typed output>',revision:2,acceptanceCriteria:[{id:'matches',description:'Matches <source>',method:'human',output:'result'},{id:'shape',description:'Matches finite type',method:'schema',output:'result'}]}},schemaAssessments:[{criterionId:'shape',decision:'passed',note:'Finite structural check only'}],workExecution:{id:'work',inputs:[{name:'source',sourceVersion:'manual-v1',digest:'input-hash',valueJson:'"<input>"'}],outputs:[{name:'result',digest:'hash',valueJson:'9223372036854775807'}]},reviews:[],nextCursor:''}}
function decisions(){return new Map([['criterion-0','passed'],['criterion-note-0','Compared exact input'],['review-note','Owner rationale'],['action','accept']])}
test('criterion review displays exact inputs/outputs and submits only owner human decisions',async()=>{
 const h=harness();h.setResult(fixture());await h.ctx.showTaskAcceptance({id:'task'});assert.equal(h.title(),'Review Build &lt;web>');assert(h.html().includes('9223372036854775807'));assert(h.html().includes('&lt;input>'));assert(h.html().includes('name="criterion-0"'));assert(!h.html().includes('name="criterion-1"'));assert(h.html().includes('Finite structural check only'));assert(h.html().includes('disabled selected>Choose a decision'));
 await h.submit(decisions());const request=h.calls[1][1];assert.equal(request.taskRevision,7);assert.equal(request.reviewRevision,2);assert.equal(request.snapshot,'snapshot');assert.equal(request.decisions.length,1);assert.equal(request.decisions[0].criterionId,'matches');assert.equal(request.decisions[0].note,'Compared exact input');assert.equal(request.outputs,undefined);
});
test('members, disabled and unready reviews preserve history without decision controls',async()=>{
 for(const variant of ['member','off','unready']){const h=harness({owner:variant!=='member'}),data=fixture();if(variant==='off')data.enabled=false;if(variant==='unready')data.ready=false;data.reviews=[{revision:1,action:'changes',reviewedBy:'owner',reviewedAt:'synthetic',note:'<old rationale>',executionId:'older',snapshot:'old',assessments:[{criterionId:'matches',method:'human',decision:'failed',note:'<review note>',outputName:'result'}]}];data.reviewedExecutions={older:{definition:{task:'Old scope'},outputs:[{name:'result',digest:'old-hash',valueJson:'"<old value>"'}]}};h.setResult(data);await h.ctx.showTaskAcceptance({id:'task'});assert(!h.html().includes('name="criterion-0"'));assert(h.html().includes('&lt;old rationale>'));assert(h.html().includes('&lt;old value>'));await h.submit(decisions());assert.equal(h.calls.length,1)}
});
test('stale room or view responses cannot open or submit a task review',async()=>{
 const h=harness();let resolve;h.ctx.api=()=>new Promise(r=>resolve=r);const pending=h.ctx.showTaskAcceptance({id:'task'});h.ctx.view='Settings';resolve(fixture());await pending;assert.equal(h.html(),'');
 const next=harness();next.setResult(fixture());await next.ctx.showTaskAcceptance({id:'task'});next.ctx.roomId='elsewhere';await assert.rejects(()=>next.submit(decisions()),/Room or review changed/);assert.equal(next.calls.length,1);
});
test('history paginates without reopening current decision controls and settings are owner-only',async()=>{
 const h=harness(),data=fixture();data.nextCursor='3';h.setResult(data);await h.ctx.showTaskAcceptance({id:'task'},'13');assert(!h.html().includes('name="criterion-0"'));assert(h.nodes['#task-review-older']);assert(h.nodes['#task-review-latest']);
 const settings=harness({view:'Settings',enabled:false});settings.ctx.drawTaskAcceptanceSettings();assert(settings.html().includes('Task criterion review · Off'));settings.nodes['#task-acceptance-setting'].onclick();await settings.submit(new Map([['enabled','on']]));assert.equal(settings.calls[0][1].name,'taskAcceptanceReview');const member=harness({view:'Settings',owner:false});member.ctx.drawTaskAcceptanceSettings();assert.equal(member.html(),'');
 const page=fs.readFileSync('web/index.html','utf8');assert(page.indexOf('src="task-acceptance.js"')>page.indexOf('src="task-definitions.js"'));assert(page.indexOf('src="task-acceptance.js"')<page.indexOf('src="feature-flag-tabs.js"'));
});

test('criterion decision form uses real request helper and retains operation key across token refresh',async()=>{
 for(const action of ['accept','changes'])for(const retryToken of [false,true]){
  const h=harness({realRequest:true,retryToken});h.setResult(fixture());
  await h.ctx.showTaskAcceptance({id:'task'});
  const data=decisions();data.set('action',action);await h.submit(data);
  const reads=h.calls.filter(([url,options])=>url.startsWith('/api/task-acceptance?')&&options?.method!=='POST');
  assert.equal(reads.length,1);
  const writes=h.calls.filter(([,options])=>options?.method==='POST');
  assert.equal(writes.length,retryToken?2:1);
  for(const [url,options] of writes){
   assert.equal(url,'/api/task-acceptance?room=engineering');
   assert.equal(options.headers['Idempotency-Key'],'synthetic-review-operation');
   assert.equal(options.headers['Content-Type'],'application/json');
   const body=JSON.parse(options.body);assert.equal(body.action,action);assert.equal(body.taskRevision,7);assert.equal(body.snapshot,'snapshot');assert.equal(body.decisions[0].criterionId,'matches');
  }
  assert.equal(writes[0][1].headers['X-CSRF-Token'],'initial-token');
  if(retryToken){assert.equal(writes[1][1].headers['X-CSRF-Token'],'renewed-token');assert.equal(writes[1][1].body,writes[0][1].body)}
 }
});
