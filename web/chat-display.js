// Personal chat display choices: chat style display (your own messages on the
// right) and highlighted agents (tint plus a marker). Display only: stored on
// the signed-in human's profile, applied in every room, never sent to agents
// and never an authority signal. Message order, text and records are untouched.
function chatDisplayEnabled(){return typeof state!=='undefined'&&!!state&&state.featureFlags?.chatDisplayChoices!==false}
function chatDisplaySelf(){const a=state?.actors?.find(a=>a.id===state.currentActorId);return a&&a.kind==='human'?a:null}
function chatStyleOn(){return chatDisplayEnabled()&&chatDisplaySelf()?.chatStyle===true}
function highlightedAgentIds(){return chatDisplayEnabled()?(chatDisplaySelf()?.highlightedAgents||[]):[]}
function agentActorId(p){return p.actorId||('bot-'+p.connectionId)}
function isOwnMessage(m){
 if(m.actorId)return m.actorType!=='agent'&&m.actorId===state.currentActorId;
 return !m.actorType&&m.author==='You'&&state.currentActorId==='local-owner';
}
function drawChatDisplay(){
 if(view!=='Room')return;
 const list=document.querySelector('.message-list');if(!list)return;
 const style=chatStyleOn(),highlighted=highlightedAgentIds();
 list.classList.toggle('chat-style',style);
 for(const row of list.querySelectorAll('[data-message-id]')){
  const m=state.messages.find(m=>m.id===row.dataset.messageId);if(!m)continue;
  row.classList.toggle('message-own',style&&isOwnMessage(m));
  const marked=m.actorType==='agent'&&highlighted.includes(m.actorId);
  row.classList.toggle('agent-highlighted',marked);
  if(marked&&!row.querySelector('.agent-highlight-mark')){
   const mark=document.createElement('span');mark.className='tag agent-highlight-mark';mark.textContent='★ Highlighted';mark.title='You highlighted this agent. Only you see this.';
   row.querySelector('.message-meta')?.append(mark);
  }
 }
}
function chatDisplayAgents(){
 const seen=new Set(),agents=[];
 for(const p of state.participants||[]){const id=agentActorId(p);if(!seen.has(id)){seen.add(id);agents.push({id,name:p.name})}}
 return agents;
}
// Inputs carry no name attribute so an enclosing form never submits them.
function chatDisplayFields(){
 const highlighted=highlightedAgentIds(),agents=chatDisplayAgents();
 return '<fieldset class="chat-display-settings"><legend>Your chat display</legend>'
  +'<label class="field"><input type="checkbox" id="chat-style-toggle" '+(chatStyleOn()?'checked':'')+'> Chat style display</label>'
  +'<p class="muted">Shows your own messages on the right, like a text conversation. This is your personal choice and applies in every room. It is saved as soon as you change it.</p>'
  +(agents.length?'<p><strong>Highlighted agents</strong></p>'+agents.map(a=>'<label class="field"><input type="checkbox" data-highlight-agent="'+esc(a.id)+'" '+(highlighted.includes(a.id)?'checked':'')+'> '+esc(a.name)+'</label>').join('')
   +'<p class="muted">A highlighted agent’s messages are tinted and marked ★ for you only, in every room. Highlighting changes nothing about what the agent may do.</p>':'')
  +'</fieldset>';
}
function saveChatDisplay(input,kind,payload){
 input.disabled=true;
 return commit(kind,payload).catch(e=>{input.checked=!input.checked;note(e.message)}).finally(()=>{input.disabled=false});
}
function wireChatDisplayFields(root){
 const toggle=root.querySelector('#chat-style-toggle');
 if(toggle)toggle.onchange=()=>saveChatDisplay(toggle,'actor.chat_style.updated',{enabled:toggle.checked});
 root.querySelectorAll('[data-highlight-agent]').forEach(box=>box.onchange=()=>saveChatDisplay(box,'actor.agent_highlight.updated',{agentId:box.dataset.highlightAgent,enabled:box.checked}));
}
function addChatDisplayToModal(){
 const actions=$('#modal .actions');if(!actions||!chatDisplayEnabled()||!chatDisplaySelf())return;
 actions.insertAdjacentHTML('beforebegin',chatDisplayFields());wireChatDisplayFields($('#modal'));
}
function openChatDisplaySettings(){
 form('Display settings','',async()=>{});
 $('#modal .actions .primary')?.remove();$('#cancel').textContent='Close';
 addChatDisplayToModal();
}
function wireChatDisplayEntry(){
 if(view!=='Room'||!chatDisplayEnabled()||!chatDisplaySelf())return;
 const settings=$('#room-settings');
 if(settings){
  if(settings.dataset.chatDisplay)return;
  settings.dataset.chatDisplay='1';const open=settings.onclick;
  settings.onclick=e=>{open.call(settings,e);addChatDisplayToModal()};
  return;
 }
 // People without Room settings still need to reach their own display choices.
 const actions=$('.toolbar .actions');if(!actions||$('#chat-display-settings'))return;
 actions.insertAdjacentHTML('beforeend','<button id="chat-display-settings">Display settings</button>');
 $('#chat-display-settings').onclick=openChatDisplaySettings;
}
function addAgentHighlightChoice(p){
 const actions=$('#modal .actions');if(!actions||!chatDisplayEnabled()||!chatDisplaySelf())return;
 const id=agentActorId(p);
 actions.insertAdjacentHTML('beforebegin','<fieldset class="chat-display-settings"><legend>Your chat display</legend><label class="field"><input type="checkbox" data-highlight-agent="'+esc(id)+'" '+(highlightedAgentIds().includes(id)?'checked':'')+'> Highlight this agent’s messages for me</label><p class="muted">Tints and marks this agent’s messages for you only, in every room. Saved as soon as you change it; it does not change what the agent may do.</p></fieldset>');
 wireChatDisplayFields($('#modal'));
}
if(typeof openAgentModelOverlay==='function'){
 const chatDisplayModelOverlay=openAgentModelOverlay;
 openAgentModelOverlay=function(p){chatDisplayModelOverlay(p);addAgentHighlightChoice(p)};
}
function drawChatDisplayFlag(){
 if(view!=='Settings'||state.currentActorId!=='local-owner')return;
 const on=chatDisplayEnabled();
 const card=document.createElement('section');card.className='card';
 card.innerHTML='<h3>Cosmetic · Chat display choices</h3><p>Let each person choose chat style display and highlight agents for themselves. Turning this off shows every room in the standard layout and keeps saved choices for later.</p><button></button>';
 const button=card.querySelector('button');button.dataset.featureFlag='chatDisplayChoices';button.textContent='Chat display choices · '+(on?'On':'Off');
 button.onclick=()=>commit('feature.flag.updated',{name:'chatDisplayChoices',enabled:!on}).catch(e=>note(e.message));
 document.querySelector('#main').append(card);
}
const chatDisplayRender=render;
render=function(){chatDisplayRender();drawChatDisplay();wireChatDisplayEntry();drawChatDisplayFlag()};
