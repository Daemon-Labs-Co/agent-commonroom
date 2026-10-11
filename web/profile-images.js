// Photos are presentation metadata on an authenticated human account.
function humanPhoto(actor, enabled=true){return enabled && /^data:image\/jpeg;base64,[A-Za-z0-9+/]+=*$/.test(actor?.image||'')?actor.image:''}
async function resizeHumanPhoto(file){
 if(!['image/jpeg','image/png'].includes(file.type)||file.size>5*1024*1024)throw Error('Choose a JPEG or PNG photo up to 5 MB.');
 const bitmap=await createImageBitmap(file);
 try{
 if(bitmap.width*bitmap.height>16000000)throw Error('Choose a photo smaller than 16 megapixels.');
 const canvas=document.createElement('canvas');const scale=Math.min(1,160/bitmap.width,160/bitmap.height);
 canvas.width=Math.max(1,Math.round(bitmap.width*scale));canvas.height=Math.max(1,Math.round(bitmap.height*scale));
 const context=canvas.getContext('2d');context.fillStyle='#ffffff';context.fillRect(0,0,canvas.width,canvas.height);context.drawImage(bitmap,0,0,canvas.width,canvas.height);
 return canvas.toDataURL('image/jpeg',0.8);
 }finally{bitmap.close()}
}
function editHumanPhoto(){
 const person=state.actors.find(a=>a.id===state.currentActorId);const enabled=state.featureFlags?.humanProfileImages!==false;let pending=null;
 form('Your profile photo','<p>Used in the squad and beside your chat messages in every room. Choose a JPEG or PNG, up to 5 MB. We resize it before saving.</p><label class="field">Photo<input type="file" id="human-photo-file" accept="image/jpeg,image/png"></label><div id="human-photo-preview"></div><label><input type="checkbox" name="remove"> Remove my photo</label><p class="muted">Photo revisions are kept in immutable history until normal retention and human-authorized purge. Removing a photo stops displaying it; it does not erase earlier records.</p>',async d=>{
 if(d.has('remove'))await commit('actor.image.updated',{image:''});
 else if(pending!==null)await commit('actor.image.updated',{image:pending});
 });
 const preview=$('#human-photo-preview');const display=value=>{preview.replaceChildren();if(value){const img=document.createElement('img');img.src=value;img.alt='Your profile photo preview';img.className='human-photo-preview';preview.append(img)}};display(humanPhoto(person));
 const fileInput=$('#human-photo-file');fileInput.disabled=!enabled;if(!enabled){preview.textContent='Profile photos are disabled. You can still remove your saved photo.'}const save=$('#modal button[type="submit"]')||$('#modal .primary');
 fileInput.onchange=async()=>{pending=null;save.disabled=true;try{if(fileInput.files[0]){pending=await resizeHumanPhoto(fileInput.files[0]);display(pending)}}catch(e){note(e.message);fileInput.value=''}finally{save.disabled=false}};
}
const photoForm=form;form=function(title,fields,save){
 photoForm(title,fields,save);
 if(['Your profile','Your room profile'].includes(title)&&(state.featureFlags?.humanProfileImages!==false||state.actors.find(a=>a.id===state.currentActorId)?.image)){
 const button=document.createElement('button');button.type='button';button.textContent='Change profile photo';button.onclick=()=>{$('#modal').close();editHumanPhoto()};$('#modal h2').after(button);
 }
};
const photoRender=render;render=function(){
 photoRender();const enabled=state.featureFlags?.humanProfileImages!==false;
 const apply=(element,actor)=>{const photo=humanPhoto(actor,enabled);if(!element||!photo)return;const img=document.createElement('img');img.src=photo;img.alt=actor.name+' profile photo';img.className='human-photo';element.replaceChildren(img)};
 document.querySelectorAll('[data-human-id]').forEach(row=>apply(row.querySelector('.avatar'),state.actors.find(a=>a.id===row.dataset.humanId)));
 document.querySelectorAll('.message').forEach((row,i)=>{const m=state.messages[i];if(m?.actorType==='human'||m?.author==='You'){const id=m.actorId|| (m.author==='You'?'local-owner':'');apply(row.querySelector('.avatar'),state.actors.find(a=>a.id===id))}});
 if(view==='Settings'&&state.currentActorId==='local-owner'){
 $('#edit-safety-flag')?.insertAdjacentHTML('afterend','<h3>Cosmetic</h3><p>Human profile photos · '+(enabled?'Enabled':'Hidden')+'</p><button id="edit-photo-flag">Configure profile photos</button>');
 $('#edit-photo-flag')?.insertAdjacentHTML('afterend','<h3>Connections</h3><p>Guided connection editor</p><p class="muted">Configured with COMMONROOM_CONNECTION_EDITOR; requires a server restart.</p><h3>Workflows</h3><p>Workflow templates</p><p class="muted">Configured with COMMONROOM_WORKFLOW_TEMPLATES; requires a server restart.</p>');
 if($('#edit-photo-flag'))$('#edit-photo-flag').onclick=()=>form('Human profile photos','<label><input type="checkbox" name="enabled" '+(enabled?'checked':'')+'> Allow profile photo uploads and display</label><p>Existing photos are retained when disabled. People can still remove their own photo. Only the owner can change this setting.</p>',d=>commit('feature.flag.updated',{name:'humanProfileImages',enabled:d.has('enabled')}));
 }
};
