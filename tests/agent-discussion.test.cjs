const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const source=fs.readFileSync('web/agent-discussion.js','utf8');
const calls=[];let renders=0;
const primary={};
const ctx={view:'Tasks',roomId:'r1',state:{featureFlags:{},rooms:[],actors:[]},render:()=>{renders++},esc:s=>String(s),note:t=>calls.push(['note',t]),refresh:async()=>calls.push(['refresh']),
 api:async(p,b)=>{calls.push([p,b]);return{agents:2,rounds:b.rounds}},commit:async(k,p)=>calls.push([k,p]),
 form:(t,h,save)=>{ctx.title=t;ctx.html=h;ctx.save=save},$:()=>primary,document:{querySelector:()=>null}};
vm.createContext(ctx);vm.runInContext(source,ctx);
const limit=vm.runInContext('roomDiscussionLimit',ctx);
assert.equal(limit({}),3);assert.equal(limit({discussionRounds:1}),1);assert.equal(limit({discussionRounds:9}),3,'a stored value can never raise the ceiling');
assert.equal((vm.runInContext('discussionRoundOptions',ctx)(2).match(/<option/g)||[]).length,2);
(async()=>{
 vm.runInContext('startAgentDiscussion',ctx)({discussionRounds:2});
 assert.equal(ctx.title,'Start a discussion');assert.doesNotMatch(ctx.html,/value="3"/,'room limit bounds the choice');assert.equal(primary.textContent,'Start discussion');
 await ctx.save({get:k=>({text:'Pick a database',rounds:'2'})[k]});
 assert.equal(JSON.stringify(calls[0]),JSON.stringify(['agents',{action:'discuss',text:'Pick a database',rounds:2}]));
 assert.match(calls[2][1],/2 agents for up to 2 rounds/);
 calls.length=0;vm.runInContext('editDiscussionLimit',ctx)({});await ctx.save({get:()=>'1'});
 assert.equal(JSON.stringify(calls[0]),JSON.stringify(['room.discussion_limit.updated',{rounds:1}]));
 vm.runInContext('render()',ctx);assert.equal(renders,1);
 assert.ok(fs.readFileSync('web/index.html','utf8').includes('agent-discussion.js'));
 console.log('Agent discussion: limit, start request and room limit change passed.');
})();
