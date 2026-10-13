// Optional accessible status help. Only safe server-projected summaries are used.
let agentStatusPopup=null,agentStatusAnchor=null;
function agentStatusText(details){return ['Doing: '+details.activity,'Why: '+details.reason,'Next: '+details.next].join('\n')}
function clearAgentStatusPopup(){if(agentStatusPopup)agentStatusPopup.remove();if(agentStatusAnchor)agentStatusAnchor.removeAttribute('aria-describedby');agentStatusPopup=null;agentStatusAnchor=null}
function showAgentStatusPopup(label,details){
 clearAgentStatusPopup();const popup=document.createElement('div');popup.id='agent-status-tooltip';popup.className='agent-status-tooltip';popup.setAttribute('role','tooltip');popup.textContent=agentStatusText(details);document.body.append(popup);
 agentStatusPopup=popup;agentStatusAnchor=label;label.setAttribute('aria-describedby',popup.id);
 const rect=label.getBoundingClientRect(),box=popup.getBoundingClientRect();popup.style.left=Math.max(8,Math.min(rect.left,window.innerWidth-box.width-8))+'px';popup.style.top=Math.max(8,rect.bottom+8+box.height<=window.innerHeight?rect.bottom+8:rect.top-box.height-8)+'px';
}
// Recovery uses existing authorized APIs. Opening this panel never invokes a
// provider, clears a failure or retries a potentially mutating model turn.
function openAgentDiagnosis(id){
 const p=state.participants.find(p=>p.id===id);
 if(!p?.statusDetails||state.featureFlags?.agentStatusDetails===false)return;
 clearAgentStatusPopup();const modal=$('#modal'),d=p.statusDetails;
 const c=state.connections.find(c=>c.id===p.connectionId),owner=state.currentActorId==='local-owner';
 modal.classList.remove('evidence-drawer');
 modal.innerHTML='<button type="button" class="modal-close" aria-label="Close dialog">×</button><button type="button" class="context-help overlay-help help-icon" aria-label="Help with agent recovery">?</button><h2>'+esc(p.name)+' · Diagnosis</h2><dl class="agent-diagnosis"><dt>Activity</dt><dd>'+esc(d.activity)+'</dd><dt>What happened</dt><dd>'+esc(d.reason)+'</dd><dt>What to do next</dt><dd>'+esc(d.next)+'</dd></dl><p class="muted">A failed turn may have changed files. Inspect partial work before sending another request. Opening this panel does not retry work or clear the failure.</p><div class="actions agent-recovery-actions">'+(owner&&c&&state.featureFlags?.connectionSignIn!==false?'<button type="button" id="agent-check-auth">Check sign-in</button>':'')+(owner&&c&&state.featureFlags?.connectionTesting!==false?'<button type="button" id="agent-test-connection">Test connection</button>':'')+(owner?'<button type="button" id="agent-view-history">View History</button>':'')+'<button type="button" id="agent-compose-request">Compose smaller request</button></div><p id="agent-recovery-result" role="status" aria-live="polite"></p>';
 modal.classList.add('has-overlay-help');modal.querySelector('.modal-close').onclick=()=>modal.close();
 modal.querySelector('.context-help').onclick=()=>openDocumentation('concepts');
 const check=modal.querySelector('#agent-check-auth');if(check)check.onclick=async()=>{
  check.disabled=true;const output=modal.querySelector('#agent-recovery-result');
  try{const result=await api('provider-auth',{action:'status',connectionId:c.id});
   if(typeof result.credentialPresent!=='boolean')throw Error('Could not determine sign-in status');
   output.textContent=result.credentialPresent?'Credentials are saved. This does not prove they are valid; Test connection verifies a provider reply.':'This connection needs sign-in. Open Connections and sign in, then test it.';
  }catch(e){output.textContent=e.message}finally{check.disabled=false}
 };
 const test=modal.querySelector('#agent-test-connection');if(test)test.onclick=()=>{modal.close();testConnection(c)};
 const history=modal.querySelector('#agent-view-history');if(history)history.onclick=()=>{modal.close();view='History & retention';render()};
 const compose=modal.querySelector('#agent-compose-request');compose.disabled=!!(state.incident||state.agentsStopped||state.agentRoomStopped||p.status==='Working');
 compose.onclick=()=>{const current=state.participants.find(member=>member.id===id);if(!current||state.incident||state.agentsStopped||state.agentRoomStopped||current.status==='Working'){note('Resolve the stop or wait for the active turn first');return}modal.close();recipient=id;const select=$('#send [name="recipient"]');if(select)select.value=id;$('#send [name="text"]')?.focus()};
 modal.showModal();
}
function drawAgentStatusDetails(){
 clearAgentStatusPopup();if(view!=='Room'||state?.featureFlags?.agentStatusDetails===false)return;
 for(const label of document.querySelectorAll('[data-agent-status]')){
  const participant=state.participants.find(p=>p.id===label.dataset.agentStatus);const details=participant?.statusDetails;if(!details)continue;
  label.tabIndex=0;label.setAttribute('aria-label',(participant.status||'Not connected')+'. '+agentStatusText(details));label.classList.add('agent-status-help');
  if(details.longRunning)label.textContent='Working · taking longer';
  label.onmouseenter=()=>showAgentStatusPopup(label,details);label.onmouseleave=clearAgentStatusPopup;label.onfocus=()=>showAgentStatusPopup(label,details);label.onblur=clearAgentStatusPopup;label.onclick=()=>openAgentDiagnosis(participant.id);label.setAttribute('role','button');label.setAttribute('aria-haspopup','dialog');label.onkeydown=e=>{if(e.key==='Enter'||e.key===' '){e.preventDefault();openAgentDiagnosis(participant.id)}};
 }
}
window.addEventListener('keydown',e=>{if(e.key==='Escape')clearAgentStatusPopup()});
window.addEventListener('scroll',clearAgentStatusPopup,true);
const agentStatusRender=render;render=function(){agentStatusRender();drawAgentStatusDetails()};
