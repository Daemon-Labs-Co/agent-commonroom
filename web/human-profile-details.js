// Read-only projection of authorized state. Never render keys or connection secrets.
function humanAccountDetails(person, currentRoom){
 const rooms=(state.rooms||[]).filter(r=>r.personalActorId===person.id||(!r.personalActorId&&(person.id==='local-owner'||(person.roomIds||[]).includes(r.id))));
 const entries=[['Company',state.company?.name||'Company'],['Assigned role',person.role||'Human'],['Account status',person.enabled?'Enabled':'Disabled'],['Account ID',person.id],['Personal room',rooms.find(r=>r.personalActorId===person.id)?.name||'Not visible in this view'],['Shared room access',rooms.filter(r=>!r.personalActorId).map(r=>r.name).join(', ')||'No visible shared rooms']];
 if(currentRoom){const owner=person.id==='local-owner';entries.push(['Current room',currentRoom.name],['Room presence',owner&&currentRoom.humanPresent===false?'Away':'Member'],['Decision seniority',String(owner?(currentRoom.humanRank??11):11)]);}
 return '<section class="human-account-details"><h3>Account and rooms</h3><dl>'+entries.map(([key,value])=>'<div><dt>'+esc(key)+'</dt><dd>'+esc(value)+'</dd></div>').join('')+'</dl><p class="muted">Account role and room access are managed by the owner. Only rooms visible to you are shown. These details are read-only.</p></section>';
}
function appendHumanAccountDetails(container,person,currentRoom){
 if(!container||!person||state?.featureFlags?.humanProfileDetails===false)return;
 const actions=container.querySelector?.('.actions');
 (actions||container).insertAdjacentHTML(actions?'beforebegin':'beforeend',humanAccountDetails(person,currentRoom));
}
const humanDetailsRender=render;render=function(){
 humanDetailsRender();if(view!=='Settings'||!state||state.featureFlags?.humanProfileDetails===false)return;
 const person=state.actors.find(a=>a.id===state.currentActorId);
 const card=[...document.querySelectorAll('#main .card')].find(c=>/^Your profile$/.test(c.querySelector('h1,h2')?.textContent||''));
 appendHumanAccountDetails(card,person);
};
const humanDetailsForm=form;form=function(title,fields,save){
 humanDetailsForm(title,fields,save);
 if(title==='Your profile'||title==='Your room profile')appendHumanAccountDetails($('#modal form'),state?.actors?.find(a=>a.id===state.currentActorId),title==='Your room profile'?state.rooms.find(r=>r.id===roomId):null);
};
if(typeof openSquadHumanProfile==='function'){
 const humanDetailsOpen=openSquadHumanProfile;openSquadHumanProfile=function(person,room,edit,leave){
  humanDetailsOpen(person,room,edit,leave);
  const container=$('#modal form');if(!container?.querySelector('.human-account-details'))appendHumanAccountDetails(container,person,room);
 };
}
