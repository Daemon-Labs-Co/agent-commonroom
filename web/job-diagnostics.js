// Required execution proof is server enforced independently of this optional,
// free owner view. No raw commands, outputs or model reasoning are displayed.
let jobDiagnosticsToken=0,jobDiagnosticsSelection=null,jobDiagnosticsPending=false;
function jobDiagnosticMarkup(result){
 const labels={allocated:'Intent recorded',submitted:'Sent to worker',shutdown_confirmed:'Shutdown verified',shutdown_unconfirmed:'Shutdown unverified'};
 const statuses={starting:'Starting',running:'Running',finished:'Finished',failed:'Failed',cancelled:'Cancelled',shutdown_unconfirmed:'Shutdown unverified',not_observed:'Not observed',unverified:'Unverified'};
 const leases={held:'Held by a supervisor',free:'No supervisor holds the lease',unknown:'Unverified',legacy:'Legacy record — no fenced proof'};
 const budget=result.turnBudget;
 const budgetMarkup=budget?'<section class="job-diagnostic-card"><h3>Daily managed-turn budget</h3><p>'+esc(budget.used)+' / '+esc(budget.limit)+' allocated · '+esc(budget.remaining)+' remaining</p><p class="muted">Resets '+esc(new Date(budget.resetsAt).toLocaleString())+' (UTC day). Includes connection checks and failed or cancelled turns; this is not a dollar spend limit.</p></section>':'';
 return workspaceDisposalMarkup(result)+budgetMarkup+'<p>'+esc(result.detail)+'</p><p class="muted">Snapshot checked: '+esc(new Date(result.checkedAt).toLocaleString())+' · Shared worker containers: '+esc(result.managedContainers||0)+' · Unresolved worker jobs: '+esc(result.unresolvedWorkerJobs||0)+'</p>'+(result.untrackedWorkerJobs?'<p class="muted">'+esc(result.untrackedWorkerJobs)+' historical or untracked worker records are outside this controller journal. They are not automatically adopted or replayed.</p>':'')+(result.untrackedUnresolvedJobs?'<p class="attention-warning">'+esc(result.untrackedUnresolvedJobs)+' untracked non-terminal jobs require administrator review before resuming.</p>':'')+(result.workerSnapshotTruncated?'<p class="attention-warning">Worker snapshot is incomplete; reconciliation is unavailable.</p>':'')+((result.jobs||[]).length?(result.jobs||[]).map(job=>'<article class="job-diagnostic-card"><h3>'+esc(job.provider)+' · '+esc(labels[job.status]||'Unverified')+'</h3><dl class="agent-diagnosis"><dt>Room identity</dt><dd>'+esc(job.roomId)+'</dd><dt>Job identity</dt><dd><code>'+esc(job.jobId)+'</code></dd><dt>Worker reports</dt><dd>'+esc(statuses[job.workerStatus]||'Unverified')+'</dd><dt>Supervisor ownership</dt><dd>'+esc(leases[job.leaseState]||'Unverified')+'</dd><dt>Last recorded transition</dt><dd>'+esc(job.workerUpdatedAt?new Date(job.workerUpdatedAt).toLocaleString():'Not verified')+'</dd><dt>Model deadline</dt><dd>'+esc(new Date(job.deadline).toLocaleString())+'</dd><dt>Controller update</dt><dd>'+esc(new Date(job.updatedAt).toLocaleString())+'</dd></dl><p class="muted">'+(job.shutdownConfirmed?'Shutdown is recorded as verified. No task was restarted.':'A deadline or missing container does not prove shutdown. Cancellation and reconciliation require a fresh fenced worker check.')+'</p>'+(!job.shutdownConfirmed?'<div class="actions"><button type="button" class="danger" data-job-cancel="'+esc(job.jobId)+'">Cancel and verify shutdown</button>'+(job.canReconcile?'<button type="button" data-job-reconcile="'+esc(job.jobId)+'">Reconcile abandoned job</button>':'')+'</div>':'')+(state?.currentActorId==='local-owner'&&result.workspaceDisposalEnabled===true&&job.version===4&&!job.shutdownConfirmed?'<button type="button" class="danger" data-workspace-propose="'+esc(job.jobId)+'">Review workspace disposal</button>':'')+'</article>').join(''):'<p>No durable jobs recorded for this agent. Older turns may predate job tracking.</p>');
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
  bindWorkspaceDisposalButtons(target);
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

function workspaceDisposalMarkup(result){
 if(state?.currentActorId!=='local-owner')return '';
 return (result.workspaceDisposals||[]).filter(p=>!jobDiagnosticsSelection||p.job.participantId===jobDiagnosticsSelection).map(p=>{
  const pending=p.status==='pending',now=new Date(result.checkedAt).getTime(),ready=now>=new Date(p.reviewReadyAt||p.notBefore).getTime()&&now<new Date(p.expiresAt).getTime();
  return '<section class="job-diagnostic-card"><h3>Workspace disposal · '+esc(p.status)+'</h3><p class="attention-warning">Permanently deletes all contents of this workspace, including unfinished edits. History, credentials and separate artifacts are outside this action. No storage credit is released by this action.</p><dl class="agent-diagnosis"><dt>Job</dt><dd><code>'+esc(p.job.jobId)+'</code></dd><dt>Workspace</dt><dd>'+esc(p.workspaceName)+'</dd><dt>Creation identity</dt><dd>'+esc(p.workspaceNonce)+' · '+esc(p.workspaceCreatedAt)+'</dd><dt>Approval available after</dt><dd>'+esc(new Date(p.reviewReadyAt||p.notBefore).toLocaleString())+'</dd><dt>Expires</dt><dd>'+esc(new Date(p.expiresAt).toLocaleString())+'</dd></dl>'+(pending?'<p>Refresh status after the review minute. Approval is never automatic.</p><div class="actions">'+(ready&&result.workspaceDisposalEnabled?'<button type="button" class="danger" data-workspace-approve="'+esc(p.id)+'" data-workspace-job="'+esc(p.job.jobId)+'">Approve permanent workspace deletion</button>':'')+'<button type="button" data-workspace-revoke="'+esc(p.id)+'">Revoke proposal</button></div>':'')+(['claimed','unconfirmed'].includes(p.status)?'<p class="attention-warning">Outcome requires reconciliation. Do not repeat the deletion.</p>':'')+'</section>';
 }).join('');
}
function bindWorkspaceDisposalButtons(target){
 for(const button of target.querySelectorAll('[data-workspace-propose],[data-workspace-approve],[data-workspace-revoke]')){
  button.onclick=()=>{
   jobDiagnosticsSelection=null;jobDiagnosticsToken++;
   if(button.dataset.workspacePropose){
    const jobId=button.dataset.workspacePropose;
    form('Propose workspace disposal','<p>This only records an intention. It does not delete data. The worker must already have stopped this job and retained its workspace. You will have at least one minute to review the exact volume before a separate approval.</p>',async()=>{await api('workspace-disposal',{action:'propose',jobId});note('Proposal recorded. Open Worker jobs and refresh after the review minute.');});
   }else if(button.dataset.workspaceApprove){
    const proposalId=button.dataset.workspaceApprove,jobId=button.dataset.workspaceJob;
    form('Approve permanent workspace deletion','<p class="attention-warning">All workspace contents, including unfinished edits, will be permanently deleted. There is no automatic retry or storage refund. Type the exact job identity to confirm: <code>'+esc(jobId)+'</code></p><label>Job identity<input name="confirm" required autocomplete="off"></label>',async data=>{await api('workspace-disposal',{action:'approve',proposalId,confirm:data.get('confirm')});await refresh();note('Workspace removal verified and audited. Storage accounting is unchanged; agents have not been resumed.');});
   }else{
    const proposalId=button.dataset.workspaceRevoke;
    form('Revoke disposal proposal','<p>Cancel this pending intention without deleting data.</p>',async()=>{await api('workspace-disposal',{action:'revoke',proposalId});note('Proposal revoked. Workspace retained.');});
   }
  };
 }
}
