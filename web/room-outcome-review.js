// Acceptance is a human judgment about a saved snapshot, not an agent capability.
let roomReviewGeneration=0;
function drawRoomOutcomeReview(){
 const generation=++roomReviewGeneration,main=$('#main');if(!main)return;
 const owner=state.currentActorId==='local-owner';
 if(view==='Settings'&&owner){
  const enabled=state.featureFlags?.roomOutcomeReview===true;
  main.insertAdjacentHTML('beforeend','<section class="card"><h3>Workflows · Room outcome review</h3><p>Record human acceptance of a defined room deliverable after its tasks are accepted. Changed work requires another review.</p><button id="room-review-setting" data-feature-flag="roomOutcomeReview">Room outcome review · '+(enabled?'On':'Off')+'</button></section>');
  $('#room-review-setting').onclick=()=>form('Room outcome review','<label><input name="enabled" type="checkbox" '+(enabled?'checked':'')+'> Enable owner acceptance and reopening</label><p>Also requires Room outcomes. Off preserves review history and removes decision controls. Acceptance never authorizes execution or data deletion.</p>',d=>commit('feature.flag.updated',{name:'roomOutcomeReview',enabled:d.has('enabled')}));
 }
 if(view!=='Room'&&view!=='Tasks')return;
 const rid=roomId,room=state.rooms.find(r=>r.id===rid);if(!room||room.personalActorId)return;
 main.insertAdjacentHTML('beforeend','<section class="card"><h2>Deliverable review</h2><div id="room-review-status">Loading…</div></section>');
 api('room-outcome-review?room='+encodeURIComponent(rid)).then(result=>{
  if(generation!==roomReviewGeneration||roomId!==rid||(view!=='Room'&&view!=='Tasks'))return;
  const panel=$('#room-review-status');if(!panel)return;
  const review=result.review||{},definition=result.definition||{};
  const taskNames=new Map((result.tasks||[]).map(t=>[t.id,t.title]));
  const snapshotEvidence=(result.evidence||[]).map(e=>{
   const association=e.taskId?'Task: '+(taskNames.get(e.taskId)||'Not in current task snapshot')+' ('+e.taskId+')':'Room-level evidence';
   return '<article><strong>'+esc(e.title||e.id)+'</strong><p>'+esc(association)+'</p><p>Kind: '+esc(e.kind||'Unspecified')+'</p><p style="white-space:pre-wrap">'+esc(e.body||'')+'</p><p>Source: '+esc(e.source||'Unspecified')+'</p></article>';
  }).join('');
  const snapshotTasks=(result.tasks||[]).map(t=>'<article><strong>'+esc(t.title)+'</strong><p>Task ID: '+esc(t.id)+' · '+esc(t.status)+'</p><p style="white-space:pre-wrap">Acceptance criteria: '+esc(t.criteria||'Not defined')+'</p></article>').join('');
  panel.innerHTML='<p>Status: '+esc(result.status)+'</p>'+(review.note?'<p style="white-space:pre-wrap">Last human review: '+esc(review.note)+'</p><p>Reviewed by '+esc(review.reviewedBy)+' · '+esc(review.reviewedAt)+'</p>':'')+'<p>Requires a defined outcome and criteria, at least one task, and all room tasks accepted without pending dispatch. The owner must assess the evidence; task status alone does not prove delivery.</p>';
  if(!owner||!result.enabled)return;
  for(const action of ['accept','reopen']){
   if(action==='accept'&&(!result.ready||result.status==='Accepted')||action==='reopen'&&review.status!=='Accepted')continue;
   panel.insertAdjacentHTML('beforeend','<button id="room-review-'+action+'">'+(action==='accept'?'Review and accept deliverable':'Reopen deliverable')+'</button>');
   $('#room-review-'+action).onclick=()=>form(action==='accept'?'Accept '+esc(definition.name||'Room'):'Reopen '+esc(definition.name||'Room'),'<p>'+esc(definition.outcome||'No current definition')+'</p><p>'+esc(definition.acceptanceCriteria||'No criteria')+'</p><h3>Task snapshot</h3>'+snapshotTasks+'<h3>Evidence snapshot</h3>'+snapshotEvidence+'<label class="field">'+(action==='accept'?'Evidence reviewed and acceptance rationale':'Reason for reopening')+'<textarea name="note" maxlength="4000" required></textarea></label><p>This records your decision. It does not launch agents, verify artifacts, grant permissions or dispose of data.</p>',async d=>{
    if(roomId!==rid){note('Room changed; reopen the review in the intended room.');return}
    await api('room-outcome-review?room='+encodeURIComponent(rid),{action,revision:review.revision||0,snapshot:result.snapshot,note:d.get('note')});await refresh();
   });
  }
 }).catch(error=>{if(generation===roomReviewGeneration&&roomId===rid){const panel=$('#room-review-status');if(panel)panel.textContent=error.message}});
}
const roomOutcomeReviewRender=render;render=function(){roomOutcomeReviewRender();drawRoomOutcomeReview()};
