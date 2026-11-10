// Agent task suggestions are prompts for a human. Capturing one uses the existing
// owner-only proposal flow; nothing here creates, approves or starts work.
function openTaskSuggestions(suggestions,room){return (suggestions||[]).filter(s=>s.roomId===room&&s.status==='open')}
function suggestionLine(s){return s.author+' suggests: '+s.title+(s.assignee?' (for '+s.assignee+')':'')+(s.criteria?' — '+s.criteria:'')}
async function captureTaskSuggestion(s){
 await api('task-proposals?room='+encodeURIComponent(s.roomId),{action:'capture',sourceMessageId:s.messageId,title:s.title,criteria:s.criteria||''});
 await refresh();note('Captured as a proposal. Review and approve it under Tasks.');
}
function drawTaskSuggestions(){
 const owner=state.currentActorId==='local-owner';
 if(view==='Settings'&&owner&&!document.querySelector('#configure-task-suggestions')){
  const on=state.featureFlags?.governingSuggestions===true;
  $('#main').insertAdjacentHTML('beforeend','<div class="card"><h3>Agents · Task suggestions</h3><p>Off by default. Let agents suggest tasks with [PROPOSE TASK], let a human ask the room’s governing agent for suggestions, and ask automatically when a discussion ends in agreement. Suggestions create nothing until a human captures and approves them.</p><button data-feature-flag="governingSuggestions" id="configure-task-suggestions">Task suggestions · '+(on?'On':'Off')+'</button></div>');
  $('#configure-task-suggestions').onclick=()=>form('Task suggestions','<label><input type="checkbox" name="enabled" '+(on?'checked':'')+'> Allow agent task suggestions</label>',async d=>{await commit('feature.flag.updated',{name:'governingSuggestions',enabled:d.has('enabled')});window.location.reload()});
 }
 if(view!=='Room'||state.featureFlags?.governingSuggestions!==true)return;
 const human=state.actors.find(a=>a.id===state.currentActorId)?.kind==='human';if(!human)return;
 const actions=document.querySelector('.toolbar .actions');
 if(actions){const ask=document.createElement('button');ask.id='suggest-tasks';ask.textContent='Suggest tasks';ask.onclick=async()=>{ask.disabled=true;try{const r=await api('agents',{action:'suggest-tasks'});await refresh();note('Asked '+r.name+' to suggest tasks')}catch(e){note(e.message)}finally{ask.disabled=false}};actions.append(ask)}
 const open=openTaskSuggestions(state.taskSuggestions,roomId);const heading=document.querySelector('.conversation .panel-heading');if(!heading||!open.length)return;
 const canCapture=owner&&state.featureFlags?.taskProposals===true;
 const banner=document.createElement('div');banner.className='attention-banner suggestion-banner';banner.setAttribute('role','status');
 const title=document.createElement('strong');title.textContent=open.length+' suggested task'+(open.length===1?'':'s');banner.append(title);
 if(owner&&!canCapture){const hint=document.createElement('p');hint.textContent='Turn on Task proposals in Settings to capture these.';banner.append(hint)}
 for(const s of open){
  const line=document.createElement('p');line.textContent=suggestionLine(s)+' ';
  if(canCapture){const c=document.createElement('button');c.textContent='Capture';c.onclick=async()=>{c.disabled=true;try{await captureTaskSuggestion(s)}catch(e){note(e.message);c.disabled=false}};line.append(c,' ')}
  const x=document.createElement('button');x.textContent='Dismiss';x.onclick=()=>commit('task.suggestion.dismissed',{id:s.id}).catch(e=>note(e.message));line.append(x);
  banner.append(line);
 }
 heading.after(banner);
}
const taskSuggestionsRender=render;render=function(){taskSuggestionsRender();drawTaskSuggestions()};
