'use strict';
const {test}=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
function harness(){const controls={},calls=[];let html='',submit,body;
 const target={querySelector(){return null},insertAdjacentHTML(_,value){html=value;controls['#configure-idle-policy']={};controls['#toggle-idle-policy']={}}};
 const ctx=vm.createContext({view:'Settings',state:{currentActorId:'local-owner',featureFlags:{}},render(){},$(id){return id==='#main'?target:controls[id]},esc(v){return String(v).replaceAll('<','&lt;').replaceAll('"','&quot;')},form(_,value,fn){body=value;submit=fn},commit:async(...v)=>calls.push(v),refresh:async()=>{}});
 vm.runInContext(fs.readFileSync('web/idle-policy.js','utf8'),ctx);
 return {ctx,controls,calls,html:()=>html,body:()=>body,submit:d=>submit(d)};
}
test('owner sees default policy; other humans and room views have no controls',()=>{
 const h=harness();h.ctx.drawIdlePolicySettings();assert.match(h.html(),/5 idle minutes/);assert.match(h.html(),/Idle shutdown · On/);assert.match(h.html(),/Production development dispatch remains gated/);
 for(const mode of ['other','room']){const h=harness();if(mode==='other')h.ctx.state.currentActorId='developer';else h.ctx.view='Room';h.ctx.drawIdlePolicySettings();assert.equal(h.html(),'')}
});
test('interval saves exact current revision and numeric value without changing the flag',async()=>{
 const h=harness();h.ctx.state.agentIdleSettings={minutes:8,revision:4};h.ctx.drawIdlePolicySettings();h.controls['#configure-idle-policy'].onclick();assert.equal(h.calls.length,0);assert.match(h.body(),/value="8"/);
 await h.submit(new Map([['minutes','12']]));assert.equal(h.calls[0][0],'agents.idle_policy.updated');assert.equal(h.calls[0][1].minutes,12);assert.equal(h.calls[0][1].revision,4);assert.equal(h.calls.length,1);
});
test('off policy keeps configuration and re-enable controls available',async()=>{
 const h=harness();h.ctx.state.featureFlags.idleAgentShutdown=false;h.ctx.drawIdlePolicySettings();assert.match(h.html(),/Idle shutdown · Off/);assert(h.controls['#configure-idle-policy']);h.controls['#toggle-idle-policy'].onclick();assert.match(h.body(),/Hard deadlines and ordinary shutdown/);await h.submit(new Map([['enabled','on']]));assert.equal(h.calls[0][0],'feature.flag.updated');assert.equal(h.calls[0][1].name,'idleAgentShutdown');assert.equal(h.calls[0][1].enabled,true);
});
