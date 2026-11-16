const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const source=fs.readFileSync('web/manager-assignments.js','utf8');
const calls=[];let renders=0;
const ctx={view:'Tasks',roomId:'r1',state:{featureFlags:{},rooms:[],currentActorId:'local-owner'},render:()=>{renders++},commit:async(k,p)=>calls.push([k,p]),form:(t,h,save)=>{ctx.title=t;ctx.html=h;ctx.save=save},document:{querySelector:()=>null}};
vm.createContext(ctx);vm.runInContext(source,ctx);
const limit=vm.runInContext('roomAssignmentLimit',ctx);
assert.equal(limit({}),5,'default');assert.equal(limit({assignmentLimit:0}),0,'zero is a real limit');assert.equal(limit({assignmentLimit:12}),12);assert.equal(limit({assignmentLimit:99}),5);
const used=vm.runInContext('assignmentsUsedToday',ctx);
assert.equal(used({assignmentDay:'2026-11-10',assignmentsToday:3},'2026-11-10'),3);assert.equal(used({assignmentDay:'2026-11-09',assignmentsToday:3},'2026-11-10'),0,'yesterday does not count');
(async()=>{
 vm.runInContext('editAssignmentLimit',ctx)({assignmentLimit:7});
 assert.match(ctx.html,/value="7"/);assert.match(ctx.html,/max="50"/);
 await ctx.save({get:()=>'9'});
 assert.equal(JSON.stringify(calls),JSON.stringify([['room.assignment_limit.updated',{limit:9}]]));
 vm.runInContext('render()',ctx);assert.equal(renders,1);
 assert.ok(fs.readFileSync('web/index.html','utf8').includes('manager-assignments.js'));
 console.log('Manager assignments: limit defaults, daily count and limit change passed.');
})();
