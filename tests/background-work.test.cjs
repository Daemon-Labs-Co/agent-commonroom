const test=require('node:test'),assert=require('node:assert/strict'),vm=require('node:vm'),fs=require('node:fs');
function harness(owner=true){
 const buttons={},calls=[],links=[];
 const details={open:true,removeAttribute(){this.open=false}};
 const panel={querySelector(selector){return links.find(button=>button.attributes?.[selector.slice(1,-1)])||null},closest(){return details},insertBefore(button,before){const index=links.indexOf(before);if(index<0)links.push(button);else links.splice(index,0,button)}};
 const label={dataset:{agentStatus:'p'},closest(){return {querySelector(){return panel}}},after(){throw Error('Action must stay inside agent menu')}};
 const modal={innerHTML:'',classList:{add(){},remove(){}},querySelector(selector){if(selector.startsWith('#')&&!this.innerHTML.includes('id="'+selector.slice(1)+'"'))return null;return buttons[selector]||(buttons[selector]={disabled:false})},showModal(){this.open=true},close(){this.open=false}};
 const content={innerHTML:'',scrollHeight:200,clientHeight:100,scrollTop:0};
 const ctx={view:'Room',state:{currentActorId:owner?'local-owner':'human',featureFlags:{},participants:[{id:'p',name:'Synthetic <agent>',background:[{label:'Started <command>',observedAt:'2026-01-01T12:00:00Z'}]}]},document:{querySelectorAll:()=>[label],createElement:()=>({attributes:{},setAttribute(key,value){this.attributes[key]=value}})},$:s=>s==='#modal'?modal:content,esc:s=>String(s??'').replace(/</g,'&lt;').replace(/>/g,'&gt;'),render(){},openDocumentation(){},api:async(...args)=>calls.push(args),refresh:async()=>{},note:()=>{}};
 vm.createContext(ctx);vm.runInContext(fs.readFileSync('web/background-work.js','utf8'),ctx);return{ctx,modal,buttons,calls,links,content,details};
}
test('background view is explicit, inert and does not launch a model',()=>{
 const h=harness();h.ctx.drawBackgroundWork();h.ctx.drawBackgroundWork();assert.equal(h.links.length,1);assert.equal(h.links[0].attributes['data-background-work'],'p');h.links[0].onclick();assert.equal(h.details.open,false);assert.equal(h.modal.open,true);assert.match(h.modal.innerHTML,/Started &lt;command&gt;/);assert.equal(h.calls.length,0);
});
test('stop all is explicit and available independently of the cosmetic safety button flag',async()=>{
 const h=harness();h.ctx.state.featureFlags.agentSafetyControls=false;h.ctx.openBackgroundWork('p');assert.equal(h.calls.length,0);await h.buttons['#background-stop-all'].onclick();assert.equal(h.calls[0][0],'agent-safety');assert.equal(h.calls[0][1].scope,'application');assert.equal(h.calls[0][1].action,'stop');
 const member=harness(false);member.ctx.openBackgroundWork('p');assert.equal(member.buttons['#background-stop-all'],undefined);
});
test('feature off removes menu actions and an open timeline; live updates preserve an older reading position',()=>{
 const h=harness();h.ctx.openBackgroundWork('p');h.ctx.drawBackgroundWork();assert.equal(h.content.scrollTop,0);assert.match(h.content.innerHTML,/Started &lt;command&gt;/);
 h.ctx.state.featureFlags.agentBackgroundWork=false;h.ctx.drawBackgroundWork();assert.equal(h.modal.open,false);
 const off=harness();off.ctx.state.featureFlags.agentBackgroundWork=false;off.ctx.drawBackgroundWork();off.ctx.openBackgroundWork('p');assert.equal(off.links.length,0);assert.equal(off.modal.open,undefined);
});

test('both optional actions share the same agent menu in stable order without duplicates',()=>{
 const h=harness();vm.runInContext(fs.readFileSync('web/agent-recovery.js','utf8'),h.ctx);h.ctx.drawSessionRecovery();h.ctx.drawBackgroundWork();h.ctx.drawSessionRecovery();h.ctx.drawBackgroundWork();assert.deepEqual(h.links.map(button=>button.textContent),['Background work','Recover session']);
 const native=harness();native.ctx.state.featureFlags.squadMenuLayout=false;native.ctx.drawBackgroundWork();native.links[0].onclick();assert.equal(native.details.open,false);assert.equal(native.modal.open,true);
});
