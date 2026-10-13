const test=require('node:test'),assert=require('node:assert/strict'),vm=require('node:vm'),fs=require('node:fs');
function harness(){
 const attrs={},body=[],events={};
 const label={dataset:{agentStatus:'a'},classList:{add(){}},setAttribute(k,v){attrs[k]=v},removeAttribute(k){delete attrs[k]},getBoundingClientRect(){return {left:100,top:80,bottom:100}}};
 const document={querySelectorAll:()=>[label],body:{append(n){body.push(n)}},createElement(){return {style:{},setAttribute(){},textContent:'',getBoundingClientRect(){return {width:300,height:100}},remove(){this.removed=true}}}};
 const ctx={document,view:'Room',render(){},state:{featureFlags:{},participants:[{id:'a',status:'Blocked',statusDetails:{activity:'Workflow <script> / Implement',reason:'Time limit exceeded',next:'Inspect partial work'}}]},window:{innerWidth:400,innerHeight:200,addEventListener(k,fn){events[k]=fn}}};
 vm.createContext(ctx);vm.runInContext(fs.readFileSync('web/agent-status.js','utf8'),ctx);return {ctx,label,attrs,body,events};
}
test('hover and keyboard focus show bounded plain-text diagnostics without HTML injection',()=>{
 const h=harness();h.ctx.drawAgentStatusDetails();assert.equal(h.label.tabIndex,0);assert.match(h.attrs['aria-label'],/Time limit exceeded/);
 h.label.onmouseenter();assert.match(h.body[0].textContent,/Doing: Workflow <script>/);assert.equal(h.body[0].innerHTML,undefined);assert.equal(h.body[0].style.left,'92px');
 h.label.onmouseleave();assert.equal(h.body[0].removed,true);assert.equal(h.attrs['aria-describedby'],undefined);
 h.label.onfocus();assert.equal(h.attrs['aria-describedby'],'agent-status-tooltip');h.events.keydown({key:'Escape'});assert.equal(h.body[1].removed,true);
});
test('disabled flag and unrelated views leave status alone; rerenders remove stale popups',()=>{
 const h=harness();h.ctx.state.featureFlags.agentStatusDetails=false;h.ctx.drawAgentStatusDetails();assert.equal(h.label.onmouseenter,undefined);
 h.ctx.state.featureFlags.agentStatusDetails=true;h.ctx.drawAgentStatusDetails();h.label.onfocus();h.ctx.view='Settings';h.ctx.drawAgentStatusDetails();assert.equal(h.body[0].removed,true);
});
test('missing historical details do not invent a client-side cause',()=>{
 const h=harness();delete h.ctx.state.participants[0].statusDetails;h.ctx.drawAgentStatusDetails();assert.equal(h.label.onfocus,undefined);assert.deepEqual(h.attrs,{});
});

test('long-running turns show a warning without claiming useful progress',()=>{
 const h=harness();const p=h.ctx.state.participants[0];p.status='Working';p.statusDetails.longRunning=true;p.statusDetails.reason='Process liveness and progress are not confirmed';
 h.ctx.drawAgentStatusDetails();assert.equal(h.label.textContent,'Working · taking longer');h.label.onfocus();assert.match(h.body[0].textContent,/not confirmed/);
});

function recoveryHarness(owner=true){
 const h=harness(),buttons={},calls=[];
 const modal={innerHTML:'',classList:{add(){},remove(){}},closed:false,querySelector(selector){
  if(selector.startsWith('#')&&!this.innerHTML.includes('id="'+selector.slice(1)+'"'))return null;
  return buttons[selector]||(buttons[selector]={textContent:'',disabled:false});
 },showModal(){this.open=true},close(){this.closed=true}};
 const input={value:'Existing unsent draft',focus(){this.focused=true}},select={value:'room'};
 Object.assign(h.ctx,{$:s=>s==='#modal'?modal:s.includes('recipient')?select:input,esc:s=>String(s??'').replace(/[<>&"]/g,c=>({'<':'&lt;','>':'&gt;','&':'&amp;','"':'&quot;'}[c])),api:async(...args)=>{calls.push(args);return {credentialPresent:true}},openDocumentation:()=>{},testConnection:c=>calls.push(['test',c.id]),note:()=>{},recipient:'room'});
 h.ctx.state.currentActorId=owner?'local-owner':'human';h.ctx.state.connections=[{id:'c',name:'Synthetic connection'}];h.ctx.state.participants[0].connectionId='c';
 return {...h,buttons,calls,modal,input,select};
}
test('diagnosis opens safely without a provider call or retry and preserves draft',()=>{
 const h=recoveryHarness();h.ctx.drawAgentStatusDetails();h.label.onclick();
 assert.equal(h.modal.open,true);assert.equal(h.calls.length,0);assert.match(h.modal.innerHTML,/Workflow &lt;script&gt;/);
 h.buttons['#agent-compose-request'].onclick();assert.equal(h.ctx.recipient,'a');assert.equal(h.select.value,'a');assert.equal(h.input.value,'Existing unsent draft');assert.equal(h.input.focused,true);assert.equal(h.calls.length,0);
});
test('connection status is distinguished from provider validity; test needs explicit click',async()=>{
 const h=recoveryHarness();h.ctx.openAgentDiagnosis('a');await h.buttons['#agent-check-auth'].onclick();
 assert.equal(h.calls[0][0],'provider-auth');assert.match(h.buttons['#agent-recovery-result'].textContent,/does not prove they are valid/);
 assert.equal(h.calls.length,1);h.buttons['#agent-test-connection'].onclick();assert.deepEqual(h.calls[1],['test','c']);
 h.buttons['#agent-view-history'].onclick();assert.equal(h.ctx.view,'History & retention');
});
test('recovery respects flags, owner permissions and emergency stop',()=>{
 const h=recoveryHarness(false);h.ctx.state.agentsStopped=true;h.ctx.openAgentDiagnosis('a');
 assert.equal(h.buttons['#agent-check-auth'],undefined);assert.equal(h.buttons['#agent-test-connection'],undefined);assert.equal(h.buttons['#agent-view-history'],undefined);assert.equal(h.buttons['#agent-compose-request'].disabled,true);
 h.buttons['#agent-compose-request'].onclick();assert.equal(h.ctx.recipient,'room');
 const off=recoveryHarness();off.ctx.state.featureFlags.agentStatusDetails=false;off.ctx.openAgentDiagnosis('a');assert.equal(off.modal.open,undefined);
 const flags=recoveryHarness();flags.ctx.state.featureFlags.connectionTesting=false;flags.ctx.state.featureFlags.connectionSignIn=false;flags.ctx.openAgentDiagnosis('a');assert.equal(flags.buttons['#agent-check-auth'],undefined);assert.equal(flags.buttons['#agent-test-connection'],undefined);
});
