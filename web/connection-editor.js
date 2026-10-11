'use strict';
// Loaded only when the server-side COMMONROOM_CONNECTION_EDITOR flag is enabled.
function connectionHint(protocol,location,provider){
 const remote=location!=='Local',claude=provider==='claude';
 if(protocol==='CLI')return remote?
  {label:'SSH endpoint',placeholder:claude?'ssh://debian-claude/claude':'ssh://jerry',text:'Use an SSH alias configured for the account running CommonRoom. No username, port, or password goes here.',example:claude?'ssh://debian-claude/claude':'ssh://jerry',status:'Available for Codex and Claude'}:
  {label:'Local command',placeholder:claude?'claude':'codex',text:'The CLI must be installed and signed in on the computer running CommonRoom. No port is needed.',example:claude?'claude':'codex',status:'Available for Codex and Claude'};
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
  let value=automatic?(provider==='claude'?'claude':'codex'):(drafts.get(current)||'');
  // Preserve nonstandard saved local commands for explicit correction/editing.
  const saved=drafts.get(current);const customLocal=automatic&&saved&&saved!==value;
  if(customLocal)value=saved;
  previous=current;
  return {value,hidden:automatic&&!customLocal,ssh:protocol==='CLI'&&location!=='Local'};
 }};
}
function connectionEditorFields(c){
 const location=c?.location||(c?.endpoint?.startsWith('ssh://')?'LAN':'Local');
 const provider=c?connectionProvider(c):'openai';
 const options=(items,current)=>items.map(([value,label])=>`<option value="${esc(value)}" ${current===value?'selected':''}>${esc(label)}</option>`).join('');
 return '<p class="connection-intro">Register an agent here, then invite it to a room.</p>'+field('Connection name','name',c?.name||'')+
 '<div class="connection-pair"><label class="field">Provider<select name="provider">'+options([['openai','OpenAI / Codex'],['claude','Claude'],['gemini','Gemini'],['meta','Meta'],['deepseek','DeepSeek'],['generic','Other / custom']],provider)+'</select></label><label class="field">Connection method<select name="protocol">'+options([['CLI','CLI program'],['HTTP','HTTP · registration only'],['WebSocket','WebSocket · registration only'],['MCP','MCP · registration only']],c?.protocol||'CLI')+'</select></label></div>'+
 '<fieldset class="connection-location"><legend>Where does it run?</legend>'+[['Local','This computer','The CommonRoom server'],['LAN','Internal network','A reachable machine on your LAN'],['Remote/Internet','Remote / internet','Outside your internal network']].map(([value,title,detail])=>`<label class="location-choice"><input type="radio" name="location" value="${value}" ${location===value?'checked':''}><span><strong>${title}</strong><small>${detail}</small></span></label>`).join('')+'</fieldset>'+
 '<label class="field"><span id="connection-address-label">Agent address</span><input name="endpoint" value="'+esc(c?.endpoint||'')+'" required aria-describedby="connection-address-help" autocomplete="off" spellcheck="false"></label>'+
 '<div id="connection-address-help" class="connection-guidance" aria-live="polite"><span id="connection-adapter-status"></span><p id="connection-short-help"></p><button type="button" id="connection-use-example">Use example</button></div>'+
 '<details class="connection-details"><summary>Setup instructions</summary><p>CLI integrations support Codex and Claude saved sessions. Local commands are <code>codex</code> and <code>claude</code>. Remote aliases use <code>ssh://alias</code> for Codex or <code>ssh://alias/claude</code> for Claude. Configure keys, usernames and ports in the runtime account’s SSH config.</p><p>HTTP, WebSocket and MCP can be registered, but their adapters are not implemented. Other provider icons do not add executable adapters.</p><p>Examples are placeholders. Never put passwords, keys or tokens in the address.</p><button type="button" id="connection-ssh-help">SSH setup guide</button></details>'+
 (c?'<p class="connection-edit-note">Changing the address or method starts fresh agent sessions. Saved messages remain intact.</p>':'');
}
const previousConnectionEditor=connection;
connection=function(c){
 form(c?'Edit connection':'New connection',connectionEditorFields(c),d=>commit(c?'connection.updated':'connection.created',{...Object.fromEntries(d),...(c?{id:c.id}:{})}));
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
  modal.querySelector('#connection-use-example').hidden=address.hidden;modal.querySelector('#connection-ssh-help').hidden=!address.ssh;
  modal.querySelector('#connection-address-label').textContent=hint.label;endpoint.placeholder=hint.placeholder;
  modal.querySelector('#connection-short-help').textContent=hint.text+(address.hidden?' Command: '+address.value+'.':'');modal.querySelector('#connection-adapter-status').textContent=hint.status;
 };
 editor.addEventListener('change',update);update();
 modal.querySelector('#connection-use-example').onclick=()=>{endpoint.value=example;endpoint.focus()};
 modal.querySelector('#connection-ssh-help').onclick=()=>openDocumentation('ssh');
 modal.querySelector('button.primary').textContent=c?'Save changes':'Add connection';
};
