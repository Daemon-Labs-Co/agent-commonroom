const assert=require('node:assert/strict');
const vm=require('node:vm');
const fs=require('node:fs');
const source=fs.readFileSync('web/connection-editor.js','utf8').split('const previousConnectionEditor=')[0];
const escape=value=>String(value??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
const context={esc:escape,field:(label,name,value)=>`<label>${label}<input name="${name}" value="${escape(value)}" required></label>`,connectionProvider:c=>c.provider||'generic'};
vm.createContext(context);vm.runInContext(source,context);
for(const [location,provider,example] of [['Local','openai','ssh://commonroom-server'],['Local','claude','ssh://commonroom-server/claude'],['LAN','openai','ssh://commonroom-worker'],['Remote/Internet','claude','ssh://commonroom-worker/claude']]){
 const hint=context.connectionHint('CLI',location,provider);assert.equal(hint.example,example);assert.match(hint.status,/not yet verified/);assert.ok(!hint.status.includes("Available for"));
}
for(const method of ['HTTP','WebSocket','MCP'])for(const location of ['Local','LAN','Remote/Internet']){
 const hint=context.connectionHint(method,location,'generic');assert.match(hint.status,/Registration only/);assert.match(hint.placeholder,location==='Local'?/127\.0\.0\.1:8080/:location==='LAN'?/agent\.internal:8080/:/agent\.example\.com/);
 assert.match(hint.text,/Chat website addresses cannot/);
}
context.commonroomServerHostname='debian-worker';
const newFields=context.connectionEditorFields();assert.match(newFields,/>debian-worker<\/strong>/);
context.commonroomServerHostname='<img src=x onerror=alert(1)>';assert.ok(!context.connectionEditorFields().includes('<img src=x'));
context.commonroomServerHostname='debian-worker';assert.equal((newFields.match(/value="" disabled selected/g)||[]).length,2);assert.ok(!/value="openai" selected/.test(newFields));assert.match(newFields,/value="Local" checked/);assert.match(newFields,/Registration alone does not mean the connection works/);assert.match(newFields,/aria-label="Open setup instructions"/);assert.ok(!newFields.includes("<details"));
const fields=context.connectionEditorFields({name:'<script>alert(1)</script>',endpoint:'ssh://custom/claude',provider:'claude',protocol:'CLI'});
assert.ok(!fields.includes('<script>'));assert.match(fields,/&lt;script&gt;/);assert.match(fields,/value="ssh:\/\/custom\/claude"/);assert.match(fields,/value="LAN" checked/);assert.match(fields,/value="claude" selected/);assert.match(fields,/Saved messages remain intact/);
console.log('Connection guidance, registration status, defaults, edit preservation and escaping checks passed.');
const model=context.connectionAddressModel();
let address=model.next('CLI','Local','openai','');assert.equal(address.hidden,true);assert.equal(address.value,'ssh://commonroom-server');assert.equal(address.ssh,false);
address=model.next('CLI','LAN','openai','codex');assert.equal(address.hidden,false);assert.equal(address.value,'');assert.equal(address.ssh,true);
address=model.next('HTTP','LAN','openai','ssh://custom');assert.equal(address.value,'');assert.equal(address.ssh,false);
address=model.next('CLI','LAN','openai','http://agent.internal/api');assert.equal(address.value,'ssh://custom');
address=model.next('HTTP','LAN','openai','ssh://custom');assert.equal(address.value,'http://agent.internal/api');
address=model.next('CLI','Local','claude',address.value);assert.equal(address.hidden,true);assert.equal(address.value,'ssh://commonroom-server/claude');
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

// Exercise field focus/click/change against the real guidance controller.
{
 const nodes=Object.fromEntries(['connection-guidance-title','connection-overview','connection-field-help','connection-configuration'].map(id=>['#'+id,{hidden:false,textContent:''}]));
 const handlers={};
 const editor={addEventListener(name,fn){(handlers[name]??=[]).push(fn)}};
 const draft={name:'',provider:'',protocol:'',location:'Local',endpoint:''};
 context.wireConnectionFieldGuidance(editor,{querySelector:s=>nodes[s]},()=>draft);
 const interact=(event,name)=>handlers[event].forEach(fn=>fn({target:{name}}));
 assert.equal(nodes['#connection-overview'].hidden,false);
 assert.equal(nodes['#connection-guidance-title'].textContent,'Set up your connection');
 interact('focusin','name');
 assert.equal(nodes['#connection-overview'].hidden,true);
 assert.match(nodes['#connection-field-help'].textContent,/label you will recognise/);
 assert.equal(nodes['#connection-guidance-title'].textContent,'Connection name');
 interact('click','provider');
 assert.equal(nodes['#connection-guidance-title'].textContent,'Provider');
 assert.match(nodes['#connection-field-help'].textContent,/OpenAI \/ Codex/);
 draft.provider='openai';interact('change','provider');
 assert.equal(nodes['#connection-overview'].hidden,true);
 assert.equal(nodes['#connection-configuration'].hidden,true);
 interact('focusin','protocol');assert.match(nodes['#connection-field-help'].textContent,/Containerized CLI/);
 draft.protocol='CLI';interact('change','protocol');
 interact('focusin','location');assert.match(nodes['#connection-field-help'].textContent,/not your browser/);
 draft.location='LAN';interact('change','location');
 interact('focusin','endpoint');assert.equal(nodes['#connection-configuration'].hidden,false);
 draft.name='Codex';draft.endpoint='ssh://approved-worker';interact('input','endpoint');
 assert.equal(nodes['#connection-overview'].hidden,false);
 assert.equal(nodes['#connection-guidance-title'].textContent,'Set up your connection');
 draft.name=' ';interact('input','name');
 assert.equal(nodes['#connection-overview'].hidden,true);
 assert.equal(nodes['#connection-guidance-title'].textContent,'Connection name');
 assert.ok(!newFields.includes('This computer'));
 assert.match(newFields,/debian-worker/);
 assert.match(newFields,/Another worker machine/);
 assert.equal(context.connectionHint('','Local','').example,'');
 const remoteFields=context.connectionEditorFields({name:'Existing remote',provider:'openai',protocol:'CLI',location:'Remote/Internet',endpoint:'ssh://saved-worker'});
 assert.match(remoteFields,/value="Remote\/Internet" checked/);
 console.log('Field focus/click guidance, empty defaults, incomplete/complete overview and saved remote locations passed.');
}

// Exercise the actual test-view controller with in-memory form/API doubles.
(async()=>{
 function element(){return {hidden:false,disabled:false,textContent:'',innerHTML:'',scrollTop:99,handlers:{},classList:{toggle(){}},setAttribute(){},addEventListener(k,f){this.handlers[k]=f},focus(){this.focused=true}}}
 const main=element(),sidebar=element(),layout=element(),output=element(),back=element(),example=element(),endpoint=element(),actions=element();
 let button,panel;actions.prepend=b=>button=b;layout.append=p=>panel=p;
 const selectors={'form>.actions':actions,'.connection-layout':layout,'.connection-main':main,'.connection-sidebar':sidebar,'#connection-use-example':example};
 const modal={querySelector:s=>selectors[s]},editor=element();
 const values=new Map([['name','Draft'],['endpoint','ssh://worker'],['provider','claude'],['location','LAN'],['protocol','CLI']]);
 context.document={createElement(tag){const e=element();if(tag==='section')e.querySelector=s=>s==='.connection-editor-test-results'?output:back;return e}};
 context.FormData=function(){return values};context.note=()=>{};context.renderConnectionTestResult=(o,r)=>{o.innerHTML='Outcome: '+r.status};
 context.connectionTestProgress=()=>()=>{};context.runConnectionTest=request=>context.api('connections/test',request);
 let calls=0;context.api=async(path,request)=>{calls++;assert.equal(path,'connections/test');assert.equal(request.draft.name,'Draft');return {status:'needs-sign-in'}};
 context.wireConnectionEditorTesting(modal,editor,endpoint,{id:'saved'});
 assert.equal(panel.hidden,true);await button.onclick();assert.equal(main.hidden,true);assert.equal(sidebar.hidden,true);assert.equal(panel.hidden,false);assert.equal(output.innerHTML,'Outcome: needs-sign-in');assert.equal(button.textContent,'Test again');assert.equal(layout.scrollTop,0);
 back.onclick();assert.equal(main.hidden,false);assert.equal(sidebar.hidden,false);assert.equal(panel.hidden,true);assert.equal(values.get('name'),'Draft');assert.equal(button.textContent,'Test connection');assert.equal(endpoint.focused,true);
 context.api=async()=>{throw Error('Worker unavailable')};await button.onclick();assert.equal(output.textContent,'Worker unavailable');assert.equal(button.disabled,false);
 context.api=async()=>{values.set('name','Changed while testing');return {status:'working'}};await button.onclick();assert.match(output.textContent,/Configuration changed/);
 values.set('endpoint','');await button.onclick();assert.equal(panel.hidden,true);assert.equal(calls,1);
 console.log('Test view transitions, retained drafts, errors, stale results and missing address passed.');
})().catch(e=>{console.error(e);process.exitCode=1});

for(const provider of ['openai','claude']){
 const selected=context.connectionAddressModel().next('CLI','Local',provider,'');
 const data=new Map([['name','Server fixture'],['provider',provider],['protocol','CLI'],['location','Local'],['endpoint',selected.value]]);
 const request=context.connectionEditorTestRequest(data);
 assert.equal(request.draft.endpoint,provider==='claude'?'ssh://commonroom-server/claude':'ssh://commonroom-server');
 assert.equal(request.draft.location,'Local');assert.equal(selected.hidden,true);
}
console.log('Server-machine selection reaches the connection-test request with the isolated worker endpoint.');
