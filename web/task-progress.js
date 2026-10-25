// Read-only durable progress. Actions keep their existing authorization paths.
let taskProgressGeneration=0;
const taskProgressLabels={legacy:'Legacy task','legacy-admitted':'Legacy work admitted','legacy-result-recorded':'Legacy result recorded','definition-required':'Define task','not-run':'Not run','work-admitted':'Work admitted','review-admitted':'Agent review admitted','reconciliation-required':'Reconcile pending work','inputs-unusable':'Inputs need attention','new-work-required':'New work required','retry-required':'Owner retry required','owner-review-required':'Owner review required','criterion-review-required':'Criterion review required',accepted:'Accepted by owner'};
function taskProgressLabel(phase){return taskProgressLabels[phase]||phase}
function taskProgressTask(id,revision){const task=state.tasks.find(t=>t.id===id&&t.roomId===roomId&&t.revision===revision);if(!task)throw Error('Task changed; refresh before taking action.');return task}
function taskProgressCard(t,result,owner){
 let actions='';
 if(t.executionId)actions+='<button data-task-progress-action="history" data-task-id="'+esc(t.taskId)+'" data-task-revision="'+esc(t.taskRevision)+'">Inspect attempts and values</button>';
 if(owner){
  actions+='<button data-task-progress-action="edit" data-task-id="'+esc(t.taskId)+'" data-task-revision="'+esc(t.taskRevision)+'">Edit task</button>';
  if(t.dispatchStatus!=='pending'&&result.structuredDispatchEnabled&&['not-run','new-work-required','retry-required','inputs-unusable'].includes(t.phase))actions+='<button data-task-progress-action="run" data-task-id="'+esc(t.taskId)+'" data-task-revision="'+esc(t.taskRevision)+'">Bind inputs and run</button>';
  if(result.criterionReviewEnabled&&['criterion-review-required','accepted'].includes(t.phase))actions+='<button data-task-progress-action="review" data-task-id="'+esc(t.taskId)+'" data-task-revision="'+esc(t.taskRevision)+'">Review criteria</button>';
 }
 return '<article class="task"><h3>'+esc(t.title)+'</h3><p><strong>'+esc(taskProgressLabel(t.phase))+'</strong> · Saved task state: '+esc(t.taskStatus)+'</p><p>'+esc(t.detail)+'</p><p>Definition revision '+esc(t.definitionRevision||'Legacy')+' · Bound input state: '+esc(t.inputState)+'</p>'+(t.executionId?'<p>Latest admitted attempt '+esc(t.executionId)+' · '+esc(t.executionMode)+' · Recorded outcome '+esc(t.executionStatus)+'</p>':'')+(t.workExecutionId?'<p>Work attempt '+esc(t.workExecutionId)+' · '+esc(t.returnedValueCount)+' returned values. Structural validity is not independent verification.</p>':'')+(t.dependencies?.length?'<details><summary>Bound input dependencies</summary><ul>'+t.dependencies.map(d=>'<li>'+esc(d.inputName)+' ← Task '+esc(d.sourceTaskId)+' revision '+esc(d.sourceTaskRevision)+' · '+esc(d.executionId+'/'+d.outputName)+' · '+(d.current?'Current at read time':'Needs attention')+'</li>').join('')+'</ul></details>':'')+actions+'</article>';
}
async function loadTaskProgress(cursor='',generation=++taskProgressGeneration,rid=roomId){
 const result=await api('task-progress?room='+encodeURIComponent(rid)+(cursor?'&after='+encodeURIComponent(cursor):''));
 if(generation!==taskProgressGeneration||roomId!==rid||view!=='Tasks'||state.featureFlags?.taskProgress!==true)return;
 const panel=$('#task-progress-panel');if(!panel)return;
 const owner=state.currentActorId==='local-owner';
 panel.innerHTML='<h2>Task progress</h2><p>Controller-recorded state only. Admission does not prove a running worker; returned values do not prove correctness or observed checks.</p><p>This page: '+esc((result.tasks||[]).length)+' of '+esc(result.totalTasks)+' room tasks. Structured dispatch '+(result.structuredDispatchEnabled?'enabled; availability still checked at dispatch':'disabled')+'.</p><p>Page summary: '+Object.entries(result.pageCounts||{}).map(([phase,count])=>esc(taskProgressLabel(phase))+': '+esc(count)).join(' · ')+'</p>'+(result.tasks||[]).map(t=>taskProgressCard(t,result,owner)).join('')+(!result.tasks?.length?'<p>No room tasks on this page.</p>':'')+(result.nextCursor?'<button id="task-progress-next">Next tasks</button>':'')+(cursor?'<button id="task-progress-first">First tasks</button>':'')+'<button id="task-progress-refresh">Refresh progress</button>';
 const load=next=>loadTaskProgress(next).catch(e=>note(e.message));
 if(result.nextCursor)$('#task-progress-next').onclick=()=>load(result.nextCursor);
 if(cursor)$('#task-progress-first').onclick=()=>load('');
 $('#task-progress-refresh').onclick=()=>load(cursor);
 document.querySelectorAll('#task-progress-panel [data-task-progress-action]').forEach(button=>button.onclick=async()=>{
  try{
   if(generation!==taskProgressGeneration||roomId!==rid||view!=='Tasks')throw Error('Progress changed; refresh before taking action.');
   const task=taskProgressTask(button.dataset.taskId,Number(button.dataset.taskRevision));
   switch(button.dataset.taskProgressAction){case 'history':await showTaskExecutionHistory(task);break;case 'review':if(!owner)throw Error('Owner review required');await showTaskAcceptance(task);break;case 'edit':if(!owner)throw Error('Owner permission required');taskForm(task);break;case 'run':if(!owner)throw Error('Owner permission required');await dispatchTask(task,'work');break}
  }catch(e){note(e.message)}
 });
}
function drawTaskProgress(){
 const generation=++taskProgressGeneration;
 if(view==='Settings'&&state.currentActorId==='local-owner'){
  const enabled=state.featureFlags?.taskProgress===true;
  $('#main').insertAdjacentHTML('beforeend','<section class="card"><h3>Workflows · Task progress</h3><p>Inspect recorded definitions, attempts, returned values, input dependencies and human acceptance in shared rooms. This never probes workers or launches work automatically.</p><button id="task-progress-setting" data-feature-flag="taskProgress">Task progress · '+(enabled?'On':'Off')+'</button></section>');
  $('#task-progress-setting').onclick=()=>form('Task progress','<label><input name="enabled" type="checkbox" '+(enabled?'checked':'')+'> Enable the recorded progress view</label><p>Off hides this view and blocks its API. Tasks, workers, reviews and history retain their ordinary controls and authority.</p>',data=>commit('feature.flag.updated',{name:'taskProgress',enabled:data.has('enabled')}));
 }
 if(view!=='Tasks'||state.featureFlags?.taskProgress!==true)return;
 const room=state.rooms.find(r=>r.id===roomId);if(!room||room.personalActorId)return;
 $('#main').insertAdjacentHTML('afterbegin','<section class="card" id="task-progress-panel"><h2>Task progress</h2><p>Loading recorded progress…</p></section>');
 const rid=roomId;loadTaskProgress('',generation,rid).catch(e=>{if(generation===taskProgressGeneration&&view==='Tasks'&&roomId===rid){const panel=$('#task-progress-panel');if(panel)panel.textContent=e.message}});
}
const taskProgressRender=render;render=function(){taskProgressRender();drawTaskProgress()};
