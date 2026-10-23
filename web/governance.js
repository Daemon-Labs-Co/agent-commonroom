// Appointment state confers coordination only; the server owns permissions.
function drawGovernanceControls(){
 if(state?.currentActorId!=='local-owner')return;
 const target=$('#main');if(!target)return;
 const enabled=state.featureFlags?.agentCoordination===true;
 if(view==='Settings'){
  target.insertAdjacentHTML('beforeend','<section class="card"><h3>Agent coordination</h3><p>Choose which agent coordinates a room. Selection never grants extra permissions or approves work.</p><button id="toggle-coordination" data-feature-flag="agentCoordination">Agent coordination · '+(enabled?'On':'Off')+'</button></section>');
  $('#toggle-coordination').onclick=()=>form('Agent coordination','<label><input type="checkbox" name="enabled" '+(enabled?'checked':'')+'> Enable room appointments</label><p>Off by default. Disabling cancels role-dependent turns and preserves appointments and history. Execution permissions, human stop and security safeguards remain enforced.</p>',async d=>commit('feature.flag.updated',{name:'agentCoordination',enabled:d.has('enabled')}));
 }
 if(view==='Room'&&enabled){
  target.insertAdjacentHTML('beforeend','<section class="card"><button id="choose-governor">Choose coordinating agent</button><button id="coordinate-turn">Ask current coordinator</button></section>');
  $('#coordinate-turn').onclick=async()=>{
   const rid=roomId,current=await api('governance?room='+encodeURIComponent(rid));
   if(roomId!==rid||view!=='Room')return;
   if(!current.available){alert('The coordinator is unavailable. Select an eligible agent first.');return;}
   const appointment=current.appointment;
   form('Ask current coordinator','<p>Room: '+esc(rid)+'. Coordinator: '+esc(appointment.participantId)+'. Revision: '+esc(appointment.revision)+'.</p><p>This turn depends on the current appointment. Replacement or disabling coordination stops it. Existing permissions and execution restrictions still apply.</p><label>Request<textarea name="text" required></textarea></label>',async d=>{await api('agents?room='+encodeURIComponent(rid),{action:'send',participantId:appointment.participantId,coordinationRevision:appointment.revision,text:d.get('text')});await refresh()});
  };
  $('#choose-governor').onclick=async()=>{
   const rid=roomId,current=await api('governance?room='+encodeURIComponent(rid));
   if(roomId!==rid||view!=='Room')return;
   const options=state.participants.filter(p=>p.roomId===rid&&(p.status==='Idle'||p.status==='Working')&&state.connections.some(c=>c.id===p.connectionId&&c.protocol==='CLI'&&(c.endpoint==='codex'||c.endpoint==='claude'||/^ssh:\/\/[A-Za-z0-9][A-Za-z0-9_.-]*(?:\/claude)?$/.test(c.endpoint)))).map(p=>'<option value="'+esc(p.id)+'" '+(p.id===current.appointment.participantId?'selected':'')+'>'+esc(p.name)+'</option>').join('');
   form('Choose coordinating agent','<p>Room: '+esc(rid)+'</p><p>Confirm replacement of the current coordinator. This grants no additional access or task approval.</p><label>Agent<select name="participantId" required>'+options+'</select></label>',async d=>{await api('governance?room='+encodeURIComponent(rid),{participantId:d.get('participantId'),revision:current.appointment.revision||0});await refresh()});
  };
 }
}
const governanceRender=render;render=function(){governanceRender();drawGovernanceControls()};
