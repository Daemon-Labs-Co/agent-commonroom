// Required execution proof is server enforced independently of this optional,
// free owner view. No raw commands, outputs or model reasoning are displayed.
let jobDiagnosticsToken=0,jobDiagnosticsSelection=null,jobDiagnosticsPending=false;
function jobDiagnosticMarkup(result){
 const labels={allocated:'Intent recorded',submitted:'Sent to worker',shutdown_confirmed:'Shutdown verified',shutdown_unconfirmed:'Shutdown unverified'};
 const statuses={starting:'Starting',running:'Running',finished:'Finished',failed:'Failed',cancelled:'Cancelled',shutdown_unconfirmed:'Shutdown unverified',not_observed:'Not observed',unverified:'Unverified'};
 const leases={held:'Held by a supervisor',free:'No supervisor holds the lease',unknown:'Unverified',legacy:'Legacy record — no fenced proof'};
 return '<p>'+esc(result.detail)+'</p><p class="muted">Snapshot checked: '+esc(new Date(result.checkedAt).toLocaleString())+' · Shared worker containers: '+esc(result.managedContainers||0)+' · Unresolved worker jobs: '+esc(result.unresolvedWorkerJobs||0)+'</p>'+(result.untrackedWorkerJobs?'<p class="muted">'+esc(result.untrackedWorkerJobs)+' historical or untracked worker records are outside this controller journal. They are not automatically adopted or replayed.</p>':'')+(result.untrackedUnresolvedJobs?'<p class="attention-warning">'+esc(result.untrackedUnresolvedJobs)+' untracked non-terminal jobs require administrator review before resuming.</p>':'')+(result.workerSnapshotTruncated?'<p class="attention-warning">Worker snapshot is incomplete; reconciliation is unavailable.</p>':'')+((result.jobs||[]).length?(result.jobs||[]).map(job=>'<article class="job-diagnostic-card"><h3>'+esc(job.provider)+' · '+esc(labels[job.status]||'Unverified')+'</h3><dl class="agent-diagnosis"><dt>Room identity</dt><dd>'+esc(job.roomId)+'</dd><dt>Job identity</dt><dd><code>'+esc(job.jobId)+'</code></dd><dt>Worker reports</dt><dd>'+esc(statuses[job.workerStatus]||'Unverified')+'</dd><dt>Supervisor ownership</dt><dd>'+esc(leases[job.leaseState]||'Unverified')+'</dd><dt>Last recorded transition</dt><dd>'+esc(job.workerUpdatedAt?new Date(job.workerUpdatedAt).toLocaleString():'Not verified')+'</dd><dt>Model deadline</dt><dd>'+esc(new Date(job.deadline).toLocaleString())+'</dd><dt>Controller update</dt><dd>'+esc(new Date(job.updatedAt).toLocaleString())+'</dd></dl><p class="muted">'+(job.shutdownConfirmed?'Shutdown is recorded as verified. No task was restarted.':'A deadline or missing container does not prove shutdown. Cancellation and reconciliation require a fresh fenced worker check.')+'</p>'+(!job.shutdownConfirmed?'<div class="actions"><button type="button" class="danger" data-job-cancel="'+esc(job.jobId)+'">Cancel and verify shutdown</button>'+(job.canReconcile?'<button type="button" data-job-reconcile="'+esc(job.jobId)+'">Reconcile abandoned job</button>':'')+'</div>':'')+'</article>').join(''):'<p>No durable jobs recorded for this agent. Older turns may predate job tracking.</p>');
}
async function refreshJobDiagnostics(token){
 if(jobDiagnosticsPending||token!==jobDiagnosticsToken||jobDiagnosticsSelection===null)return;
 jobDiagnosticsPending=true;
 const selected=jobDiagnosticsSelection;
 try{
  const result=await api('job-diagnostics?scope=application'+(selected?'&participantId='+encodeURIComponent(selected):''));
  if(token!==jobDiagnosticsToken||jobDiagnosticsSelection!==selected||!$('#modal').open)return;
  const target=$('#job-diagnostics-content');if(!target)return;
  const previous=target.scrollTop;target.innerHTML=jobDiagnosticMarkup(result);target.scrollTop=previous;
  for(const button of target.querySelectorAll('[data-job-cancel],[data-job-reconcile]')){
   button.onclick=()=>{
    const action=button.dataset.jobCancel?'cancel':'reconcile',jobId=button.dataset.jobCancel||button.dataset.jobReconcile;
    jobDiagnosticsSelection=null;jobDiagnosticsToken++;
    form(action==='cancel'?'Cancel worker job':'Reconcile abandoned job','<p>Verify supervisor and container shutdown. This does not undo file changes, resume agents or replay a task.</p><label>Reason<textarea name="reason" required maxlength="500"></textarea></label>',async data=>{const result=await api('job-recovery?scope=application',{action,jobId,reason:data.get('reason')});await refresh();note(result.detail||'Shutdown verified; no task restarted.');});
   };
  }
 }catch(error){if(token===jobDiagnosticsToken&&$('#job-diagnostics-content'))$('#job-diagnostics-content').textContent=error.message}
 finally{jobDiagnosticsPending=false;if(token!==jobDiagnosticsToken&&jobDiagnosticsSelection!==null&&$('#modal').open)refreshJobDiagnostics(jobDiagnosticsToken)}
}
function openJobDiagnostics(id=''){
 if(view!=='Room'||state?.currentActorId!=='local-owner'||state.featureFlags?.jobDiagnostics===false)return;
 const modal=$('#modal');if(typeof closeSquadMenu==='function')closeSquadMenu();
 jobDiagnosticsSelection=id;const token=++jobDiagnosticsToken;
 modal.classList.remove('evidence-drawer');modal.classList.add('has-overlay-help');
 modal.innerHTML='<button type="button" class="modal-close" aria-label="Close dialog">×</button><button type="button" class="context-help overlay-help help-icon" aria-label="Help with worker job diagnostics">?</button><h2>Worker job diagnostics</h2><div id="job-diagnostics-content" aria-live="polite"><p>Checking durable records and the approved worker…</p></div><div class="actions"><button type="button" id="job-diagnostics-refresh">Refresh status</button></div>';
 modal.querySelector('.modal-close').onclick=()=>{jobDiagnosticsSelection=null;jobDiagnosticsToken++;modal.close()};modal.querySelector('.context-help').onclick=()=>openDocumentation('worker-jobs');modal.querySelector('#job-diagnostics-refresh').onclick=()=>refreshJobDiagnostics(token);modal.addEventListener('close',()=>{if(token===jobDiagnosticsToken){jobDiagnosticsSelection=null;jobDiagnosticsToken++}},{once:true});modal.showModal();refreshJobDiagnostics(token);
}
function drawJobDiagnostics(){
 const enabled=state?.currentActorId==='local-owner'&&state.featureFlags?.jobDiagnostics!==false&&view==='Room';
 if(!enabled){if(jobDiagnosticsSelection!==null){jobDiagnosticsSelection=null;jobDiagnosticsToken++;$('#modal').close()}return}
 for(const label of document.querySelectorAll('[data-agent-status]')){
  const p=state.participants.find(p=>p.id===label.dataset.agentStatus);const panel=label.closest('.participant')?.querySelector('.menu-panel');if(!p||!panel||panel.querySelector('[data-job-diagnostics]'))continue;
  const button=document.createElement('button');button.type='button';button.textContent='Job diagnostics';button.setAttribute('data-job-diagnostics',p.id);button.setAttribute('aria-label','Job diagnostics for '+p.name);button.onclick=()=>{panel.closest('details')?.removeAttribute('open');openJobDiagnostics(p.id)};panel.insertBefore(button,panel.querySelector('[data-remove]'));
 }
 const toolbar=$('#main .toolbar');if(toolbar&&!toolbar.querySelector('[data-room-job-diagnostics]')){const button=document.createElement('button');button.type='button';button.textContent='Worker jobs';button.setAttribute('data-room-job-diagnostics','true');button.onclick=()=>openJobDiagnostics();toolbar.append(button)}
}
const jobDiagnosticsRender=render;render=function(){jobDiagnosticsRender();drawJobDiagnostics()};
