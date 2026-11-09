const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const features=fs.readFileSync('web/features.js','utf8');
const a=features.indexOf('function agentModelField'),b=features.indexOf('function openContextPolicy');
const calls=[],commits=[];
const modal={actions:{prepend(n){calls.push('prepend:'+n.textContent)}},close(){calls.push('close')}};
const ctx={state:{participants:[],featureFlags:{}},esc:s=>String(s),profile:p=>calls.push('profile:'+p.name),
 $:()=>modal.actions,commit:async(k,p)=>commits.push([k,p]),
 document:{createElement:()=>({set onclick(f){this.f=f}})},
 form:(title,html,save)=>{ctx.saved=save;ctx.html=html;ctx.title=title}};
vm.createContext(ctx);vm.runInContext(features.slice(a,b),ctx);
(async()=>{
 const agent={id:'p1',name:'Alex',model:'m-1',revision:2,rank:1};
 ctx.openAgentModelOverlay(agent);
 assert.equal(ctx.title,'Model for Alex');assert.match(ctx.html,/Current model: <strong>m-1<\/strong>/);assert.match(ctx.html,/name="model"/);
 assert.ok(calls.some(c=>c==='prepend:Full profile…'),'full profile stays reachable');
 await ctx.saved({get:()=>' m-1 '});
 assert.equal(commits.length,0,'unchanged model changes nothing');
 await ctx.saved({get:()=>' m-2 '});
 assert.equal(JSON.stringify(commits),JSON.stringify([['participant.updated',{...agent,model:'m-2'}]]));
 ctx.saved({get:()=>''});await new Promise(r=>setTimeout(r));
 assert.equal(commits[1][1].model,'','blank returns to the provider default');
 const squad=fs.readFileSync('web/squad-profiles.js','utf8');
 assert.ok(squad.includes('openAgentModelOverlay(agent)')&&squad.includes("agentModelSelection!==false"),'agent name opens the overlay only while model choice is on');
 console.log('Agent model overlay switches, clears and ignores unchanged models.');
})();
