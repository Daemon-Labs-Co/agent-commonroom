// Stable actor types outrank mutable display names. Legacy ambiguity is visible.
function messageIdentityLabel(message){return message.actorType==='human'?'Human':message.actorType==='agent'?'Agent':message.actorType==='system'?'System':'Legacy · identity not recorded'}
function drawMessageIdentity(){
 if(view!=='Room')return;
 for(const row of document.querySelectorAll('[data-message-id]')){
  const message=state.messages.find(m=>m.id===row.dataset.messageId);if(!message)continue;
  const badge=document.createElement('span');badge.className='tag message-identity';badge.textContent=messageIdentityLabel(message);row.querySelector('.message-meta')?.append(badge);
 }
}
const messageIdentityRender=render;render=function(){messageIdentityRender();drawMessageIdentity()};
