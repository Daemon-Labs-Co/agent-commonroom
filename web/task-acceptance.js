// Owner judgments bind exact immutable output snapshots. Model review is advice.
let taskAcceptanceViewGeneration=0;
function taskAssessmentHistory(review,execution){return '<article><h3>Review '+esc(review.revision)+' · '+esc(review.action)+'</h3><p>'+esc(review.reviewedBy)+' · '+esc(review.reviewedAt)+'</p><p>'+esc(review.note)+'</p><p>Work execution '+esc(review.executionId)+' · Snapshot '+esc(review.snapshot)+'</p>'+(execution?'<h4>Values from this reviewed attempt</h4>'+(execution.outputs||[]).map(o=>'<p>'+esc(o.name)+' · '+esc(o.digest)+'</p><pre>'+esc(o.valueJson)+'</pre>').join('')+tdSummary(execution.definition):'<p>Reviewed execution is unavailable; no value is inferred.</p>')+review.assessments.map(a=>'<section><strong>'+esc(a.criterionId)+' · '+esc(a.method)+' · '+esc(a.decision)+'</strong><p>'+esc(a.note)+'</p><p>'+esc(a.outputName)+' · SHA-256 '+esc(a.outputDigest||'No returned value')+'</p></section>').join('')+'</article>'}
function taskCriterionReviewFields(definition,schemas){
 return definition.acceptanceCriteria.map((c,i)=>{
  const heading='<h3>'+esc(c.id)+'</h3><p>'+esc(c.description)+'</p><p>Output: '+esc(c.output)+'</p>';
  if(c.method==='schema'){const a=schemas.find(a=>a.criterionId===c.id);return heading+'<p>Server structural check: '+esc(a?.decision||'unavailable')+'</p><p>'+esc(a?.note||'No structural assessment available.')+'</p>'}
  return heading+'<label class="field">Human judgment<select name="criterion-'+i+'" required><option value="" disabled selected>Choose a decision</option><option value="passed">Passed</option><option value="failed">Failed</option></select></label><label class="field">What you checked and why<textarea name="criterion-note-'+i+'" maxlength="1000" required></textarea></label>';
 }).join('');
}
async function showTaskAcceptance(task,cursor=''){
 const generation=++taskAcceptanceViewGeneration,rid=roomId,requestedView=view;
 const result=await api('task-acceptance?room='+encodeURIComponent(rid)+'&taskId='+encodeURIComponent(task.id)+(cursor?'&before='+encodeURIComponent(cursor):''));
 if(generation!==taskAcceptanceViewGeneration||roomId!==rid||view!==requestedView)return;
 const snapshotTask=result.task,work=result.workExecution||{},owner=state.currentActorId==='local-owner';
 const canDecide=owner&&result.enabled&&result.ready&&!cursor;
 const inputs=(work.inputs||[]).map(input=>'<p>'+esc(input.name)+' · '+esc(input.sourceVersion)+' · SHA-256 '+esc(input.digest)+'</p><pre>'+esc(input.valueJson)+'</pre>').join('');
 const outputs=(work.outputs||[]).map(o=>'<article><h3>'+esc(o.name)+'</h3><p>Immutable output '+esc(work.id+'/'+o.name)+' · SHA-256 '+esc(o.digest)+'</p><pre>'+esc(o.valueJson)+'</pre></article>').join('');
 const formFields=canDecide?taskCriterionReviewFields(snapshotTask.definition,result.schemaAssessments||[])+selectField('Owner decision','action',['changes','accept'],'changes')+'<label class="field">Overall review rationale<textarea name="review-note" maxlength="2000" required></textarea></label>':'<p>'+(!result.enabled?'Criterion review is disabled; recorded history is preserved.':!result.ready?'Current valid outputs and input dependencies are required before a decision.':'History view. Return to the latest review to decide.')+'</p>';
 form((canDecide?'Review ':'Criterion reviews · ')+esc(snapshotTask.title),'<p>Criterion review: '+esc(result.status)+' · Current task state: '+esc(snapshotTask.status)+'</p><p>Definition revision '+esc(snapshotTask.definition.revision)+' · Task revision '+esc(snapshotTask.revision)+'</p>'+tdSummary(snapshotTask.definition)+'<h3>Exact bound inputs</h3>'+inputs+'<h3>Current returned values</h3>'+outputs+formFields+'<p>Schema checks establish structure only. Agent findings do not establish human acceptance or independently observed checks. Saving does not launch work.</p><h3>Recorded reviews</h3>'+(result.reviews||[]).map(review=>taskAssessmentHistory(review,result.reviewedExecutions?.[review.executionId])).join('')+(result.nextCursor?'<button type="button" id="task-review-older">Older reviews</button>':'')+(cursor?'<button type="button" id="task-review-latest">Latest review</button>':''),async data=>{
  if(!canDecide)return;
  if(roomId!==rid||view!==requestedView||generation!==taskAcceptanceViewGeneration)throw Error('Room or review changed; reopen the task review.');
  const decisions=[];snapshotTask.definition.acceptanceCriteria.forEach((c,i)=>{if(c.method==='human')decisions.push({criterionId:c.id,decision:data.get('criterion-'+i),note:data.get('criterion-note-'+i)})});
  await api('task-acceptance?room='+encodeURIComponent(rid),{taskId:snapshotTask.id,taskRevision:snapshotTask.revision,reviewRevision:result.review?.revision||0,snapshot:result.snapshot,action:data.get('action'),note:data.get('review-note'),decisions});await refresh();
 });
 if(result.nextCursor)$('#task-review-older').onclick=()=>showTaskAcceptance(task,result.nextCursor).catch(e=>note(e.message));
 if(cursor)$('#task-review-latest').onclick=()=>showTaskAcceptance(task).catch(e=>note(e.message));
}
const taskAcceptanceCard=taskCard;taskCard=function(t){const html=taskAcceptanceCard(t);return t.definition?html.replace('</article>','<button data-task-criterion-review="'+esc(t.id)+'">'+(state.featureFlags?.taskAcceptanceReview===true&&state.currentActorId==='local-owner'?'Review task criteria':'Criterion review history')+'</button></article>'):html};
const taskAcceptanceWireTasks=wireTasks;wireTasks=function(){taskAcceptanceWireTasks();document.querySelectorAll('[data-task-criterion-review]').forEach(b=>b.onclick=()=>showTaskAcceptance(state.tasks.find(t=>t.id===b.dataset.taskCriterionReview)).catch(e=>note(e.message)))};
function drawTaskAcceptanceSettings(){
 if(view!=='Settings'||state.currentActorId!=='local-owner')return;
 const enabled=state.featureFlags?.taskAcceptanceReview===true;
 $('#main').insertAdjacentHTML('beforeend','<section class="card"><h3>Workflows · Task criterion review</h3><p>Require owner decisions on every human criterion against exact returned values before accepting structured tasks. Requires Task definitions.</p><button id="task-acceptance-setting" data-feature-flag="taskAcceptanceReview">Task criterion review · '+(enabled?'On':'Off')+'</button></section>');
 $('#task-acceptance-setting').onclick=()=>form('Task criterion review','<label><input name="enabled" type="checkbox" '+(enabled?'checked':'')+'> Enable explicit criterion review</label><p>Off preserves reviews and ordinary owner acceptance. Turning this on requires fresh explicit review of structured tasks used as accepted inputs or room deliverables. No model can accept work.</p>',data=>commit('feature.flag.updated',{name:'taskAcceptanceReview',enabled:data.has('enabled')}));
}
const taskAcceptanceRender=render;render=function(){taskAcceptanceRender();drawTaskAcceptanceSettings()};
