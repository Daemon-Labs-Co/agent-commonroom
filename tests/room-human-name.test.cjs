const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const source=fs.readFileSync('web/features.js','utf8');
const helper=source.includes('function roomHumanDisplayName')?source.slice(source.indexOf('function roomHumanDisplayName'),source.indexOf('// Keep the badge')):'';
const expression=source.match(/const name=(.*?);const title=person.title/)[1];
const ctx={person:{id:'local-owner',name:'Example Owner'},human:{humanName:'Change Me'}};vm.createContext(ctx);vm.runInContext(helper,ctx);
const name=()=>vm.runInContext(expression,ctx);
for(const value of ['Change Me','You','', '  Change Me  ']){ctx.human.humanName=value;assert.equal(name(),'Example Owner','placeholder must inherit saved profile')}
ctx.human.humanName='Room-specific name';assert.equal(name(),'Room-specific name');
ctx.person={id:'another-human',name:'Example Colleague'};assert.equal(name(),'Example Colleague');
assert.ok(source.includes("field('Your name','humanName'"));
console.log('Room names inherit the owner profile for placeholders and preserve intentional room/member names.');
// The room form edits the one profile name; room-only names are released.
const choice=vm.runInContext('roomHumanNameChoice',ctx);
const owner={id:'local-owner',name:'Change Me'};
assert.deepEqual({...choice(owner,{humanName:'Mark'},'Mark')},{humanName:'',useProfileName:true,profileName:'Mark'},'a name shown only by a room override becomes the profile name');
assert.deepEqual({...choice({id:'local-owner',name:'Mark'},{humanName:''},'Mark')},{humanName:'',useProfileName:true},'unchanged name changes no profile');
assert.deepEqual({...choice({id:'local-owner',name:'Mark'},{humanName:''},'Mark D')},{humanName:'',useProfileName:true,profileName:'Mark D'});
assert.deepEqual({...choice({id:'local-owner',name:'Mark'},{humanName:'Old'},'')},{humanName:'',useProfileName:true},'empty entry changes no profile');
console.log('Room profile saves update the one profile name.');
