// Reparent existing controls so FormData and existing permission checks remain intact.
function overlayFieldSection(title,node){
 if(node.classList?.contains('human-account-details'))return 'Account & rooms';
 if(title==='Edit profile'){
  if(node.querySelector?.('[name="discipline"]'))return 'Role';
  if(node.querySelector?.('[name="additionalContext"]'))return 'Context';
  if(node.querySelector?.('[name="instructions"],[name="rank"],[name="accept"]'))return 'Behaviour';
  return 'Agent';
 }
 return 'Profile';
}
function tabProfileOverlay(title,modal){
 if(state?.featureFlags?.compactProfileOverlays===false||!['Edit profile','Your profile','Your room profile','Human profile'].includes(title))return;
 const form=modal.querySelector('form');if(!form)return;
 const fields=[...form.children].filter(node=>!node.classList.contains('actions'));
 const groups=[...new Set(fields.map(node=>overlayFieldSection(title,node)))];if(groups.length<2)return;
 modal.classList.add('compact-profile');modal.addEventListener('close',()=>modal.classList.remove('compact-profile'),{once:true});
 const nav=document.createElement('div');nav.className='profile-overlay-tabs';nav.setAttribute('role','tablist');nav.setAttribute('aria-label','Profile sections');form.prepend(nav);
 const panels={},buttons={};
 groups.forEach((group,index)=>{
  const panel=document.createElement('section');panel.id='profile-panel-'+index;panel.className='profile-overlay-panel';panel.setAttribute('role','tabpanel');panel.setAttribute('aria-labelledby','profile-tab-'+index);panels[group]=panel;nav.after(panel);
  const button=document.createElement('button');button.type='button';button.id='profile-tab-'+index;button.textContent=group;button.setAttribute('role','tab');button.setAttribute('aria-controls',panel.id);button.onclick=()=>activate(group);button.onkeydown=e=>{let next=index;if(e.key==='ArrowRight')next=(index+1)%groups.length;else if(e.key==='ArrowLeft')next=(index+groups.length-1)%groups.length;else if(e.key==='Home')next=0;else if(e.key==='End')next=groups.length-1;else return;e.preventDefault();activate(groups[next]);buttons[groups[next]].focus()};buttons[group]=button;nav.append(button);
 });
 function activate(group){groups.forEach(key=>{panels[key].hidden=key!==group;buttons[key].tabIndex=key===group?0:-1;buttons[key].setAttribute('aria-selected',String(key===group))})}
 fields.forEach(node=>{
  const group=overlayFieldSection(title,node);
  // Expand engineering fields and starting context without requiring another click.
  if(group==='Role'&&node.tagName==='DETAILS'){
   const container=document.createElement('div');container.className='profile-role-grid';[...node.children].filter(child=>child.tagName!=='SUMMARY').forEach(child=>container.append(child));panels[group].append(container);node.remove();
  }else{if(group==='Context'&&node.tagName==='DETAILS')node.open=true;panels[group].append(node)}
 });
 activate(groups[0]);form.addEventListener('invalid',e=>{const panel=e.target.closest('[role="tabpanel"]');const group=groups.find(key=>panels[key]===panel);if(group)activate(group)},true);
}
const compactProfileForm=form;form=function(title,fields,save){compactProfileForm(title,fields,save);tabProfileOverlay(title,$('#modal'))};
if(typeof openSquadHumanProfile==='function'){
 const compactHumanOpen=openSquadHumanProfile;openSquadHumanProfile=function(...args){compactHumanOpen(...args);const modal=$('#modal');if(!modal.querySelector('.profile-overlay-tabs'))tabProfileOverlay('Human profile',modal)};
}
