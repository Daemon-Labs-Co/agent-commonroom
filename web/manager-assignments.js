// Manager-assigned work: an agent may hand work to an agent that reports to it.
// The owner switches it on and sets how many assignments a room allows per day.
const defaultAssignmentLimit=5,maxAssignmentLimit=50;
function roomAssignmentLimit(room){const n=room?.assignmentLimit;return Number.isInteger(n)&&n>=0&&n<=maxAssignmentLimit?n:defaultAssignmentLimit}
function assignmentsUsedToday(room,today){return room?.assignmentDay===today?(room.assignmentsToday||0):0}
function editAssignmentLimit(room){
 form('Manager-assigned tasks per day','<label class="field">Daily limit for this room<input name="limit" type="number" min="0" max="'+maxAssignmentLimit+'" step="1" value="'+roomAssignmentLimit(room)+'" required></label><p class="muted">How many tasks manager agents may assign to their reports in this room each day (UTC). 0 refuses every assignment. Takes effect immediately; assignments already made today still count.</p>',d=>commit('room.assignment_limit.updated',{limit:Number(d.get('limit'))}));
}
function drawManagerAssignments(){
 if(state.currentActorId!=='local-owner')return;
 const on=state.featureFlags?.managerAssignedTasks===true;
 if(view==='Settings'&&!document.querySelector('#configure-manager-assignments')){
  $('#main').insertAdjacentHTML('beforeend','<div class="card"><h3>Agents · Manager-assigned tasks</h3><p>Off by default. Let an agent assign work to an agent that reports to it in the company chart, with [ASSIGN name]. The work starts without your approval, inside the report’s existing permissions and container, up to each room’s daily limit. It can never approve, accept or publish work.</p><button data-feature-flag="managerAssignedTasks" id="configure-manager-assignments">Manager-assigned tasks · '+(on?'On':'Off')+'</button></div>');
  $('#configure-manager-assignments').onclick=()=>form('Manager-assigned tasks','<label><input type="checkbox" name="enabled" '+(on?'checked':'')+'> Allow managers to assign work to their reports</label>',async d=>{await commit('feature.flag.updated',{name:'managerAssignedTasks',enabled:d.has('enabled')});window.location.reload()});
 }
 if(view!=='Room'||!on)return;
 const actions=document.querySelector('.toolbar .actions');const room=state.rooms.find(r=>r.id===roomId);if(!actions||!room)return;
 const b=document.createElement('button');b.id='manager-assignment-limit';b.textContent='Assignments: '+assignmentsUsedToday(room,new Date().toISOString().slice(0,10))+'/'+roomAssignmentLimit(room)+' today';b.title='Change this room’s daily limit for manager-assigned tasks';b.onclick=()=>editAssignmentLimit(room);actions.append(b);
}
const managerAssignmentsRender=render;render=function(){managerAssignmentsRender();drawManagerAssignments()};
