const test=require('node:test'),assert=require('node:assert/strict'),vm=require('node:vm'),fs=require('node:fs');
function fixture(){
 const observers=[];
 const ctx=vm.createContext({MutationObserver:class{constructor(fn){this.fn=fn;observers.push(this)}observe(target,options){this.target=target;this.options=options}disconnect(){this.disconnected=true}}});
 const source=fs.readFileSync('web/features.js','utf8');
 vm.runInContext(source.slice(0,source.indexOf('let roomId=')),ctx);
 function squad(n){const badge={setAttribute(k,v){this[k]=v}},members={rows:Array(n).fill({}),querySelectorAll(selector){assert.equal(selector,':scope > .participant');return this.rows}};return {badge,members,querySelector(selector){return selector==='.squad-members'?members:badge}}}
 return {ctx,observers,squad};
}
test('badge follows agent invites/removals and human joins/leaves without a full render',()=>{
 const f=fixture(),s=f.squad(5);f.ctx.trackSquadCount(s);assert.equal(s.badge.textContent,'5');
 for(const n of [6,5,4,1,0]){s.members.rows=Array(n).fill({});f.observers[0].fn();assert.equal(s.badge.textContent,String(n));assert.equal(s.badge['aria-label'],n+' squad member'+(n===1?'':'s'))}
 assert.equal(f.observers[0].options.childList,true);assert.equal(f.observers[0].options.subtree,undefined,'profile/menu changes are not membership changes');
});
test('room changes and leaving the room disconnect stale observers',()=>{
 const f=fixture(),a=f.squad(5),b=f.squad(2);f.ctx.trackSquadCount(a);f.ctx.trackSquadCount(b);assert.equal(f.observers[0].disconnected,true);assert.equal(b.badge.textContent,'2');f.ctx.trackSquadCount(null);assert.equal(f.observers[1].disconnected,true);
});
test('render binds the count after inserting human and agent rows, regardless of cosmetic flags',()=>{
 const source=fs.readFileSync('web/features.js','utf8');assert(source.includes('trackSquadCount(null);'));assert(source.includes("squad.querySelectorAll('.participant').forEach(row=>members.append(row));trackSquadCount(squad);"));assert(!source.includes("squad.querySelector('.count').textContent=state.participants.length+people.length"));
});
