// Attention is acknowledgment only, never approval or permission to execute.
let lastAttentionKey=null;
function pendingHumanRequests(messages){return messages.filter(m=>m.actorType==='agent'&&m.humanRequest&&!m.attentionAcknowledgedBy)}
function focusAttentionMessage(id){
 const list=document.querySelector('.message-list');const row=[...document.querySelectorAll('[data-message-id]')].find(n=>n.dataset.messageId===id);if(!list||!row)return;
 list.scrollTop+=row.getBoundingClientRect().top-list.getBoundingClientRect().top-12;row.tabIndex=-1;row.focus({preventScroll:true});
}
async function acknowledgeHumanRequest(id){try{await api('events',{kind:'message.attention.acknowledged',payload:{messageId:id}});await refresh()}catch(e){note(e.message)}}
function drawHumanAttention(){
 if(state?.featureFlags?.humanAttention===false)return;
 for(const button of document.querySelectorAll('[data-room]')){
  const count=state.rooms.find(r=>r.id===button.dataset.room)?.pendingHumanRequests||0;if(!count)continue;
  const badge=document.createElement('span');badge.className='attention-room-count';badge.textContent=String(count);badge.setAttribute('aria-label',count+' pending human requests');const label=document.createElement('span');label.className='attention-room-label';for(const child of [...(button.childNodes||[])])label.append(child);button.classList.add('attention-room-button');button.append(label);button.append(badge);
 }
 if(view!=='Room')return;
 const pending=pendingHumanRequests(state.messages);
 for(const row of document.querySelectorAll('[data-message-id]')){
  const m=state.messages.find(m=>m.id===row.dataset.messageId);if(!m?.humanRequest||m.actorType!=='agent')continue;
  if(m.humanRequest==='explicit'&&typeof m.text==='string'){const text=row.querySelector('.message-content > p');if(text)text.textContent=m.text.replace(/^\s*\[HUMAN ACTION REQUIRED\]\s*/i,'')}
  const handled=!!m.attentionAcknowledgedBy;row.classList.add(handled?'attention-handled':'attention-required');
  const badge=document.createElement('span');badge.className='attention-message-label';badge.textContent=handled?'Acknowledged':m.humanRequest==='explicit'?'Needs your attention':'May need your attention';row.querySelector('.message-meta').append(badge);
  if(!handled){const b=document.createElement('button');b.className='attention-acknowledge';b.textContent='Acknowledge';b.onclick=()=>acknowledgeHumanRequest(m.id);row.querySelector('.message-content').append(b)}
 }
 const heading=document.querySelector('.conversation .panel-heading');if(!heading||!pending.length){lastAttentionKey=null;return}
 const banner=document.createElement('div');banner.className='attention-banner';const key=roomId+':'+pending.map(m=>m.id).join(',');banner.setAttribute('role','status');banner.setAttribute('aria-live',key===lastAttentionKey?'off':'polite');lastAttentionKey=key;
 const text=document.createElement('strong');text.textContent=pending.length+' message'+(pending.length===1?' needs':'s need')+' your attention';banner.append(text);
 const button=document.createElement('button');button.textContent='View request';button.onclick=()=>focusAttentionMessage(pending[0].id);banner.append(button);heading.after(banner);
}
const humanAttentionRender=render;render=function(){humanAttentionRender();drawHumanAttention()};
