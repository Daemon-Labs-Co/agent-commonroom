const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const source=fs.readFileSync('web/task-suggestions.js','utf8');
const calls=[];let renders=0;
const ctx={view:'Tasks',roomId:'r1',state:{featureFlags:{},actors:[]},render:()=>{renders++},note:t=>calls.push(['note',t]),refresh:async()=>calls.push(['refresh']),api:async(p,b)=>{calls.push([p,b]);return{}},commit:async()=>{},document:{querySelector:()=>null},encodeURIComponent};
vm.createContext(ctx);vm.runInContext(source,ctx);
const list=[{id:'m#1',roomId:'r1',status:'open',author:'Avery',title:'Add login',criteria:'Users can sign in',assignee:'Blake',messageId:'m'},{id:'m#2',roomId:'r1',status:'captured'},{id:'n#1',roomId:'r2',status:'open'},{id:'m#3',roomId:'r1',status:'dismissed'}];
assert.equal(vm.runInContext('openTaskSuggestions',ctx)(list,'r1').length,1,'only open suggestions in this room');
assert.equal(vm.runInContext('suggestionLine',ctx)(list[0]),'Avery suggests: Add login (for Blake) — Users can sign in');
assert.equal(vm.runInContext('suggestionLine',ctx)({author:'A',title:'T'}),'A suggests: T');
(async()=>{
 await vm.runInContext('captureTaskSuggestion',ctx)(list[0]);
 assert.equal(JSON.stringify(calls[0]),JSON.stringify(['task-proposals?room=r1',{action:'capture',sourceMessageId:'m',title:'Add login',criteria:'Users can sign in'}]),'capture uses the existing proposal endpoint');
 vm.runInContext('render()',ctx);assert.equal(renders,1);
 assert.ok(fs.readFileSync('web/index.html','utf8').includes('task-suggestions.js'));
 console.log('Task suggestions: filtering, summary and capture request passed.');
})();
