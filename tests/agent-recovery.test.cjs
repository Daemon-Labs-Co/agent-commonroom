const test=require('node:test'),assert=require('node:assert/strict'),vm=require('node:vm'),fs=require('node:fs');
function harness(){
 const buttons={},calls=[],links=[],input={value:'Unsent draft',focus(){this.focused=true}},select={value:'room'};
 const modal={classList:{add(){},remove(){}},innerHTML:'',querySelector(s){return buttons[s]||(buttons[s]={disabled:s.includes('resume')||s.includes('fresh'),textContent:''})},showModal(){this.open=true},close(){this.open=false}};
 const ctx={state:{currentActorId:'local-owner',featureFlags:{},participants:[{id:'p',name:'Synthetic <agent>',threadId:'saved'}]},view:'Room',recipient:'room',$:s=>s==='#modal'?modal:s.includes('recipient')?select:input,esc:s=>String(s).replace(/</g,'&lt;'),document:{querySelectorAll:()=>[{dataset:{agentStatus:'p'},after:b=>links.push(b)}],createElement:()=>({setAttribute(){}})},api:async(path,body)=>{calls.push([path,body]);return body.action==='check'?{idle:true,detail:'Worker idle'}:{idle:true,recovered:true,detail:'Ready'}},refresh:async()=>{input.value=''},render(){},note(){},openDocumentation(){}};
 vm.createContext(ctx);vm.runInContext(fs.readFileSync('web/agent-recovery.js','utf8'),ctx);return{ctx,buttons,calls,modal,input,select,links};
}
test('recovery requires explicit worker check and never replays a task',async()=>{
 const h=harness();h.ctx.openSessionRecovery('p');assert.equal(h.calls.length,0);assert.equal(h.buttons['#recovery-resume'].disabled,true);
 await h.buttons['#recovery-check'].onclick();assert.equal(h.buttons['#recovery-resume'].disabled,false);await h.buttons['#recovery-resume'].onclick();
 assert.equal(h.calls.length,2);assert.equal(h.calls[1][1].action,'resume');assert.equal(h.ctx.recipient,'p');assert.equal(h.input.value,'Unsent draft');assert.equal(h.input.focused,true);assert.equal(h.modal.open,false);
});
test('busy, unknown, absent sessions and feature/permission restrictions fail closed',async()=>{
 const h=harness();delete h.ctx.state.participants[0].threadId;h.ctx.openSessionRecovery('p');await h.buttons['#recovery-check'].onclick();assert.equal(h.buttons['#recovery-resume'].disabled,true);assert.equal(h.buttons['#recovery-fresh'].disabled,false);
 h.ctx.api=async()=>({idle:false,detail:'Worker busy'});await h.buttons['#recovery-check'].onclick();assert.equal(h.buttons['#recovery-fresh'].disabled,true);
 h.ctx.api=async()=>{throw Error('Unknown worker')};await h.buttons['#recovery-check'].onclick();assert.match(h.buttons['#recovery-status'].textContent,/Unknown worker/);assert.equal(h.buttons['#recovery-fresh'].disabled,true);
 const off=harness();off.ctx.state.featureFlags.agentRecovery=false;off.ctx.drawSessionRecovery();off.ctx.openSessionRecovery('p');assert.equal(off.links.length,0);assert.equal(off.modal.open,undefined);
 const member=harness();member.ctx.state.currentActorId='human';member.ctx.drawSessionRecovery();assert.equal(member.links.length,0);
});
