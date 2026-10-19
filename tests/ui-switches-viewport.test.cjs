'use strict';
const {test}=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
function node(){return {children:[],dataset:{},attributes:{},disabled:false,classList:{contains(){return false}},setAttribute(k,v){this.attributes[k]=v},removeAttribute(k){delete this.attributes[k]},append(...children){this.children.push(...children)},replaceWith(row){this.replacement=row}}}
function switches(){const calls=[],notes=[],storage=[];let reloads=0,resolve,reject;
 const context=vm.createContext({Set,view:'Settings',state:{currentActorId:'local-owner',featureFlags:{}},render(){},document:{createElement:node,querySelectorAll(){return []}},api(path,body){calls.push({path,body});return new Promise((ok,bad)=>{resolve=ok;reject=bad})},note(value){notes.push(value)},sessionStorage:{setItem(...v){storage.push(v)}},window:{location:{reload(){reloads++}}},activeFeatureCategory:'Cosmetic'});
 vm.runInContext(fs.readFileSync('web/feature-switches.js','utf8'),context);return {context,calls,notes,storage,reloads:()=>reloads,resolve:()=>resolve(),reject:()=>reject(new Error('Rejected'))};}
test('switch defaults preserve existing flags, with newest messages explicitly off',()=>{
 const f=switches();assert.equal(f.context.featureSwitchEnabled('roomViewportFit',{}),true);assert.equal(f.context.featureSwitchEnabled('newestMessagesFirst',{}),false);assert.equal(f.context.featureSwitchEnabled('newestMessagesFirst',{newestMessagesFirst:true}),true);assert.equal(f.context.featureSwitchEnabled('agentSafetyControls',{agentSafetyControls:false}),false);
});
test('accessible switch saves once through the authorized event API and preserves feature category',async()=>{
 const f=switches(),original=node(),control=f.context.makeFeatureSwitch(original,'newestMessagesFirst','Newest room messages');
 assert.equal(control.attributes.role,'switch');assert.equal(control.attributes['aria-checked'],'false');assert.equal(control.attributes['aria-label'],'Newest room messages');
 const pending=control.onclick();assert.equal(control.disabled,true);await control.onclick();assert.equal(f.calls.length,1);assert.equal(f.calls[0].path,'events');assert.equal(f.calls[0].body.kind,'feature.flag.updated');assert.equal(f.calls[0].body.payload.enabled,true);f.resolve();await pending;
 assert.equal(f.reloads(),1);assert.deepEqual(f.storage[0],['commonroom-feature-panel','Cosmetic']);assert.equal(control.disabled,true,'successful save stays protected until reload');
});
test('failed save restores the switch and reports the error without reloading',async()=>{
 const f=switches(),control=f.context.makeFeatureSwitch(node(),'roomViewportFit','Fit room');const pending=control.onclick();assert.equal(control.attributes['aria-checked'],'false');f.reject();await pending;assert.equal(control.attributes['aria-checked'],'true');assert.equal(control.disabled,false);assert.deepEqual(f.notes,['Rejected']);assert.equal(f.reloads(),0);
});
test('non-owners cannot save and disabled presentation leaves existing controls alone',async()=>{
 const f=switches(),control=f.context.makeFeatureSwitch(node(),'roomViewportFit','Fit room');f.context.state.currentActorId='developer';await control.onclick();assert.equal(f.calls.length,0);
 f.context.state.currentActorId='local-owner';f.context.state.featureFlags.featureFlagSwitches=false;f.context.document.querySelectorAll=()=>{throw new Error('Disabled module touched controls')};f.context.drawFeatureSwitches();
});
test('room viewport starts at top, keeps background position, and leaves other views untouched',()=>{
 const classes=new Set(),moves=[];const context=vm.createContext({view:'Room',state:{featureFlags:{}},render(){},document:{querySelector(){return {}},body:{classList:{contains(name){return classes.has(name)},toggle(name,on){if(on)classes.add(name);else classes.delete(name)}}}},window:{scrollTo(value){moves.push(value)}}});
 vm.runInContext(fs.readFileSync('web/room-viewport.js','utf8'),context);context.fitRoomViewport();assert(classes.has('room-viewport-fit'));assert.equal(moves[0].top,0);context.fitRoomViewport();assert.equal(moves.length,1);
 context.view='Settings';context.fitRoomViewport();assert(!classes.has('room-viewport-fit'));context.view='Room';context.state.featureFlags.roomViewportFit=false;context.fitRoomViewport();assert(!classes.has('room-viewport-fit'));
 context.state.featureFlags.roomViewportFit=true;context.document.querySelector=()=>null;context.fitRoomViewport();assert(!classes.has('room-viewport-fit'),'workflow/task room tabs keep normal layouts');
});
test('workspace disposal switch stays off until explicitly enabled',()=>{
 const f=switches();assert.equal(f.context.featureSwitchEnabled('workspaceDisposal',{}),false);assert.equal(f.context.featureSwitchEnabled('workspaceDisposal',{workspaceDisposal:false}),false);assert.equal(f.context.featureSwitchEnabled('workspaceDisposal',{workspaceDisposal:true}),true);
 const control=f.context.makeFeatureSwitch(node(),'workspaceDisposal','Workspace disposal');assert.equal(control.attributes['aria-checked'],'false');
});
