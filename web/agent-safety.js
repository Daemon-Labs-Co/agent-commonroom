'use strict';
// This flag controls button visibility only. Backend safety remains mandatory.
function agentSafetyChoice(globalStop,roomStop,scope){return {action:(scope==='application'?globalStop:roomStop)?'resume':'stop',disabled:scope==='room'&&globalStop}}
function agentSafetyForm(scope,action){
 const target=scope==='application'?'all agents':'this room’s agents';
 form((action==='stop'?'Stop ':'Resume ')+target,'<p>'+(action==='stop'?'Block new turns and request cancellation of managed turns. Remote processes and already launched external tools may continue; verify their shutdown separately.':'Allow new requests again. Cancelled work is not automatically restarted.')+'</p><label class="field">Reason (optional)<input name="reason" maxlength="500"></label>',async data=>{await api('agent-safety',{action,scope,reason:data.get('reason')});await refresh();note(action==='stop'?'Dispatch blocked; cancellation requested.':'New agent requests enabled.');});
 const button=document.querySelector('#modal button.primary');button.textContent=action==='stop'?'Stop agents':'Resume dispatch';if(action==='stop')button.classList.add('danger');
}
const safetyRender=render;render=function(){
 safetyRender();document.querySelector('#emergency-all')?.remove();
 if(state.agentsStopped||state.agentRoomStopped){
  document.querySelectorAll('#newchat,#attach,#invite,[data-chat],[data-run-task],[data-run-review]').forEach(b=>b.disabled=true);
  const targets=document.querySelector('#send select');if(targets){targets.querySelectorAll('option').forEach(o=>o.disabled=o.value!=='room');targets.value='room';recipient='room'}
 }
 if(state.currentActorId!=='local-owner'||state.featureFlags?.agentSafetyControls===false)return;
 const all=agentSafetyChoice(state.agentsStopped,state.agentRoomStopped,'application');const button=document.createElement('button');button.id='emergency-all';button.textContent=all.action==='stop'?'Stop all agents':'Resume all agents';button.className=all.action==='stop'?'danger':'';button.onclick=()=>agentSafetyForm('application',all.action);document.querySelector('header .environment').prepend(button);
 if(view==='Room') {const action=agentSafetyChoice(state.agentsStopped,state.agentRoomStopped,'room');const local=document.createElement('button');local.textContent=action.action==='stop'?'Stop room agents':'Resume room agents';local.disabled=action.disabled;local.className=action.action==='stop'?'danger':'';local.onclick=()=>agentSafetyForm('room',action.action);document.querySelector('#main .toolbar .actions')?.append(local)}
};
