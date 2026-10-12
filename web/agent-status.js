// Optional accessible status help. Only safe server-projected summaries are used.
let agentStatusPopup=null,agentStatusAnchor=null;
function agentStatusText(details){return ['Doing: '+details.activity,'Why: '+details.reason,'Next: '+details.next].join('\n')}
function clearAgentStatusPopup(){if(agentStatusPopup)agentStatusPopup.remove();if(agentStatusAnchor)agentStatusAnchor.removeAttribute('aria-describedby');agentStatusPopup=null;agentStatusAnchor=null}
function showAgentStatusPopup(label,details){
 clearAgentStatusPopup();const popup=document.createElement('div');popup.id='agent-status-tooltip';popup.className='agent-status-tooltip';popup.setAttribute('role','tooltip');popup.textContent=agentStatusText(details);document.body.append(popup);
 agentStatusPopup=popup;agentStatusAnchor=label;label.setAttribute('aria-describedby',popup.id);
 const rect=label.getBoundingClientRect(),box=popup.getBoundingClientRect();popup.style.left=Math.max(8,Math.min(rect.left,window.innerWidth-box.width-8))+'px';popup.style.top=Math.max(8,rect.bottom+8+box.height<=window.innerHeight?rect.bottom+8:rect.top-box.height-8)+'px';
}
function drawAgentStatusDetails(){
 clearAgentStatusPopup();if(view!=='Room'||state?.featureFlags?.agentStatusDetails===false)return;
 for(const label of document.querySelectorAll('[data-agent-status]')){
  const participant=state.participants.find(p=>p.id===label.dataset.agentStatus);const details=participant?.statusDetails;if(!details)continue;
  label.tabIndex=0;label.setAttribute('aria-label',(participant.status||'Not connected')+'. '+agentStatusText(details));label.classList.add('agent-status-help');
  label.onmouseenter=()=>showAgentStatusPopup(label,details);label.onmouseleave=clearAgentStatusPopup;label.onfocus=()=>showAgentStatusPopup(label,details);label.onblur=clearAgentStatusPopup;label.onclick=()=>showAgentStatusPopup(label,details);
 }
}
window.addEventListener('keydown',e=>{if(e.key==='Escape')clearAgentStatusPopup()});
window.addEventListener('scroll',clearAgentStatusPopup,true);
const agentStatusRender=render;render=function(){agentStatusRender();drawAgentStatusDetails()};
