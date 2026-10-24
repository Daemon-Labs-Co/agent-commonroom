// Public means shared company rooms, never anonymous or internet access.
function roomSectionKey(room){return room.personalActorId?'private':'public'}
function roomSectionPreferences(){
 try{const saved=JSON.parse(localStorage.getItem('commonroom-room-sections-'+state.currentActorId)||'{}');return {public:saved.public===true,private:saved.private===true}}catch{return {public:false,private:false}}
}
function groupRoomNavigation(){
 if(!state||state.featureFlags?.roomSections===false)return;
 const heading=document.querySelector('#nav .rooms-heading');if(!heading)return;
 const add=heading.querySelector('#new-room');const buttons=[...document.querySelectorAll('#nav [data-room]')];const sections={};const preferences=roomSectionPreferences();
 for(const [key,label] of [['public','Public rooms'],['private','Private rooms']]){
  const section=document.createElement('section');section.className='room-nav-section';const title=document.createElement('div');title.className='rooms-heading';const toggle=document.createElement('button');toggle.type='button';toggle.className='room-section-toggle';toggle.textContent=label;toggle.setAttribute('aria-controls','room-list-'+key);toggle.setAttribute('aria-expanded',String(!preferences[key]));title.append(toggle);if(key==='public'&&add)title.append(add);const list=document.createElement('div');list.className='room-section-list';list.id='room-list-'+key;list.hidden=preferences[key];toggle.onclick=()=>{preferences[key]=!preferences[key];list.hidden=preferences[key];toggle.setAttribute('aria-expanded',String(!preferences[key]));try{localStorage.setItem('commonroom-room-sections-'+state.currentActorId,JSON.stringify(preferences))}catch{}};section.append(title,list);sections[key]=list;heading.before(section);
 }
 for(const button of buttons){const room=state.rooms.find(r=>r.id===button.dataset.room);if(room)sections[roomSectionKey(room)].append(button)}
 heading.remove();
 for(const key of ['public','private'])if(!sections[key].querySelector('[data-room]')){const empty=document.createElement('small');empty.className='room-nav-empty';empty.textContent='No '+key+' rooms';sections[key].append(empty)}
}
const roomSectionsRender=render;render=function(){roomSectionsRender();groupRoomNavigation()};

// Compatibility control for the old implicit agent-room catalog, owner only.
function drawAutomaticAgentRoomsSetting(){
 if(typeof view==='undefined'||view!=='Settings'||state?.currentActorId!=='local-owner')return;
 const main=$('#main');if(!main)return;
 const enabled=state.featureFlags?.automaticAgentRooms===true;
 main.insertAdjacentHTML('beforeend','<section class="card"><h3>Agent rooms</h3><p>Agents work in assigned shared rooms. Existing agent rooms with saved history or work remain available.</p><button id="automatic-agent-rooms" data-feature-flag="automaticAgentRooms">Automatic agent rooms · '+(enabled?'On':'Off')+'</button></section>');
 $('#automatic-agent-rooms').onclick=()=>form('Automatic agent rooms','<label><input type="checkbox" name="enabled" '+(enabled?'checked':'')+'> Restore legacy automatic agent rooms</label><p>Off removes unused implicit agent rooms from the room list. Saved history and work remain available. Human personal rooms are unchanged.</p>',async d=>commit('feature.flag.updated',{name:'automaticAgentRooms',enabled:d.has('enabled')}));
}
const automaticAgentRoomsRender=render;render=function(){automaticAgentRoomsRender();drawAutomaticAgentRoomsSetting()};
drawAutomaticAgentRoomsSetting();
