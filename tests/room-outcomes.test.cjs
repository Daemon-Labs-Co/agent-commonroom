'use strict';
const {test}=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
function harness({enabled=true,owner=true,view='Room',personal=false}={}){
 let html='',save;const calls=[],button={};
 const room={id:'work',personalActorId:personal?'human':'',outcome:'<untrusted outcome>',acceptanceCriteria:'Verify & review'};
 const ctx=vm.createContext({state:{currentActorId:owner?'local-owner':'member',featureFlags:{roomOutcomes:enabled},rooms:[room]},view,roomId:'work',render(){},$:s=>s==='#main'?{insertAdjacentHTML(_,value){html+=value}}:button,esc:s=>String(s).replaceAll('&','&amp;').replaceAll('<','&lt;'),form(_,body,fn){html+=body;save=fn},commit:async(...args)=>calls.push(args)});
 vm.runInContext(fs.readFileSync('web/room-outcomes.js','utf8'),ctx);
 return{ctx,room,button,calls,html:()=>html,save:()=>save(new Map([['enabled','on']]))};
}
test('room outcome fields are escaped, bounded and unavailable for personal rooms or disabled flag',()=>{
 const h=harness();const fields=h.ctx.roomOutcomeFields(h.room);assert.match(fields,/&lt;untrusted outcome>/);assert.match(fields,/maxlength="2000" required/);assert.match(fields,/maxlength="8000" required/);
 assert.equal(h.ctx.roomOutcomeFields({personalActorId:'human'}),'');assert.equal(harness({enabled:false}).ctx.roomOutcomeFields(h.room),'');
});
test('room and task views show saved intent, including read-only history while disabled',()=>{
 for(const view of ['Room','Tasks']){const h=harness({view,enabled:false,owner:false});h.ctx.drawRoomOutcomes();assert.match(h.html(),/Room deliverable/);assert.match(h.html(),/&lt;untrusted outcome>/);assert.match(h.html(),/Completion is assessed by a human/)}
 const p=harness({personal:true});p.ctx.drawRoomOutcomes();assert.equal(p.html(),'');
});
test('only owner sees the default-off setting and it uses authenticated events',async()=>{
 const h=harness({view:'Settings',enabled:false});h.ctx.drawRoomOutcomes();assert.match(h.html(),/Room outcomes · Off/);h.button.onclick();await h.save();assert.equal(h.calls[0][0],'feature.flag.updated'); // commit takes kind, payload
 assert.equal(h.calls[0][1].name,'roomOutcomes');assert.equal(h.calls[0][1].enabled,true);
 const member=harness({view:'Settings',owner:false});member.ctx.drawRoomOutcomes();assert.equal(member.html(),'');
 const page=fs.readFileSync('web/index.html','utf8');assert(page.indexOf('src="room-outcomes.js"')<page.indexOf('src="feature-flag-tabs.js"'));assert(page.indexOf('src="room-outcomes.js"')<page.indexOf('src="settings-tabs.js"'));
});

test('legacy room settings allow blank definitions while new and defined rooms require both fields',()=>{
 const h=harness();
 const legacy={id:'engineering',name:'Renamed legacy room',description:'Unrelated description'};
 const fields=h.ctx.roomOutcomeFields(legacy);
 const controls=[...fields.matchAll(/<textarea\b([^>]*)>(.*?)<\/textarea>/g)];
 assert.equal(controls.length,2);
 const payload={...legacy};
 for(const [,attributes,value] of controls){
  const name=attributes.match(/name="([^"]+)"/)[1];
  // These are the constraints the browser checks before the existing form's
  // submit handler can send Object.fromEntries(FormData).
  assert(!/\brequired\b/.test(attributes)&&value==='');payload[name]=value;
 }
 assert.equal(payload.outcome,'');assert.equal(payload.acceptanceCriteria,'');
 assert.equal(payload.description,'Unrelated description');
 for(const room of [{},{id:'work',outcome:'Defined'},{id:'work',acceptanceCriteria:'Verify'},h.room]){
  const requiredFields=[...h.ctx.roomOutcomeFields(room).matchAll(/<textarea\b([^>]*)>/g)];
  assert.equal(requiredFields.length,2);for(const [,attributes]of requiredFields)assert.match(attributes,/\brequired\b/);
 }
});
