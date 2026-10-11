// Preserve the invitation contract; improve only its optional presentation.
function agentInvitationFields(){
 return '<p class="invitation-intro">Choose a connection and give this agent a role in your squad.</p><div class="invitation-grid"><section><h3>Agent & identity</h3><label class="field">Agent connection<select name="connectionId">'+state.connections.map(c=>'<option value="'+esc(c.id)+'">'+esc(c.name)+'</option>').join('')+'</select></label><p class="muted">Uses an agent connection you have already plugged in.</p>'+field('Name in this room','name',randomName())+field('Room role or custom title','role','Room assistant')+'<p class="muted">The name and role apply to this room. You can edit them later.</p></section><section><h3>Role & behaviour</h3>'+profileFields()+'<p class="muted">New agents start at decision seniority 1 and cannot accept completed work. Adjust their profile after inviting.</p></section><section class="invitation-context"><h3>Starting context <span class="muted">Optional</span></h3><p class="muted">Share notes or information from another conversation to help this agent get started.</p>'+contextFields()+'</section></div>';
}
// Tabs keep the bounded invitation surface readable without burying actions.
function wireInvitationTabs(modal){
 const panels=[...modal.querySelectorAll('.invitation-grid > section')];
 const nav=document.createElement('div');nav.className='invitation-tabs';nav.setAttribute('role','tablist');nav.setAttribute('aria-label','Invitation sections');
 const buttons=['Agent','Role','Context'].map((label,index)=>{
  const button=document.createElement('button');button.type='button';button.textContent=label;button.id='invitation-tab-'+index;button.setAttribute('role','tab');button.setAttribute('aria-controls','invitation-panel-'+index);
  button.onclick=()=>activate(index);button.onkeydown=e=>{let next=index;if(e.key==='ArrowRight')next=(index+1)%3;else if(e.key==='ArrowLeft')next=(index+2)%3;else if(e.key==='Home')next=0;else if(e.key==='End')next=2;else return;e.preventDefault();activate(next);buttons[next].focus()};nav.append(button);return button;
 });
 function activate(selected){panels.forEach((panel,index)=>{panel.hidden=index!==selected;panel.id='invitation-panel-'+index;panel.setAttribute('role','tabpanel');panel.setAttribute('aria-labelledby','invitation-tab-'+index);buttons[index].setAttribute('aria-selected',String(index===selected));buttons[index].tabIndex=index===selected?0:-1;})}
 modal.querySelector('.invitation-grid').before(nav);activate(0);
 modal.querySelector('form').addEventListener('invalid',e=>{const panel=e.target.closest('[role=tabpanel]');const index=panels.indexOf(panel);if(index>=0)activate(index)},true);
 panels.forEach((panel,index)=>{const details=panel.querySelector('details');if(!details)return;if(index===1){const fields=document.createElement('div');fields.className='invitation-role-fields';[...details.children].filter(child=>child.tagName!=='SUMMARY').forEach(child=>fields.append(child));details.replaceWith(fields)}else details.open=true});
}
const previousAgentInvitation=invite;
invite=function(){
 if(state.featureFlags?.agentInvitationLayout===false)return previousAgentInvitation();
 if(!state.connections.length){note('Plug in a connection first');return}
 form('Invite agent',agentInvitationFields(),d=>commit('participant.created',{...Object.fromEntries(d),rank:1,accept:false,challenge:Number(d.get('challenge'))}));
 const modal=$('#modal');modal.classList.add('agent-invitation');modal.addEventListener('close',()=>modal.classList.remove('agent-invitation'),{once:true});
 wireInvitationTabs(modal);
 modal.querySelector('.primary').textContent='Invite agent';wireContextPreview();
};
