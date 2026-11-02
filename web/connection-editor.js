'use strict';
// Loaded only when the server-side COMMONROOM_CONNECTION_EDITOR flag is enabled.
function connectionHint(protocol,location,provider){
 if(!protocol||!provider)return {label:'Worker or service address',placeholder:'Choose a provider and connection method first',text:'Choose a provider and connection method before entering an address.',example:'',status:'Configuration incomplete'};
 const remote=location!=='Local',claude=provider==='claude';
 if(protocol==='CLI')return remote?
  {label:'Container worker address',placeholder:claude?'ssh://commonroom-worker/claude':'ssh://commonroom-worker',text:'Use the approved Linux worker alias configured in CommonRoom. The worker runs the agent in an isolated container; it does not launch an existing host CLI. Sign in and test this connection after registration.',example:claude?'ssh://commonroom-worker/claude':'ssh://commonroom-worker',status:['openai','claude'].includes(provider)?'Requires worker setup and connection sign-in · not yet verified':'Provider adapter unavailable · registration only'}:
  {label:'Container provider',placeholder:claude?'ssh://commonroom-server/claude':'ssh://commonroom-server',text:'This runs on the CommonRoom server, not the computer running your browser. That server requires Linux, rootless Docker, a pinned agent image and COMMONROOM_SERVER_WORKER=true. Direct host execution is blocked.',example:claude?'ssh://commonroom-server/claude':'ssh://commonroom-server',status:['openai','claude'].includes(provider)?'Linux container worker required · not yet verified':'Provider adapter unavailable · registration only'};
 const scheme=protocol==='WebSocket'?(remote?'wss':'ws'):(remote?'https':'http');
 const path=protocol==='MCP'?'mcp':protocol==='WebSocket'?'ws':'api';
 const host=location==='Local'?'127.0.0.1:8080':location==='LAN'?'agent.internal:8080':'agent.example.com';
 return {label:'Service URL',placeholder:scheme+'://'+host+'/'+path,text:'Use the complete service URL, including its port and path. Chat website addresses cannot be connected directly.',example:scheme+'://'+host+'/'+path,status:'Registration only · room replies are not available yet'};
}
// Keep separate address drafts so switching methods never submits an SSH alias
// as a web URL. Hidden local commands are still included in FormData.
function connectionAddressModel(c){
 const key=(protocol,location,provider)=>protocol==='CLI'?(location==='Local'?'local:'+provider:'ssh:'+provider):protocol+':'+location;
 const drafts=new Map();let previous;
 if(c){const location=c.location||(c.endpoint?.startsWith('ssh://')?'LAN':'Local');drafts.set(key(c.protocol,location,connectionProvider(c)),c.endpoint||'')}
 return {next(protocol,location,provider,currentValue){
  if(previous!==undefined)drafts.set(previous,currentValue);
  const current=key(protocol,location,provider);
  const automatic=protocol==='CLI'&&location==='Local'&&['openai','claude'].includes(provider);
  let value=automatic?(provider==='claude'?'ssh://commonroom-server/claude':'ssh://commonroom-server'):(drafts.get(current)||'');
  // Preserve nonstandard saved local commands for explicit correction/editing.
  const saved=drafts.get(current);const customLocal=automatic&&saved&&saved!==value&&saved!==(provider==='claude'?'claude':'codex');
  if(customLocal)value=saved;
  previous=current;
  return {value,hidden:automatic&&!customLocal,ssh:protocol==='CLI'&&location!=='Local'};
 }};
}
function connectionFieldGuidance(field,draft,initial=false){
 const complete=['name','provider','protocol','location','endpoint'].every(key=>String(draft[key]||'').trim()!=='');
 const fields={
  name:{title:'Connection name',text:'Choose a label you will recognise, such as Codex coding agent. A model name is fine as a label, but does not select the model or provider.'},
  provider:{title:'Provider',text:'Choose the company or CLI that supplies the agent. For OpenAI select OpenAI / Codex. Provider sign-in happens after registration; do not enter credentials here.'},
  protocol:{title:'Connection method',text:'Choose how CommonRoom reaches the agent. Containerized CLI runs the supported provider through an isolated Linux worker. HTTP, WebSocket and MCP currently register service details only.'},
  location:{title:'Where the agent runs',text:'CommonRoom server means the machine hosting this application, not your browser. Another worker machine uses an approved worker address, on your network or over the internet. Managed execution requires Linux with rootless Docker; a Mac or Windows browser is fine.'},
  endpoint:{title:'Worker or service address',text:connectionHint(draft.protocol,draft.location,draft.provider).text}
 };
 const guidance=Object.hasOwn(fields,field)?fields[field]:null;
 return {title:initial||complete?'Set up your connection':guidance?.title||'Connection details',text:initial?'':guidance?.text||'',overview:initial||complete,configuration:!!draft.protocol&&!!draft.provider&&(complete||field==='endpoint')};
}
function wireConnectionFieldGuidance(editor,modal,readDraft){
 let initial=true,field='';
 const refresh=()=>{
  const guidance=connectionFieldGuidance(field,readDraft(),initial);
  modal.querySelector('#connection-guidance-title').textContent=guidance.title;
  modal.querySelector('#connection-overview').hidden=!guidance.overview;
  const detail=modal.querySelector('#connection-field-help');detail.textContent=guidance.text;detail.hidden=!guidance.text;
  modal.querySelector('#connection-configuration').hidden=!guidance.configuration;
 };
 const interact=event=>{if(!['name','provider','protocol','location','endpoint'].includes(event.target.name))return;initial=false;field=event.target.name;refresh()};
 for(const event of ['focusin','click','input','change'])editor.addEventListener(event,interact);
 refresh();return {refresh};
}
function connectionEditorFields(c){
 const location=c?.location||(c?.endpoint?.startsWith('ssh://')?'LAN':'Local');
 const provider=c?connectionProvider(c):'';
 const options=(items,current)=>items.map(([value,label])=>`<option value="${esc(value)}" ${current===value?'selected':''}>${esc(label)}</option>`).join('');
 return '<div class="connection-layout"><section class="connection-main" aria-label="Connection details"><p class="connection-intro">Choose your provider and where its worker runs.</p>'+field('Connection name','name',c?.name||'')+
 '<div class="connection-pair"><label class="field">Provider<select name="provider" required><option value="" disabled '+(!provider?'selected':'')+'>Choose a provider…</option>'+options([['openai','OpenAI / Codex'],['claude','Claude'],['gemini','Gemini'],['meta','Meta'],['deepseek','DeepSeek'],['generic','Other / custom']],provider)+'</select></label><label class="field">Connection method<select name="protocol" required><option value="" disabled '+(!c?.protocol?'selected':'')+'>Choose a connection method…</option>'+options([['CLI','Containerized CLI'],['HTTP','HTTP · registration only'],['WebSocket','WebSocket · registration only'],['MCP','MCP · registration only']],c?.protocol||'')+'</select></label></div>'+
 '<fieldset class="connection-location"><legend>Where is the worker or service?</legend>'+[['Local',globalThis.commonroomServerHostname||'Server name unavailable','CommonRoom runs on this machine, not your browser'],['LAN','Another worker machine','An approved Linux worker on your network or the internet'],...(c?.location==='Remote/Internet'?[['Remote/Internet','Remote / internet','Saved remote worker location']]:[])].map(([value,title,detail])=>`<label class="location-choice"><input type="radio" name="location" value="${value}" ${location===value?'checked':''}><span><strong>${esc(title)}</strong><small>${esc(detail)}</small></span></label>`).join('')+'</fieldset>'+
 '<label class="field"><span id="connection-address-label">Agent address</span><input name="endpoint" value="'+esc(c?.endpoint||'')+'" required aria-describedby="connection-address-help" autocomplete="off" spellcheck="false"></label>'+
 (c?'<p class="connection-edit-note">Changing the address or method starts fresh agent sessions. Saved messages remain intact.</p>':'')+'</section><aside class="connection-sidebar" aria-label="Setup guidance"><div class="connection-sidebar-heading"><h3 id="connection-guidance-title">Set up your connection</h3><button type="button" id="connection-ssh-help" aria-haspopup="dialog" aria-label="Open setup instructions">Setup guide</button></div><div id="connection-overview"><p class="connection-intro">Register a connection here, then sign in and test it before inviting agents. Registration alone does not mean the connection works.</p><ol class="connection-steps"><li><strong>Register</strong><span>Save the provider and worker address.</span></li><li><strong>Sign in</strong><span>Authenticate this connection with the provider.</span></li><li><strong>Test connection</strong><span>Verify a reply before inviting agents.</span></li></ol></div><p id="connection-field-help" class="connection-field-help" role="status" aria-live="polite" hidden></p><div id="connection-configuration"><h4>For this configuration</h4><div id="connection-address-help" class="connection-guidance" aria-live="polite"><span id="connection-adapter-status"></span><p id="connection-short-help"></p><button type="button" id="connection-use-example">Use example</button></div>'+
 '</div></aside></div>';
}
function connectionEditorTestRequest(data,saved){return {connectionId:saved?.id||'',draft:{name:String(data.get('name')||''),provider:String(data.get('provider')||''),protocol:String(data.get('protocol')||''),location:String(data.get('location')||''),endpoint:String(data.get('endpoint')||'').trim()}};}
// Testing switches the existing overlay body, retaining the form and its drafts.
function wireConnectionEditorTesting(modal,editor,endpoint,c){
 const actions=modal.querySelector('form>.actions'),layout=modal.querySelector('.connection-layout'),main=modal.querySelector('.connection-main'),sidebar=modal.querySelector('.connection-sidebar');
 const button=document.createElement('button');button.type='button';button.id='connection-editor-test';button.textContent='Test connection';actions.prepend(button);
 const panel=document.createElement('section');panel.className='connection-editor-test-panel';panel.hidden=true;panel.setAttribute('aria-label','Connection test');
 panel.innerHTML='<div class="connection-test-heading"><h3>Connection test</h3><button type="button" id="connection-test-back">Back to editing</button></div><div class="connection-editor-test-results" role="status" aria-live="polite" tabindex="-1"></div>';
 layout.append(panel);const output=panel.querySelector('.connection-editor-test-results');
 const show=testing=>{main.hidden=testing;sidebar.hidden=testing;panel.hidden=!testing;layout.classList.toggle('testing',testing);layout.scrollTop=0;button.textContent=testing?'Test again':'Test connection';if(testing)output.focus();else endpoint.focus()};
 panel.querySelector('#connection-test-back').onclick=()=>show(false);
 const payload=()=>connectionEditorTestRequest(new FormData(editor),c);
 const invalidate=()=>{output.textContent='Configuration changed. Test again to check the current values.';};
 editor.addEventListener('input',invalidate);editor.addEventListener('change',invalidate);modal.querySelector('#connection-use-example').addEventListener('click',invalidate);
 button.onclick=async()=>{const request=payload();if(!request.draft.endpoint){show(false);note('Enter a worker address or service URL before testing.');return}show(true);button.disabled=true;output.textContent='Testing current configuration…';try{const progress=connectionTestProgress(output);const result=await runConnectionTest(request,check=>{if(JSON.stringify(request)===JSON.stringify(payload()))progress(check)},modal);if(JSON.stringify(request)!==JSON.stringify(payload())){invalidate();return}renderConnectionTestResult(output,result,modal)}catch(error){output.textContent=error.message}finally{button.disabled=false}};
}
const previousConnectionEditor=connection;
connection=function(c){
 form(c?'Edit connection':'New connection',connectionEditorFields(c),d=>commit(c?'connection.updated':'connection.created',{...Object.fromEntries(d),...(c?{id:c.id,revision:c.revision||0}:{})}));
 const modal=document.querySelector('#modal');modal.classList.add('connection-editor');
 // The legacy provider wrapper prepends an icon field for Edit connection.
 const selects=modal.querySelectorAll('select[name="provider"]');if(selects.length>1)selects[0].closest('label').remove();
 modal.addEventListener('close',()=>modal.classList.remove('connection-editor'),{once:true});
 const editor=modal.querySelector('form'),endpoint=editor.elements.endpoint;
 let example='';const addresses=connectionAddressModel(c);
 const update=()=>{
  const protocol=editor.elements.protocol.value,location=editor.elements.location.value,provider=editor.elements.provider.value;
  const hint=connectionHint(protocol,location,provider);example=hint.example;
  const address=addresses.next(protocol,location,provider,endpoint.value);endpoint.value=address.value;endpoint.type=address.hidden?'hidden':'text';endpoint.required=!address.hidden;endpoint.closest('label').hidden=address.hidden;
  modal.querySelector('#connection-use-example').hidden=address.hidden;
  modal.querySelector('#connection-address-label').textContent=hint.label;endpoint.placeholder=hint.placeholder;
  modal.querySelector('#connection-short-help').textContent=hint.text+(address.hidden?' Command: '+address.value+'.':'');modal.querySelector('#connection-adapter-status').textContent=hint.status;
 };
 editor.addEventListener('change',update);update();
 wireConnectionFieldGuidance(editor,modal,()=>({name:editor.elements.name.value,provider:editor.elements.provider.value,protocol:editor.elements.protocol.value,location:editor.elements.location.value,endpoint:endpoint.value}));
 modal.querySelector('#connection-use-example').onclick=()=>{endpoint.value=example;endpoint.focus()};
 modal.querySelector('#connection-ssh-help').onclick=()=>openDocumentation(editor.elements.protocol.value==='CLI'?'connection-sign-in':'connection-testing');
 modal.querySelector('button.primary').textContent=c?'Save changes':'Register connection';
 if(state.currentActorId==='local-owner'&&state.featureFlags?.connectionTesting!==false){
 wireConnectionEditorTesting(modal,editor,endpoint,c);
 }

};
