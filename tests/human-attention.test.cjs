const test=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
function node(){return {children:[],attrs:{},classList:{values:new Set(),add(v){this.values.add(v)}},append(n){this.children.push(n)},setAttribute(k,v){this.attrs[k]=v},after(n){this.afterNode=n}}}
function harness(){
 const meta=node(),content=node(),row=node(),heading=node(),nav=node();row.dataset={messageId:'request'};row.querySelector=s=>s==='.message-meta'?meta:content;row.getBoundingClientRect=()=>({top:300});row.focus=opts=>{row.focused=opts};nav.dataset={room:'engineering'};
 const list={scrollTop:100,getBoundingClientRect:()=>({top:200})};const calls=[],notes=[];let refreshes=0;
 const ctx={roomId:'engineering',view:'Room',render(){},state:{featureFlags:{},rooms:[{id:'engineering',pendingHumanRequests:1}],messages:[{id:'request',actorType:'agent',humanRequest:'explicit'}]},document:{createElement:node,querySelectorAll(s){return s==='[data-room]'?[nav]:[row]},querySelector(s){return s==='.message-list'?list:heading}},api:async(path,body)=>{calls.push({path,body})},refresh:async()=>{refreshes++},note:t=>notes.push(t)};
 vm.createContext(ctx);vm.runInContext(fs.readFileSync('web/human-attention.js','utf8'),ctx);return {ctx,row,meta,content,heading,nav,list,calls,notes,refreshes:()=>refreshes};
}
test('requests get visible badges, room counts and a direct attention action',async()=>{
 const h=harness();h.ctx.drawHumanAttention();assert(h.row.classList.values.has('attention-required'));assert.equal(h.meta.children[0].textContent,'Needs your attention');assert.equal(h.nav.children.at(-1).textContent,'1');assert.match(h.heading.afterNode.children[0].textContent,/1 message needs/);
 h.heading.afterNode.children[1].onclick();assert.equal(h.list.scrollTop,188);assert.equal(h.row.focused.preventScroll,true);
 await h.content.children[0].onclick();assert.equal(h.calls[0].path,'events');assert.equal(h.calls[0].body.kind,'message.attention.acknowledged');assert.equal(h.calls[0].body.payload.messageId,'request');assert.equal(h.refreshes(),1);
});
test('suggestions are distinguished and acknowledgment is separate from approval',()=>{
 const h=harness();h.ctx.state.messages[0].humanRequest='suggested';h.ctx.drawHumanAttention();assert.equal(h.meta.children[0].textContent,'May need your attention');
 const seen=harness();seen.ctx.state.messages[0].attentionAcknowledgedBy='human';seen.ctx.drawHumanAttention();assert(seen.row.classList.values.has('attention-handled'));assert.equal(seen.meta.children[0].textContent,'Acknowledged');assert.equal(seen.content.children.length,0);
});
test('disabled and human messages do not create attention controls; failed acknowledgment stays visible',async()=>{
 const h=harness();h.ctx.state.featureFlags.humanAttention=false;h.ctx.drawHumanAttention();assert.equal(h.content.children.length,0);assert.equal(h.heading.afterNode,undefined);
 const human=harness();human.ctx.state.messages[0].actorType='human';human.ctx.drawHumanAttention();assert.equal(human.content.children.length,0);
 const failed=harness();failed.ctx.api=async()=>{throw Error('Room access denied')};await failed.ctx.acknowledgeHumanRequest('request');assert.deepEqual(failed.notes,['Room access denied']);assert.equal(failed.refreshes(),0);
});
