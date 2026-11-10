// Agent discussions: a human asks the room's agents to talk it through. Every
// turn is an ordinary bounded agent turn; nothing here approves or authorizes work.
const discussionRoundCeiling=3;
function roomDiscussionLimit(room){const n=room?.discussionRounds;return n>=1&&n<discussionRoundCeiling?n:discussionRoundCeiling}
function discussionRoundOptions(limit){let html='';for(let n=limit;n>=1;n--)html+='<option value="'+n+'">'+n+(n===1?' round':' rounds')+'</option>';return html}
function startAgentDiscussion(room){
 const limit=roomDiscussionLimit(room);
 form('Start a discussion','<label class="field">Topic<textarea name="text" required maxlength="4000" placeholder="What should the agents talk through?"></textarea></label><label class="field">Rounds<select name="rounds">'+discussionRoundOptions(limit)+'</select></label><p class="muted">Each agent replies in turn and sees the others’ replies. It ends when every agent agrees, an agent escalates to you, the rounds run out, or you post or stop it. Every reply counts against the daily turn budget.</p>',async d=>{
  const result=await api('agents',{action:'discuss',text:d.get('text'),rounds:Number(d.get('rounds'))});await refresh();
  note('Discussion started with '+result.agents+' agents for up to '+result.rounds+(result.rounds===1?' round':' rounds'));
 });
 $('#modal .primary').textContent='Start discussion';
}
function editDiscussionLimit(room){
 form('Discussion round limit','<label class="field">Most rounds allowed in this room<select name="rounds">'+discussionRoundOptions(discussionRoundCeiling).replace('value="'+roomDiscussionLimit(room)+'"','value="'+roomDiscussionLimit(room)+'" selected')+'</select></label><p class="muted">You can lower the limit for this room. It cannot be raised above '+discussionRoundCeiling+'.</p>',d=>commit('room.discussion_limit.updated',{rounds:Number(d.get('rounds'))}));
}
function drawAgentDiscussion(){
 const owner=state.currentActorId==='local-owner';
 if(view==='Settings'&&owner&&!document.querySelector('#configure-agent-discussion')){
  const on=state.featureFlags?.agentDiscussion===true;
  $('#main').insertAdjacentHTML('beforeend','<div class="card"><h3>Agents · Discussions</h3><p>Off by default. Let a human ask a room’s agents to discuss a topic: they reply to each other in turn for a limited number of rounds. Turning this off ends any running discussion at its next turn; messages are kept.</p><button data-feature-flag="agentDiscussion" id="configure-agent-discussion">Discussions · '+(on?'On':'Off')+'</button></div>');
  $('#configure-agent-discussion').onclick=()=>form('Agent discussions','<label><input type="checkbox" name="enabled" '+(on?'checked':'')+'> Allow agent discussions</label>',async d=>{await commit('feature.flag.updated',{name:'agentDiscussion',enabled:d.has('enabled')});window.location.reload()});
 }
 if(view!=='Room'||state.featureFlags?.agentDiscussion!==true)return;
 const actions=document.querySelector('.toolbar .actions');const room=state.rooms.find(r=>r.id===roomId);
 if(!actions||!room||state.actors.find(a=>a.id===state.currentActorId)?.kind!=='human')return;
 const button=document.createElement('button');button.id='agent-discussion';
 if(room.discussionId){button.textContent='Stop discussion';button.className='danger';button.onclick=async()=>{button.disabled=true;try{await api('agents',{action:'stop-discussion'});await refresh()}catch(e){note(e.message);button.disabled=false}}}
 else{button.textContent='Start discussion';button.onclick=()=>startAgentDiscussion(room)}
 actions.append(button);
 if(owner&&!room.discussionId){const limit=document.createElement('button');limit.id='agent-discussion-limit';limit.textContent='Rounds: '+roomDiscussionLimit(room);limit.title='Change this room’s discussion round limit';limit.onclick=()=>editDiscussionLimit(room);actions.append(limit)}
}
const agentDiscussionRender=render;render=function(){agentDiscussionRender();drawAgentDiscussion()};
