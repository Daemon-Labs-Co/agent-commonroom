'use strict';
const {test}=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
function fixture(enabled){
 const messages=Object.freeze([Object.freeze({id:'old',author:'Human'}),Object.freeze({id:'new',author:'Agent'})]);
 let list;
 const context=vm.createContext({view:'Room',roomId:'engineering',state:{messages,featureFlags:enabled===undefined?{}:{newestMessagesFirst:enabled}},document:{querySelector(){return list}},render(){
  list={rows:[{id:'old',avatar:'human photo'},{id:'new',avatar:'provider icon'}],scrollTop:250,querySelectorAll(){return [...this.rows]},appendChild(row){this.rows.splice(this.rows.indexOf(row),1);this.rows.push(row)}};
 }});
 vm.runInContext(fs.readFileSync('web/room-message-order.js','utf8'),context);
 return {context,messages,list:()=>list};
}
test('default and disabled flags keep chronological order and bottom behavior',()=>{
 for(const enabled of [undefined,false]){const f=fixture(enabled);f.context.render();assert.deepEqual(f.list().rows.map(r=>r.id),['old','new']);assert.equal(f.list().scrollTop,250);assert.equal(f.messages[0].id,'old')}
});
test('enabled order moves intact nodes, preserves avatar identity, and scrolls new messages to top',()=>{
 const f=fixture(true);f.context.render();assert.deepEqual(f.list().rows.map(r=>r.id),['new','old']);assert.equal(f.list().rows[0].avatar,'provider icon');assert.equal(f.list().rows[1].avatar,'human photo');assert.equal(f.list().scrollTop,0);assert.equal(f.messages[0].id,'old');
 f.list().scrollTop=80;f.context.render();assert.equal(f.list().scrollTop,80,'background refresh preserves browsing position');
 f.context.state.messages=Object.freeze([...f.messages,{id:'newer'}]);f.context.render();assert.equal(f.list().scrollTop,0);
 f.list().scrollTop=90;f.context.roomId='qa';f.context.render();assert.equal(f.list().scrollTop,0);
});
test('disabling the flag restores chronological rendering; empty and other views are safe',()=>{
 const f=fixture(true);f.context.render();f.context.state.featureFlags.newestMessagesFirst=false;f.context.render();assert.deepEqual(f.list().rows.map(r=>r.id),['old','new']);
 f.context.view='Settings';f.context.render();assert.deepEqual(f.list().rows.map(r=>r.id),['old','new']);
 const empty={querySelectorAll(){return []},appendChild(){throw new Error('unexpected row')}};f.context.arrangeNewestRoomMessages(empty);
});
test('feature is off by default in Settings and send/new-message scrolling respects its edge',()=>{
 const features=fs.readFileSync('web/features.js','utf8');assert(features.includes("flag==='newestMessagesFirst'?state.featureFlags?.[flag]===true"));assert(features.includes('Off by default: new room messages'));
 const app=fs.readFileSync('web/app.js','utf8');assert(app.includes('state.featureFlags?.newestMessagesFirst===true?0:list.scrollHeight'));assert(app.includes("state.featureFlags?.newestMessagesFirst===true?0:$('.message-list').scrollHeight"));
 const index=fs.readFileSync('web/index.html','utf8');assert(index.indexOf('room-message-order.js')>index.indexOf('profile-images.js'),'ordering runs after avatar decorations');
});
