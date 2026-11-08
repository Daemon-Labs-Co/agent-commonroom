const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const source=fs.readFileSync('web/features.js','utf8');
const helper=source.includes('function roomHumanDisplayName')?source.slice(source.indexOf('function roomHumanDisplayName'),source.indexOf('// Keep the badge')):'';
const expression=source.match(/const name=(.*?);const title=person.title/)[1];
const ctx={person:{id:'local-owner',name:'Example Owner'},human:{humanName:'Change Me'}};vm.createContext(ctx);vm.runInContext(helper,ctx);
const name=()=>vm.runInContext(expression,ctx);
for(const value of ['Change Me','You','', '  Change Me  ']){ctx.human.humanName=value;assert.equal(name(),'Example Owner','placeholder must inherit saved profile')}
ctx.human.humanName='Room-specific name';assert.equal(name(),'Room-specific name');
ctx.person={id:'another-human',name:'Example Colleague'};assert.equal(name(),'Example Colleague');
assert.ok(source.includes("field('Name in this room','humanName'"));assert.ok(source.includes('Settings → Your profile'));
console.log('Room names inherit the owner profile for placeholders and preserve intentional room/member names.');
// Saving the room profile form must not freeze the displayed profile name.
const choice=vm.runInContext('roomHumanNameChoice',ctx);
const owner={id:'local-owner',name:'New Name'};
assert.deepEqual({...choice(owner,{humanName:''},'New Name',false)},{humanName:'',useProfileName:true},'untouched save while inheriting stays inherited');
assert.deepEqual({...choice(owner,{humanName:'Change Me'},'New Name',false)},{humanName:'',useProfileName:true});
assert.deepEqual({...choice(owner,{humanName:'Old Name'},'Old Name',false)},{humanName:'Old Name'},'untouched save keeps an intentional override');
assert.deepEqual({...choice(owner,{humanName:'Old Name'},'Old Name',true)},{humanName:'',useProfileName:true},'checkbox resets a stuck override');
assert.deepEqual({...choice(owner,{humanName:''},'Room Only',false)},{humanName:'Room Only'},'typing a name sets a room override');
assert.deepEqual({...choice(owner,{humanName:'Old Name'},'',false)},{humanName:'Old Name'},'empty entry keeps the stored name');
console.log('Room profile saves keep following the profile unless a room name is chosen.');
