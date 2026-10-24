// Room purpose is human-managed metadata, never an execution authorization.
function roomOutcomeFields(room={}){
 if(state.featureFlags?.roomOutcomes!==true||room.personalActorId)return '';
 return '<label class="field">Defined outcome<textarea name="outcome" maxlength="2000" required>'+esc(room.outcome||'')+'</textarea></label><label class="field">Acceptance criteria<textarea name="acceptanceCriteria" maxlength="8000" required>'+esc(room.acceptanceCriteria||'')+'</textarea></label>';
}
function drawRoomOutcomes(){
 const main=$('#main');if(!main)return;
 if(view==='Settings'&&state.currentActorId==='local-owner'){
  const enabled=state.featureFlags?.roomOutcomes===true;
  main.insertAdjacentHTML('beforeend','<section class="card"><h3>Workflows · Room outcomes</h3><p>Define what a shared room should deliver and how humans will assess it. Existing rooms can be updated gradually.</p><button id="room-outcomes-setting" data-feature-flag="roomOutcomes">Room outcomes · '+(enabled?'On':'Off')+'</button></section>');
  $('#room-outcomes-setting').onclick=()=>form('Room outcomes','<label><input name="enabled" type="checkbox" '+(enabled?'checked':'')+'> Require an outcome and acceptance criteria for new shared rooms</label><p>Off preserves saved outcomes and disables editing them. This does not grant execution permissions or mark any work complete.</p>',d=>commit('feature.flag.updated',{name:'roomOutcomes',enabled:d.has('enabled')}));
 }
 if(view!=='Room'&&view!=='Tasks')return;
 const room=state.rooms.find(r=>r.id===roomId);
 if(!room||room.personalActorId)return;
 if(!room.outcome&&!room.acceptanceCriteria&&state.featureFlags?.roomOutcomes!==true)return;
 main.insertAdjacentHTML('afterbegin','<section class="card room-outcome"><h2>Room deliverable</h2><p style="white-space:pre-wrap">'+esc(room.outcome||'No outcome defined yet. The owner can add one in Room settings.')+'</p><h3>Acceptance criteria</h3><p style="white-space:pre-wrap">'+esc(room.acceptanceCriteria||'Not defined yet.')+'</p><p class="muted">Completion is assessed by a human. This definition does not authorize agent actions.</p></section>');
}
const roomOutcomesRender=render;render=function(){roomOutcomesRender();drawRoomOutcomes()};
