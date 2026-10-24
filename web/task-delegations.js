// Only human review controls. Agent identity is never taken from this browser.
let delegationViewGeneration=0;
function delegationReviewFields(d){
 const t=state.tasks.find(t=>t.id===d.taskId);
 const operations=['read','write','patch','command'].filter((op,i)=>(d.parentOperations&(1<<i))!==0).join(', ');
 return '<p>Task: '+esc(t?.title||d.taskId)+'</p><p>'+esc(t?.criteria||'Task unavailable')+'</p><p>Proposer: '+esc(d.sourceParticipantId)+' → delegate: '+esc(d.delegateId)+'</p><p>Repository: '+esc(d.repositoryDigest)+'</p><p>Parent operation limit: '+esc(operations)+'</p><p>Deadline: '+esc(d.expiresAt)+'</p><p>Approval only reassigns this bounded task after verified parent shutdown and outcome. No provider launches or additional permissions are granted. Dispatch remains a separate human action.</p>';
}
function drawDelegationControls(){
 const generation=++delegationViewGeneration;
 if(state?.currentActorId!=='local-owner')return;
 const main=$('#main');if(!main)return;
 const enabled=state.featureFlags?.taskDelegation===true;
 if(view==='Settings'){
  main.insertAdjacentHTML('beforeend','<section class="card"><h3>Bounded task delegation</h3><p>Review assignment proposals from authenticated worker jobs. Credentials and execution permissions are never transferred.</p><button id="configure-task-delegation" data-feature-flag="taskDelegation">Task delegation · '+(enabled?'On':'Off')+'</button></section>');
  $('#configure-task-delegation').onclick=()=>form('Task delegation','<label><input type="checkbox" name="enabled" '+(enabled?'checked':'')+'> Enable bounded assignment proposals and review</label><p>Off invalidates old proposals and delegated work, preserves history, and leaves independent human work under its own permissions. Worker cleanup remains required.</p>',async d=>commit('feature.flag.updated',{name:'taskDelegation',enabled:d.has('enabled')}));
 }
 if(view!=='Tasks')return;
 const rid=roomId;
 main.insertAdjacentHTML('beforeend','<section class="card"><h2>Delegation proposals</h2><p>Only authenticated worker jobs can submit proposals. Current provider adapters do not yet automatically emit them.</p><div id="task-delegation-list">Loading…</div></section>');
 api('task-delegations?room='+encodeURIComponent(rid)).then(result=>{
  if(generation!==delegationViewGeneration||roomId!==rid||view!=='Tasks')return;
  const list=$('#task-delegation-list');if(!list)return;
  list.innerHTML=result.delegations.map(d=>'<article class="card">'+delegationReviewFields(d)+'<p>Status: '+esc(d.status)+'</p>'+(result.enabled&&d.status==='Proposed'?'<button data-review-delegation="'+esc(d.id)+'">Review reassignment</button><button data-reject-delegation="'+esc(d.id)+'">Reject</button>':'')+'</article>').join('')||'<p>No delegation proposals.</p>';
  for(const [selector,action,key] of [['[data-review-delegation]','approve','reviewDelegation'],['[data-reject-delegation]','reject','rejectDelegation']])document.querySelectorAll(selector).forEach(b=>b.onclick=()=>{
   const d=result.delegations.find(d=>d.id===b.dataset[key]);
   form(action==='approve'?'Approve bounded reassignment':'Reject delegation',delegationReviewFields(d),async()=>{await api('task-delegations?room='+encodeURIComponent(rid),{action,id:d.id,revision:d.revision});await refresh()});
  });
 }).catch(e=>{if(generation===delegationViewGeneration&&view==='Tasks'&&roomId===rid){const list=$('#task-delegation-list');if(list)list.textContent='Delegations unavailable';note(e.message)}});
}
const delegationRender=render;render=function(){delegationRender();drawDelegationControls()};
