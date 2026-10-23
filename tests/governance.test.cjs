'use strict';
const {test}=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
function harness(owner=true,enabled=false,view='Settings'){
 const controls={},calls=[];let html='',submit;
 const target={insertAdjacentHTML(_,value){html+=value;for(const m of value.matchAll(/id="([^"]+)"/g))controls['#'+m[1]]={}}};
 const ctx=vm.createContext({view,roomId:'engineering',state:{currentActorId:owner?'local-owner':'agent',featureFlags:{agentCoordination:enabled},connections:[{id:'c',protocol:'CLI',endpoint:'codex'}],participants:[{id:'a',name:'<agent>',roomId:'engineering',status:'Idle',connectionId:'c'}]},render(){},$(id){return id==='#main'?target:controls[id]},esc(v){return String(v).replaceAll('<','&lt;').replaceAll('"','&quot;')},form(_,body,fn){html+=body;submit=fn},commit:async(...args)=>calls.push(args),api:async(...args)=>{calls.push(args);return {available:true,appointment:{participantId:'a',revision:7}}},refresh:async()=>{}});
 vm.runInContext(fs.readFileSync('web/governance.js','utf8'),ctx);
 return {ctx,controls,calls,html:()=>html,submit:d=>submit(d)};
}
test('off owner control remains accessible and grants no permissions',async()=>{
 const h=harness();h.ctx.drawGovernanceControls();assert.match(h.html(),/Agent coordination · Off/);h.controls['#toggle-coordination'].onclick();await h.submit(new Map([['enabled','on']]));assert.equal(h.calls[0][0],'feature.flag.updated');assert.equal(h.calls[0][1].name,'agentCoordination');assert.equal(h.calls[0][1].enabled,true);
});
test('non-owner cannot see controls; off room has no selection control',()=>{
 for(const h of [harness(false,true),harness(true,false,'Room')]){h.ctx.drawGovernanceControls();assert.equal(h.html(),'')}
});
test('replacement confirms current server revision and escapes display names',async()=>{
 const h=harness(true,true,'Room');h.ctx.drawGovernanceControls();await h.controls['#choose-governor'].onclick();assert.match(h.html(),/&lt;agent>/);assert.match(h.html(),/Confirm replacement/);await h.submit(new Map([['participantId','a']]));assert.equal(h.calls[1][0],'governance?room=engineering');assert.equal(h.calls[1][1].revision,7);assert.equal(h.calls[1][1].participantId,'a');
});

test('unsupported connections are excluded from the candidate list',async()=>{const h=harness(true,true,'Room');h.ctx.state.connections[0].endpoint='unsupported';h.ctx.drawGovernanceControls();await h.controls['#choose-governor'].onclick();assert.doesNotMatch(h.html(),/<option/)});

test('coordinator request binds displayed identity and revision',async()=>{
 const h=harness(true,true,'Room');h.ctx.drawGovernanceControls();await h.controls['#coordinate-turn'].onclick();assert.match(h.html(),/Revision: 7/);assert.match(h.html(),/Replacement or disabling/);await h.submit(new Map([['text','Synthetic request']]));assert.equal(h.calls[1][0],'agents?room=engineering');assert.equal(h.calls[1][1].participantId,'a');assert.equal(h.calls[1][1].coordinationRevision,7);assert.equal(h.calls[1][1].text,'Synthetic request');
});
test('room change while reading appointment prevents a stale request form',async()=>{
 const h=harness(true,true,'Room');h.ctx.drawGovernanceControls();h.ctx.api=async()=>{h.ctx.roomId='other';return {available:true,appointment:{participantId:'a',revision:7}}};await h.controls['#coordinate-turn'].onclick();assert.doesNotMatch(h.html(),/<textarea/);
});
