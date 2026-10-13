// Recovery checks the worker again at action time. It does not replay old work.
function openSessionRecovery(id){
 const p=state.participants.find(p=>p.id===id);
 if(!p||state.currentActorId!=='local-owner'||state.featureFlags?.agentRecovery===false)return;
 const modal=$('#modal');modal.classList.remove('evidence-drawer');
 modal.innerHTML='<button type="button" class="modal-close" aria-label="Close dialog">×</button><button type="button" class="context-help overlay-help help-icon" aria-label="Help with session recovery">?</button><h2>'+esc(p.name)+' · Session recovery</h2><p>Check the shared worker before recovery. Other active jobs prevent a reset; an unavailable worker is not assumed stopped.</p><p class="muted">Resume keeps the saved provider session. Start fresh clears that binding, while preserving room history, profile, credentials and workspace files. Neither option undoes partial file changes or automatically replays the previous task.</p><div class="actions agent-recovery-actions"><button type="button" id="recovery-check">Check worker</button><button type="button" id="recovery-resume" disabled>Resume session</button><button type="button" id="recovery-fresh" disabled>Start fresh session</button></div><p id="recovery-status" role="status" aria-live="polite">Check the worker to enable recovery.</p>';
 modal.classList.add('has-overlay-help');modal.querySelector('.modal-close').onclick=()=>modal.close();modal.querySelector('.context-help').onclick=()=>openDocumentation('concepts');
 const check=modal.querySelector('#recovery-check'),resume=modal.querySelector('#recovery-resume'),fresh=modal.querySelector('#recovery-fresh'),status=modal.querySelector('#recovery-status');
 const guard=!!(state.incident||state.agentsStopped||state.agentRoomStopped);check.disabled=guard;if(guard)status.textContent='Resolve the integrity incident or emergency stop first.';
 check.onclick=async()=>{check.disabled=true;resume.disabled=true;fresh.disabled=true;try{const result=await api('agent-recovery',{action:'check',participantId:id});status.textContent=result.detail;resume.disabled=!result.idle||!p.threadId;fresh.disabled=!result.idle}catch(e){status.textContent=e.message}finally{check.disabled=false}};
 const recover=async action=>{
  check.disabled=true;resume.disabled=true;fresh.disabled=true;
  try{const result=await api('agent-recovery',{action,participantId:id});status.textContent=result.detail;if(!result.recovered)return;
   const draft=$('#send [name="text"]')?.value;modal.close();await refresh();recipient=id;const select=$('#send [name="recipient"]');if(select)select.value=id;const input=$('#send [name="text"]');if(input){if(draft!==undefined)input.value=draft;input.focus()}note('Session ready. Inspect partial work, then send a new request.');
  }catch(e){status.textContent=e.message}finally{check.disabled=false}
 };
 resume.onclick=()=>recover('resume');fresh.onclick=()=>recover('fresh');modal.showModal();
}
function drawSessionRecovery(){
 if(view!=='Room'||state?.currentActorId!=='local-owner'||state.featureFlags?.agentRecovery===false)return;
 for(const label of document.querySelectorAll('[data-agent-status]')){
  const p=state.participants.find(p=>p.id===label.dataset.agentStatus);if(!p||p.status==='Working')continue;
  const panel=label.closest('.participant')?.querySelector('.menu-panel');if(!panel||panel.querySelector('[data-session-recovery]'))continue;
  const button=document.createElement('button');button.type='button';button.textContent='Recover session';button.setAttribute('data-session-recovery',p.id);button.setAttribute('aria-label','Recover session for '+p.name);button.onclick=()=>{if(typeof closeSquadMenu==='function')closeSquadMenu();panel.closest('details')?.removeAttribute('open');openSessionRecovery(p.id)};panel.insertBefore(button,panel.querySelector('[data-remove]'));
 }
}
const recoveryRender=render;render=function(){recoveryRender();drawSessionRecovery()};
