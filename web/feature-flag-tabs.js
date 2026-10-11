// Group existing controls without replacing their handlers or permissions.
function featureFlagCategory(title){
 if(title==='Room presentation')return 'Cosmetic';
 return ['Security','Cosmetic','Connections','Workflows'].find(group=>title.startsWith(group+' ·'))||null;
}
let activeFeatureCategory='Security';
function groupFeatureFlags(){
 if(view!=='Settings'||state.currentActorId!=='local-owner'||state.featureFlags?.featureFlagTabs===false)return;
 const main=$('#main');const cards=[...main.querySelectorAll(':scope > .card')];
 const base=cards.find(card=>card.querySelector(':scope > h2')?.textContent==='Feature flags');if(!base)return;
 const section=document.createElement('section');section.className='card feature-flags';section.innerHTML='<h2>Feature flags</h2><p class="muted">Choose a category. These settings retain their existing permissions and safety checks.</p>';
 const tabs=document.createElement('div');tabs.className='invitation-tabs feature-tabs';tabs.setAttribute('role','tablist');tabs.setAttribute('aria-label','Feature categories');section.append(tabs);
 const groups=['Security','Cosmetic','Connections','Workflows'];const panels={},buttons={};
 groups.forEach((group,index)=>{
  const panel=document.createElement('div');panel.id='feature-panel-'+index;panel.setAttribute('role','tabpanel');panel.setAttribute('aria-labelledby','feature-tab-'+index);panels[group]=panel;section.append(panel);
  const button=document.createElement('button');button.type='button';button.textContent=group;button.id='feature-tab-'+index;button.setAttribute('role','tab');button.setAttribute('aria-controls',panel.id);button.onclick=()=>activate(group);button.onkeydown=e=>{let next=index;if(e.key==='ArrowRight')next=(index+1)%groups.length;else if(e.key==='ArrowLeft')next=(index+groups.length-1)%groups.length;else if(e.key==='Home')next=0;else if(e.key==='End')next=groups.length-1;else return;e.preventDefault();activate(groups[next]);buttons[groups[next]].focus()};buttons[group]=button;tabs.append(button);
 });
 function activate(group){activeFeatureCategory=group;groups.forEach(g=>{panels[g].hidden=g!==group;buttons[g].tabIndex=g===group?0:-1;buttons[g].setAttribute('aria-selected',String(g===group))})}
 let group='Security';for(const node of [...base.children]){
  if(node.tagName==='H2')continue;
  if(node.tagName==='H3'&&groups.includes(node.textContent)){group=node.textContent;continue}
  panels[group].append(node);
 }
 base.replaceWith(section);
 for(const card of cards){if(card===base)continue;const heading=card.querySelector(':scope > h2,:scope > h3');const category=featureFlagCategory(heading?.textContent||'');if(category){card.classList.add('feature-flag-item');panels[category].append(card)}}
 activate(groups.includes(activeFeatureCategory)?activeFeatureCategory:'Security');
}
const featureTabsRender=render;render=function(){featureTabsRender();groupFeatureFlags()};
