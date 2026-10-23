// Human capture/review only. Model content never authorizes task execution.
let proposalViewGeneration=0;
function drawTaskProposalControls(){
 const generation=++proposalViewGeneration;
 if(state?.currentActorId!=='local-owner')return;
 const target=$('#main');if(!target)return;
 const enabled=state.featureFlags?.taskProposals===true;
 if(view==='Settings'){
  target.insertAdjacentHTML('beforeend','<section class="card"><h3>Task proposals</h3><p>Capture a conversation suggestion for explicit human review. Approval creates a draft and never starts work.</p><button id="configure-task-proposals" data-feature-flag="taskProposals">Task proposals · '+(enabled?'On':'Off')+'</button></section>');
  $('#configure-task-proposals').onclick=()=>form('Task proposals','<label><input type="checkbox" name="enabled" '+(enabled?'checked':'')+'> Enable proposal capture and review</label><p>Off preserves proposals, tasks and history. Existing independent task permissions remain unchanged.</p>',async d=>commit('feature.flag.updated',{name:'taskProposals',enabled:d.has('enabled')}));
 }
 if(view!=='Tasks')return;
 const rid=roomId;
 target.insertAdjacentHTML('beforeend','<section class="card"><h2>Conversation proposals</h2><p>Suggestions are not approvals. Review the original source before creating a draft.</p>'+(enabled?'<button id="capture-task-proposal">Capture a suggestion</button>':'<p>Capture and review are disabled. Retained proposals remain visible.</p>')+'<div id="task-proposal-list">Loading proposals…</div></section>');
 if(enabled)$('#capture-task-proposal').onclick=()=>{
  const options=state.messages.filter(m=>m.id&&m.roomId===rid).map(m=>'<option value="'+esc(m.id)+'">'+esc(m.author||'Unknown author')+' · '+esc(m.text.slice(0,100))+'</option>').join('');
  form('Capture a task proposal','<p>This records your interpretation of the selected source. It does not approve or run work.</p><label>Source message<select name="sourceMessageId" required>'+options+'</select></label>'+field('Suggested title','title')+'<label>Proposed acceptance criteria<textarea name="criteria"></textarea></label>',async d=>{await api('task-proposals?room='+encodeURIComponent(rid),{action:'capture',sourceMessageId:d.get('sourceMessageId'),title:d.get('title'),criteria:d.get('criteria')||''});await refresh()});
 };
 api('task-proposals?room='+encodeURIComponent(rid)).then(result=>{
  if(generation!==proposalViewGeneration||view!=='Tasks'||roomId!==rid)return;
  const list=$('#task-proposal-list');if(!list)return;
  list.innerHTML=result.proposals.map(entry=>{
   const p=entry.proposal;
   return '<article class="card"><h3>'+esc(p.title)+'</h3><p>'+esc(p.status)+' · captured by '+esc(p.capturedBy)+'</p><p>'+esc(p.criteria||'Acceptance criteria need review.')+'</p>'+proposalSourceMarkup(entry)+(p.taskId?'<p>Created task: '+esc(p.taskId)+'</p>':'')+(result.enabled&&p.status==='Proposed'&&entry.sourceAvailable?'<button data-approve-proposal="'+esc(p.id)+'">Review and create draft</button><button data-reject-proposal="'+esc(p.id)+'">Reject proposal</button>':'')+'</article>';
  }).join('')||'<p>No proposals captured.</p>';
  document.querySelectorAll('[data-approve-proposal]').forEach(b=>b.onclick=()=>{
   const entry=result.proposals.find(e=>e.proposal.id===b.dataset.approveProposal),p=entry.proposal;
   form('Review and create draft',proposalSourceMarkup(entry)+field('Reviewed title','title',p.title)+'<label>Reviewed acceptance criteria<textarea name="criteria" required>'+esc(p.criteria)+'</textarea></label><p>Approval creates an unassigned Draft task. Assigning, dispatching, observing and accepting work are separate decisions.</p>',async d=>{await api('task-proposals?room='+encodeURIComponent(rid),{action:'approve',proposalId:p.id,revision:p.revision,title:d.get('title'),criteria:d.get('criteria')});await refresh()});
  });
  document.querySelectorAll('[data-reject-proposal]').forEach(b=>b.onclick=()=>{
   const entry=result.proposals.find(e=>e.proposal.id===b.dataset.rejectProposal),p=entry.proposal;
   form('Reject proposal',proposalSourceMarkup(entry)+'<p>Keep the original suggestion and record your rejection. No task will be created.</p>',async()=>{await api('task-proposals?room='+encodeURIComponent(rid),{action:'reject',proposalId:p.id,revision:p.revision});await refresh()});
  });
 }).catch(e=>{if(generation===proposalViewGeneration&&view==='Tasks'&&roomId===rid){const list=$('#task-proposal-list');if(list)list.textContent='Proposals unavailable';note(e.message)}});
}
function proposalSourceMarkup(entry){const p=entry.proposal;return '<details><summary>Original source · '+esc(p.sourceAuthor||'Unknown author')+' · '+esc(p.sourceActorType||'unrecorded type')+'</summary><p>Message '+esc(p.sourceMessageId)+' · actor '+esc(p.sourceActorId||'unrecorded')+'</p><pre>'+esc(entry.sourceAvailable?entry.source.text:'Original source is unavailable; review is blocked.')+'</pre></details>'}
const proposalRender=render;render=function(){proposalRender();drawTaskProposalControls()};
