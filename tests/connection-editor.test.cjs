const assert=require('node:assert/strict');
const vm=require('node:vm');
const fs=require('node:fs');
const source=fs.readFileSync('web/connection-editor.js','utf8').split('const previousConnectionEditor=')[0];
const escape=value=>String(value??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
const context={esc:escape,field:(label,name,value)=>`<label>${label}<input name="${name}" value="${escape(value)}" required></label>`,connectionProvider:c=>c.provider||'generic'};
vm.createContext(context);vm.runInContext(source,context);
for(const [location,provider,example] of [['Local','openai','codex'],['Local','claude','claude'],['LAN','openai','ssh://commonroom-worker'],['Remote/Internet','claude','ssh://commonroom-worker/claude']]){
 const hint=context.connectionHint('CLI',location,provider);assert.equal(hint.example,example);assert.match(hint.status,/not yet verified/);assert.ok(!hint.status.includes("Available for"));
}
for(const method of ['HTTP','WebSocket','MCP'])for(const location of ['Local','LAN','Remote/Internet']){
 const hint=context.connectionHint(method,location,'generic');assert.match(hint.status,/Registration only/);assert.match(hint.placeholder,location==='Local'?/127\.0\.0\.1:8080/:location==='LAN'?/agent\.internal:8080/:/agent\.example\.com/);
 assert.match(hint.text,/Chat website addresses cannot/);
}
const newFields=context.connectionEditorFields();assert.match(newFields,/value="openai" selected/);assert.match(newFields,/value="LAN" checked/);assert.match(newFields,/Registration alone does not mean the connection works/);assert.match(newFields,/aria-label="Open setup instructions"/);assert.ok(!newFields.includes("<details"));
const fields=context.connectionEditorFields({name:'<script>alert(1)</script>',endpoint:'ssh://custom/claude',provider:'claude',protocol:'CLI'});
assert.ok(!fields.includes('<script>'));assert.match(fields,/&lt;script&gt;/);assert.match(fields,/value="ssh:\/\/custom\/claude"/);assert.match(fields,/value="LAN" checked/);assert.match(fields,/value="claude" selected/);assert.match(fields,/Saved messages remain intact/);
console.log('Connection guidance, registration status, defaults, edit preservation and escaping checks passed.');
const model=context.connectionAddressModel();
let address=model.next('CLI','Local','openai','');assert.equal(address.hidden,true);assert.equal(address.value,'codex');assert.equal(address.ssh,false);
address=model.next('CLI','LAN','openai','codex');assert.equal(address.hidden,false);assert.equal(address.value,'');assert.equal(address.ssh,true);
address=model.next('HTTP','LAN','openai','ssh://custom');assert.equal(address.value,'');assert.equal(address.ssh,false);
address=model.next('CLI','LAN','openai','http://agent.internal/api');assert.equal(address.value,'ssh://custom');
address=model.next('HTTP','LAN','openai','ssh://custom');assert.equal(address.value,'http://agent.internal/api');
address=model.next('CLI','Local','claude',address.value);assert.equal(address.hidden,true);assert.equal(address.value,'claude');
const saved=context.connectionAddressModel({provider:'openai',protocol:'CLI',location:'Local',endpoint:'custom-command'});address=saved.next('CLI','Local','openai','custom-command');assert.equal(address.hidden,false);assert.equal(address.value,'custom-command');
console.log('Conditional addresses, automatic local commands and per-method draft restoration passed.');

assert.match(context.connectionHint("CLI","LAN","generic").status,/adapter unavailable/);
assert.ok(!newFields.includes("installed and signed in"));

assert.match(newFields,/class="connection-layout"/);assert.match(newFields,/aria-label="Connection details"/);assert.match(newFields,/aria-label="Setup guidance"/);
assert.ok(newFields.indexOf('name="endpoint"')<newFields.indexOf('connection-sidebar'));
assert.ok(newFields.indexOf('id="connection-address-help"')>newFields.indexOf('connection-sidebar'));
for(const name of ['name','provider','protocol','location','endpoint'])assert.ok(newFields.includes('name="'+name+'"'));
console.log('Two-column connection form keeps editable fields, contextual guidance and setup steps.');
const draftData=new Map([['name','Unsaved name'],['protocol','CLI'],['provider','claude'],['location','LAN'],['endpoint',' ssh://commonroom-worker/claude ']]);
const createTest=context.connectionEditorTestRequest(draftData);assert.equal(createTest.connectionId,'');assert.equal(createTest.draft.endpoint,'ssh://commonroom-worker/claude');assert.equal(createTest.draft.name,'Unsaved name');assert.ok(!('id' in createTest.draft));
const editTest=context.connectionEditorTestRequest(draftData,{id:'saved-id',endpoint:'ssh://old-worker'});assert.equal(editTest.connectionId,'saved-id');assert.equal(editTest.draft.endpoint,'ssh://commonroom-worker/claude');assert.equal(editTest.draft.provider,'claude');
console.log('Create/edit tests use current unsaved fields and only the saved credential identity.');

assert.ok(newFields.indexOf('id="connection-ssh-help"')<newFields.indexOf('connection-steps'));
assert.equal((newFields.match(/id="connection-ssh-help"/g)||[]).length,1);
console.log('Setup guide is directly accessible above the steps without expanding the form.');
