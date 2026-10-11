// Optional presentation only: existing APIs retain all authorization checks.
function squadProfilePolicy(person, room, actorId){
 const self=person.id===actorId;
 return {self,edit:self||actorId==='local-owner',leave:self&&person.id==='local-owner'&&!room.personalActorId};
}
function openSquadHumanProfile(person, room, oldEdit, oldLeave){
 const policy=squadProfilePolicy(person,room,state.currentActorId);
 if(policy.self&&oldEdit){
  oldEdit.onclick();
  if(policy.leave&&oldLeave){
   const leave=document.createElement('button');leave.type='button';leave.className='danger leave-profile';leave.textContent='Leave room';
   leave.onclick=async()=>{leave.disabled=true;try{await oldLeave.onclick();$('#modal').close()}catch(e){note(e.message)}finally{leave.disabled=false}};
   $('#modal .actions').prepend(leave);
  }
 }else if(policy.edit&&person.id!=='local-owner'){
  memberAccessForm(person);
 }else{
  form('Human profile','<p><strong>'+esc(person.name)+'</strong></p><p>'+esc(person.title||person.role)+' · Human</p><p class="muted">'+(policy.self?'Your profile is managed in Settings.':'Only the owner can change another person’s assigned role or room access.')+'</p>',()=>{});
  $('#modal .primary').hidden=true;$('#cancel').textContent='Close';
 }
}
function squadProfileName(row,open){
 const strong=row.querySelector('.participant-info strong');if(!strong)return;
 const button=document.createElement('button');button.type='button';button.className='squad-name';button.textContent=strong.textContent;button.title='Open profile for '+strong.textContent;button.onclick=open;strong.replaceChildren(button);
}
const squadProfileRender=render;
render=function(){
 squadProfileRender();if(view!=='Room'||state.featureFlags?.squadProfileActions===false)return;
 const room=state.rooms.find(r=>r.id===roomId);
 document.querySelectorAll('.squad .participant').forEach(row=>{
  if(row.dataset.humanId){
   const person=state.actors.find(a=>a.id===row.dataset.humanId);if(!person)return;
   const edit=row.querySelector('[data-human-edit]');const leave=row.querySelector('#leave-room');
   squadProfileName(row,()=>openSquadHumanProfile(person,room,edit,leave));edit?.remove();leave?.remove();
  }else{
   const edit=row.querySelector('[data-profile]');if(edit)squadProfileName(row,()=>edit.onclick());
  }
 });
};
