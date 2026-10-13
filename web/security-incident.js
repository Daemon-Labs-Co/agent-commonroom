// Mandatory incident recovery remains owner-authorized and verified server-side.
function drawSecurityIncident(){
 if(!state?.incident||state.currentActorId!=='local-owner')return;
 const button=document.createElement('button');button.type='button';button.className='danger';button.textContent='Review integrity incident';
 button.onclick=()=>{
  form('Review integrity incident','<p>Agents remain stopped until the audit chain is valid and the approved worker is confirmed idle. Recovery records your decision; it cannot rewrite history or automatically replay work.</p><label class="field">Recovery decision<textarea name="reason" required maxlength="500" placeholder="What was investigated and why recovery is safe"></textarea></label>',async data=>{await api('integrity',{reason:data.get('reason')});await refresh();note('Verified recovery recorded. Previous tasks were not restarted.');});
  const modal=$('#modal'),help=document.createElement('button');help.type='button';help.className='context-help overlay-help help-icon';help.textContent='?';help.setAttribute('aria-label','Help with integrity recovery');help.onclick=()=>openDocumentation('operations');modal.append(help);modal.classList.add('has-overlay-help');modal.querySelector('button.primary').textContent='Verify and record recovery';
 };
 $('#main .toolbar')?.append(button);
}
const securityIncidentRender=render;render=function(){securityIncidentRender();drawSecurityIncident()};
