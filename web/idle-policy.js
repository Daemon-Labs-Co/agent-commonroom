// Policy controls remain available when idle shutdown is off, so owners can
// re-enable it. Runtime admission enforces the saved flag independently.
function drawIdlePolicySettings(){
 if(view!=='Settings'||state?.currentActorId!=='local-owner')return;
 const target=$('#main');if(!target||target.querySelector('#configure-idle-policy'))return;
 const enabled=state.featureFlags?.idleAgentShutdown!==false;
 const saved=state.agentIdleSettings||{},minutes=saved.minutes||5;
 const notices=state.featureFlags?.idleStopNotices!==false;
 target.insertAdjacentHTML('beforeend','<section class="card"><h3>Security · Idle agents</h3><p>Policy for newly admitted development jobs: '+esc(minutes)+' idle minutes. Active work and source are protected. Production development dispatch remains gated.</p><button data-feature-flag="idleAgentShutdown" id="toggle-idle-policy">Idle shutdown · '+(enabled?'On':'Off')+'</button><button id="configure-idle-policy">Set idle interval</button></section>');
 $('#toggle-idle-policy').onclick=()=>form('Idle shutdown','<label><input type="checkbox" name="enabled" '+(enabled?'checked':'')+'> Stop idle pairs for newly admitted jobs</label><p>Existing jobs retain their recorded policy. Hard deadlines and ordinary shutdown remain enforced.</p>',async d=>{await commit('feature.flag.updated',{name:'idleAgentShutdown',enabled:d.has('enabled')});await refresh()});
 target.insertAdjacentHTML('beforeend','<section class="card"><h3>Security · Idle-stop notices</h3><p>Show recorded idle-stop intentions in their room. Notices do not verify cleanup. Hiding them preserves required audit records.</p><button data-feature-flag="idleStopNotices" id="toggle-idle-notices">Idle-stop notices · '+(notices?'On':'Off')+'</button></section>');
 $('#toggle-idle-notices').onclick=()=>form('Idle-stop notices','<label><input type="checkbox" name="enabled" '+(notices?'checked':'')+'> Show room idle-stop notices</label><p>Audit recording remains mandatory. Existing notices reappear when enabled.</p>',async d=>{await commit('feature.flag.updated',{name:'idleStopNotices',enabled:d.has('enabled')});await refresh()});
 $('#configure-idle-policy').onclick=()=>{
 const revision=saved.revision||0;
 form('Idle interval','<label>Idle minutes<input type="number" name="minutes" min="1" max="60" step="1" value="'+esc(minutes)+'" required></label><p>Default: 5 minutes. Changes apply to new job allocations. The current hard job deadline can stop a pair sooner.</p>',async d=>{await commit('agents.idle_policy.updated',{minutes:Number(d.get('minutes')),revision});await refresh()});
 };
}
const idlePolicyRender=render;render=function(){idlePolicyRender();drawIdlePolicySettings()};
