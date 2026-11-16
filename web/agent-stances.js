// Stance markers are requests recorded by the room, never approvals or permissions.
const stanceLabels={agree:'Agrees',object:'Objects',escalate:'Escalates',propose:'Proposes a task',discuss:'Starts a discussion',withdraw:'Withdraws objection'};
function openRoomObjections(objections,room){return (objections||[]).filter(o=>o.roomId===room&&o.status==='open')}
function objectionSummary(o){return o.author+' objects'+(o.targetTitle?' to "'+o.targetTitle+'"':'')+(o.targetKind?'':' (blocks nothing)')+': '+(o.reason||'no reason given')}
function resolveObjection(o){
 form('Resolve objection','<p>'+esc(objectionSummary(o))+'</p><label class="field">Decision<select name="decision"><option value="overruled">Overrule: continue as planned</option><option value="upheld">Uphold: the objection is right</option></select></label><label class="field">Note <span class="muted">Optional</span><textarea name="note" maxlength="500"></textarea></label><p class="muted">Either decision closes the objection and removes its block. Upholding does not change the task or proposal; edit or reject it yourself.</p>',d=>commit('objection.resolved',{messageId:o.messageId,decision:d.get('decision'),note:d.get('note')}));
}
function drawAgentStances(){
 if(view==='Settings'&&state.currentActorId==='local-owner'&&!document.querySelector('#configure-agent-objections')){
  const on=state.featureFlags?.agentObjections===true;
  $('#main').insertAdjacentHTML('beforeend','<div class="card"><h3>Agents · Stances and objections</h3><p>Off by default. Let agents mark a reply as agreeing, objecting or escalating. An open objection blocks only the task or proposal it names until a human resolves it or the agent withdraws it. Turning this off stops blocking; messages are kept.</p><button data-feature-flag="agentObjections" id="configure-agent-objections">Stances and objections · '+(on?'On':'Off')+'</button></div>');
  $('#configure-agent-objections').onclick=()=>form('Stances and objections','<label><input type="checkbox" name="enabled" '+(on?'checked':'')+'> Record agent stances and let objections block their target</label>',async d=>{await commit('feature.flag.updated',{name:'agentObjections',enabled:d.has('enabled')});window.location.reload()});
 }
 if(view!=='Room'||state.featureFlags?.agentObjections!==true)return;
 for(const row of document.querySelectorAll('[data-message-id]')){
  const m=state.messages.find(m=>m.id===row.dataset.messageId);if(!m?.stance||!stanceLabels[m.stance])continue;
  const badge=document.createElement('span');badge.className='attention-message-label stance-'+m.stance;badge.textContent=stanceLabels[m.stance];row.querySelector('.message-meta')?.append(badge);
 }
 const open=openRoomObjections(state.objections,roomId);const heading=document.querySelector('.conversation .panel-heading');if(!heading||!open.length)return;
 const human=state.actors.find(a=>a.id===state.currentActorId)?.kind==='human';
 const banner=document.createElement('div');banner.className='attention-banner objection-banner';banner.setAttribute('role','status');
 const title=document.createElement('strong');title.textContent=open.length+' open objection'+(open.length===1?'':'s');banner.append(title);
 for(const o of open){const line=document.createElement('p');line.textContent=objectionSummary(o);if(human){const b=document.createElement('button');b.textContent='Resolve';b.onclick=()=>resolveObjection(o);line.append(' ',b)}banner.append(line)}
 heading.after(banner);
}
const agentStancesRender=render;render=function(){agentStancesRender();drawAgentStances()};
