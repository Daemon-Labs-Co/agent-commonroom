// Optional company authority editor. Persistence and authorization live on the server.
function organizationNodes(data){
 const known=new Set(data.actors.map(a=>a.id));const saved=new Map((data.organization?.nodes||[]).map(n=>[n.actorId,n]));
 return data.actors.map((a,index)=>{const n=saved.get(a.id);return {actorId:a.id,reportsTo:known.has(n?.reportsTo)?n.reportsTo:'',x:n?.x??40+(index%4)*250,y:n?.y??160+Math.floor(index/4)*150}});
}
function organizationWouldCycle(nodes,actor,parent){
 const parents=new Map(nodes.map(n=>[n.actorId,n.reportsTo]));parents.set(actor,parent);const seen=new Set();
 for(let cursor=actor;cursor;cursor=parents.get(cursor)){if(seen.has(cursor))return true;seen.add(cursor)}return false;
}
function arrangeOrganization(nodes){
 const map=new Map(nodes.map(n=>[n.actorId,n]));const rows=new Map();
 for(const node of nodes){let level=0,cursor=node.reportsTo;const seen=new Set([node.actorId]);while(map.has(cursor)&&!seen.has(cursor)){seen.add(cursor);level++;cursor=map.get(cursor).reportsTo}if(!rows.has(level))rows.set(level,[]);rows.get(level).push(node)}
 for(const [level,row] of rows)row.forEach((node,index)=>{node.x=40+index*250;node.y=160+level*150});return nodes;
}
function organizationNodeHTML(actor,node,selected){return '<button type="button" class="org-node '+(selected===actor.id?'selected':'')+'" data-org-node="'+esc(actor.id)+'" aria-label="'+esc(actor.name)+' · '+esc(actor.kind)+'"><span class="org-node-kind">'+(actor.kind==='human'?'Human':'Agent')+'</span><strong>'+esc(actor.name)+'</strong><small>'+esc(actor.title||actor.role)+(actor.enabled?'':' · Disabled')+'</small><i class="org-port" aria-hidden="true"></i></button>'}
// Reuse the existing authorized editors; chart navigation grants no new access.
function editOrganizationIdentity(id){
 if(state.currentActorId!=='local-owner'||state.featureFlags?.companyOrgChart===false)return false;
 const person=state.actors?.find(a=>a.id===id);if(!person){note('This identity is no longer available. Reopen the chart.');return false}
 if(person.kind==='agent'){
  const c=state.connections?.find(c=>'bot-'+c.id===id);if(!c){note('This agent no longer has a connection.');return false}connection(c);
 }else if(person.kind==='human'){
  if(id==='local-owner'){
   const edit=$('#edit-human-name');if(!edit){note('Open the chart from Settings to edit your profile.');return false}edit.onclick();
  }else memberAccessForm(person);
 }else return false;
 return true;
}
function wireOrganizationIdentity(button,id,select,edit){button.onclick=()=>select(id);button.ondblclick=()=>edit(id)}
async function openOrganizationChart(){
 if(state.currentActorId!=='local-owner'||state.featureFlags?.companyOrgChart===false)return;
 let data;try{data=await api('company/organization')}catch(e){note(e.message);return}
 const dialog=document.createElement('dialog');dialog.className='organization-dialog';dialog.setAttribute('aria-label','Company organization chart');
 dialog.innerHTML='<div class="org-header"><div><h2>'+esc(data.company.name)+' · Organization</h2><p class="muted">Double-click an identity to edit it. Reporting lines replace numbered seniority. Humans retain final override.</p></div><div class="actions"><button type="button" id="org-save" class="primary">Save chart</button><button type="button" id="org-help" class="help-icon" aria-label="Organization chart help">?</button><button type="button" class="help-close" aria-label="Close organization chart">×</button></div></div><div class="org-tools"><button type="button" id="org-arrange">Auto-arrange</button><button type="button" id="org-zoom-out" aria-label="Zoom out">−</button><output id="org-zoom">100%</output><button type="button" id="org-zoom-in" aria-label="Zoom in">+</button><button type="button" id="org-mode">List view</button><span id="org-status" role="status">No unsaved changes</span></div><div class="org-editor"><aside class="org-roster" aria-label="Company identities"></aside><div class="org-viewport"><div class="org-world"><svg class="org-lines" aria-hidden="true"></svg><div class="org-company">'+esc(data.company.name)+'</div><div class="org-nodes"></div></div><div class="org-list" hidden></div></div><aside class="org-inspector" aria-label="Identity properties"></aside></div>';
 document.body.append(dialog);dialog.showModal();let nodes=organizationNodes(data),selected=data.actors[0]?.id,dirty=false,zoom=1,listMode=false,drag=null,saving=false;
 const q=s=>dialog.querySelector(s);const status=q('#org-status');
 q('#org-save').disabled=true;
 const changed=()=>{dirty=true;q('#org-save').disabled=false;status.textContent='Unsaved changes'};
 function actor(id){return data.actors.find(a=>a.id===id)}
 function edges(){const svg=q('.org-lines');const width=Math.max(1100,...nodes.map(n=>n.x+260)),height=Math.max(600,...nodes.map(n=>n.y+150));q('.org-world').style.width=width+'px';q('.org-world').style.height=height+'px';svg.setAttribute('width',width);svg.setAttribute('height',height);svg.innerHTML=nodes.map(n=>{const parent=nodes.find(p=>p.actorId===n.reportsTo),x=parent?parent.x+110:550,y=parent?parent.y+105:82;return '<path d="M '+x+' '+y+' C '+x+' '+(y+30)+', '+(n.x+110)+' '+(n.y-30)+', '+(n.x+110)+' '+n.y+'"/>'}).join('')}
 function select(id){selected=id;dialog.querySelectorAll('[data-org-node]').forEach(b=>b.classList.toggle('selected',b.dataset.orgNode===id));drawInspector();if(listMode){[...q('.org-list').querySelectorAll('[data-org-select]')].find(b=>b.dataset.orgSelect===id)?.focus()}else q('.org-node.selected')?.focus()}
 function draw(){
  q('.org-roster').innerHTML=['human','agent'].map(kind=>'<h3>'+ (kind==='human'?'Humans':'Agents')+'</h3>'+data.actors.filter(a=>a.kind===kind).map(a=>'<button type="button" data-org-select="'+esc(a.id)+'">'+esc(a.name)+'</button>').join('')).join('');
  q('.org-nodes').innerHTML=data.actors.map(a=>organizationNodeHTML(a,nodes.find(n=>n.actorId===a.id),selected)).join('');edges();
  q('.org-list').innerHTML='<table><thead><tr><th>Identity</th><th>Type</th><th>Reports to</th></tr></thead><tbody>'+data.actors.map(a=>{const n=nodes.find(n=>n.actorId===a.id);return '<tr><td><button type="button" data-org-select="'+esc(a.id)+'">'+esc(a.name)+'</button></td><td>'+esc(a.kind)+'</td><td>'+esc(actor(n.reportsTo)?.name||data.company.name)+'</td></tr>'}).join('')+'</tbody></table>';
  drawInspector();
  dialog.querySelectorAll('[data-org-select]').forEach(b=>{wireOrganizationIdentity(b,b.dataset.orgSelect,select,edit)});
  dialog.querySelectorAll('[data-org-node]').forEach(b=>{
   const position=nodes.find(n=>n.actorId===b.dataset.orgNode);b.style.left=position.x+'px';b.style.top=position.y+'px';
   wireOrganizationIdentity(b,b.dataset.orgNode,select,edit);
   b.onkeydown=e=>{const delta={ArrowLeft:[-10,0],ArrowRight:[10,0],ArrowUp:[0,-10],ArrowDown:[0,10]}[e.key];if(!delta||saving)return;e.preventDefault();const node=nodes.find(n=>n.actorId===b.dataset.orgNode);const step=e.shiftKey?5:1;node.x=Math.max(0,Math.min(50000,node.x+delta[0]*step));node.y=Math.max(0,Math.min(50000,node.y+delta[1]*step));selected=node.actorId;changed();draw();q('.org-node.selected').focus()};
   b.onpointerdown=e=>{if(e.button!==0||saving)return;const node=nodes.find(n=>n.actorId===b.dataset.orgNode);selected=node.actorId;drag={node,startX:e.clientX,startY:e.clientY,x:node.x,y:node.y,moved:false};b.setPointerCapture(e.pointerId)};
  });
 }
 function edit(id){if(saving){status.textContent='Wait for saving to finish';return}if(!editOrganizationIdentity(id))return;$('#modal').addEventListener('close',async()=>{try{const fresh=await api('company/organization');data.actors=fresh.actors;data.profiles=fresh.profiles;draw()}catch(e){status.textContent=e.message}},{once:true})}
 function drawInspector(){
  const a=actor(selected),n=nodes.find(n=>n.actorId===selected);q('.org-inspector').innerHTML=a?'<h3>'+esc(a.name)+'</h3><p>'+esc(a.title||a.role)+' · '+esc(a.kind)+'</p><button type="button" id="org-edit">'+(a.kind==='human'?'Edit profile':'Edit connection')+'</button><label class="field">Reports to<select id="org-manager"><option value="">Company · no manager</option>'+data.actors.filter(other=>other.id!==a.id).map(other=>'<option value="'+esc(other.id)+'" '+(n.reportsTo===other.id?'selected':'')+'>'+esc(other.name)+'</option>').join('')+'</select></label><p class="muted">Managers outrank their reports. Peers must reach consensus. Reporting lines do not grant room or tool access.</p><h3>Room profiles</h3>'+data.profiles.filter(p=>p.actorId===a.id).map(p=>'<p><strong>'+esc(p.name)+'</strong><br>'+esc(p.role)+'<br><small>'+esc(p.room)+'</small></p>').join('')+(a.kind==='human'?'<p class="muted">Human account roles and room access remain owner-managed.</p>':'<p class="muted">One connection identity may have different profiles in different rooms.</p>'):'<p>No company identities.</p>';
  if(a)q('#org-edit').onclick=()=>edit(a.id);
  if(a)q('#org-manager').onchange=e=>{if(saving)return;if(organizationWouldCycle(nodes,a.id,e.target.value)){status.textContent='That reporting line would create a cycle';e.target.value=n.reportsTo;return}n.reportsTo=e.target.value;changed();draw()};
 }
 q('.org-viewport').onpointermove=e=>{if(!drag)return;if(!drag.moved&&Math.hypot(e.clientX-drag.startX,e.clientY-drag.startY)<3)return;drag.moved=true;drag.node.x=Math.max(0,Math.min(50000,drag.x+(e.clientX-drag.startX)/zoom));drag.node.y=Math.max(0,Math.min(50000,drag.y+(e.clientY-drag.startY)/zoom));const b=[...dialog.querySelectorAll('[data-org-node]')].find(b=>b.dataset.orgNode===drag.node.actorId);b.style.left=drag.node.x+'px';b.style.top=drag.node.y+'px';edges();changed()};
 q('.org-viewport').onpointerup=q('.org-viewport').onpointercancel=()=>{if(drag){const moved=drag.moved;drag=null;if(moved)draw()}};
 q('#org-arrange').onclick=()=>{if(saving)return;arrangeOrganization(nodes);changed();draw()};
 function setZoom(value){zoom=Math.max(.5,Math.min(1.5,value));q('.org-world').style.zoom=zoom;q('#org-zoom').textContent=Math.round(zoom*100)+'%'}
 q('#org-zoom-out').onclick=()=>setZoom(zoom-.1);q('#org-zoom-in').onclick=()=>setZoom(zoom+.1);
 q('#org-mode').onclick=()=>{listMode=!listMode;q('.org-world').hidden=listMode;q('.org-list').hidden=!listMode;q('#org-mode').textContent=listMode?'Canvas view':'List view'};
 q('#org-save').onclick=async()=>{if(saving||!dirty)return;saving=true;q('#org-save').disabled=true;try{await api('events',{kind:'company.organization.updated',payload:{revision:data.organization.revision,nodes}});data.organization.revision++;dirty=false;status.textContent='Chart saved to immutable history'}catch(e){status.textContent=e.message}finally{saving=false;q('#org-save').disabled=!dirty;draw()}};
 q('#org-help').onclick=()=>openDocumentation('concepts','company-organization-and-decision-authority');
 const close=()=>{if(saving){status.textContent='Wait for saving to finish';return}if(dirty&&!window.confirm('Discard unsaved organization changes?'))return;dialog.close()};q('.help-close').onclick=close;dialog.addEventListener('cancel',e=>{e.preventDefault();close()});dialog.addEventListener('close',()=>dialog.remove(),{once:true});draw();
}
const organizationRender=render;render=function(){
 organizationRender();if(!state||state.featureFlags?.companyOrgChart===false)return;
 document.querySelectorAll('.participant-info>span').forEach(n=>{n.textContent=n.textContent.replace(/\s*·?\s*Seniority \d+/g,'')});
 if(view==='Settings'&&state.currentActorId==='local-owner'){const card=[...document.querySelectorAll('#main .card')].find(c=>c.querySelector('h2')?.textContent==='Company');if(card){const button=document.createElement('button');button.type='button';button.textContent='Organization chart';button.onclick=openOrganizationChart;card.append(button)}}
};
const organizationForm=form;form=function(title,fields,save){organizationForm(title,fields,save);if(state?.featureFlags?.companyOrgChart===false)return;
 for(const name of ['rank','humanRank']){const input=$('#modal input[name="'+name+'"]');if(!input)continue;const hidden=document.createElement('input');hidden.type='hidden';hidden.name=input.name;hidden.value=input.value;const label=input.closest('label.field');label.before(hidden);const row=label.parentElement;if(row.classList.contains('field-guidance-row'))row.remove();else label.remove()}
};
