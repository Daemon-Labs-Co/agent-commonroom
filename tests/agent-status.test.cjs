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
