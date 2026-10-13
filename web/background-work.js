// Operational telemetry only: private reasoning, raw commands and outputs are
// intentionally unavailable. This feature never extends a job's hard limit.
let backgroundParticipant=null;
function backgroundMarkup(p){
 const events=p.background||[];
 return '<p class="muted">Worker observations and reported tool activity. A running process does not prove useful progress. Claude currently reports worker lifecycle only. Command contents, outputs and private reasoning are omitted.</p>'+(events.length?'<ol class="background-timeline">'+events.map(e=>'<li><time>'+esc(new Date(e.observedAt).toLocaleTimeString())+'</time><strong>'+esc(e.label)+'</strong><small>'+esc(e.jobId?'Worker job '+e.jobId.slice(0,8):'Controller/provider observation')+'</small></li>').join('')+'</ol>':'<p>No activity has been recorded. Existing turns may predate monitoring.</p>')+'<p class="muted">Latest 64 retained observations. Older records remain in History. Updates stop when this controller is unavailable; restart reconciliation is not implemented.</p>';
}
function openBackgroundWork(id){
 const p=state.participants.find(p=>p.id===id);if(!p||state.featureFlags?.agentBackgroundWork===false)return;
 const modal=$('#modal');backgroundParticipant=id;modal.classList.remove('evidence-drawer');
 modal.innerHTML='<button type="button" class="modal-close" aria-label="Close dialog">×</button><button type="button" class="context-help overlay-help help-icon" aria-label="Help with background work">?</button><h2>'+esc(p.name)+' · Background work</h2><div id="background-work-content">'+backgroundMarkup(p)+'</div><div class="actions">'+(state.currentActorId==='local-owner'?'<button type="button" class="danger" id="background-stop-all">Stop all agents</button>':'')+'</div>';
 modal.classList.add('has-overlay-help');modal.querySelector('.modal-close').onclick=()=>{backgroundParticipant=null;modal.close()};modal.querySelector('.context-help').onclick=()=>openDocumentation('concepts');
 const stop=modal.querySelector('#background-stop-all');if(stop){stop.disabled=!!state.agentsStopped;if(state.agentsStopped)stop.textContent='Stop requested';stop.onclick=async()=>{stop.disabled=true;try{await api('agent-safety',{action:'stop',scope:'application',reason:'Stopped from background work'});await refresh();note('All new agent turns blocked; cancellation requested. Check worker shutdown observations.')}catch(e){note(e.message)}finally{stop.disabled=!!state?.agentsStopped}};}
 modal.showModal();
}
function drawBackgroundWork(){
 if(state?.featureFlags?.agentBackgroundWork===false){if(backgroundParticipant){backgroundParticipant=null;$('#modal').close()}return}
 if(view==='Room')for(const label of document.querySelectorAll('[data-agent-status]')){
  const p=state.participants.find(p=>p.id===label.dataset.agentStatus);if(!p)continue;
  const panel=label.closest('.participant')?.querySelector('.menu-panel');if(!panel||panel.querySelector('[data-background-work]'))continue;
  const button=document.createElement('button');button.type='button';button.textContent='Background work';button.setAttribute('data-background-work',p.id);button.setAttribute('aria-label','Background work for '+p.name);button.onclick=()=>{if(typeof closeSquadMenu==='function')closeSquadMenu();panel.closest('details')?.removeAttribute('open');openBackgroundWork(p.id)};panel.insertBefore(button,panel.querySelector('[data-session-recovery]')||panel.querySelector('[data-remove]'));
 }
 if(backgroundParticipant){const target=$('#background-work-content');const p=state.participants.find(p=>p.id===backgroundParticipant);if(!target){backgroundParticipant=null;return}if(!p||view!=='Room'){backgroundParticipant=null;$('#modal').close();return}
  const nearEnd=target.scrollHeight-target.clientHeight-target.scrollTop<24,previous=target.scrollTop;
  target.innerHTML=backgroundMarkup(p);target.scrollTop=nearEnd?target.scrollHeight:previous;
  const stop=$('#modal').querySelector('#background-stop-all');if(stop){stop.disabled=!!state.agentsStopped;stop.textContent=state.agentsStopped?'Stop requested':'Stop all agents'}
 }
}
const backgroundRender=render;render=function(){backgroundRender();drawBackgroundWork()};
