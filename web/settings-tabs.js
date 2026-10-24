// Settings tabs reorganize existing controls; they do not change capabilities.
function settingsCardCategory(title){
 if(/^Your profile|^Your room profile/.test(title))return 'Profile';
 if(/^Company/.test(title))return 'Company';
 if(title==='Agent rooms'||title==='Feature flags'||title==='Room presentation'||/^(Security|Cosmetic|Connections|Workflows) ·/.test(title))return 'Features';
 if(title==='Retention')return 'Retention';
 if(title==='Help display')return 'Help';
 return 'About';
}
let activeSettingsCategory='Profile';
function tabbedSettings(){
 if(state?.featureFlags?.settingsTabs===false)return;
 const history=document.querySelector('[data-global="History & retention"]');if(history)history.textContent='History';
 if(view==='History & retention'){
  const heading=$('#main h1');if(heading)heading.textContent='History';
  const notice=$('#audit')?.parentElement.querySelector(':scope > p');if(notice)notice.remove();return;
 }
 if(view!=='Settings')return;
 const main=$('#main');
 let toolbar=main.querySelector(':scope > .toolbar');if(!toolbar){toolbar=document.createElement('div');toolbar.className='toolbar';toolbar.innerHTML='<h1>Settings</h1>'}main.prepend(toolbar);
 if(state.currentActorId==='local-owner')main.insertAdjacentHTML('beforeend','<section class="card"><h2>Retention</h2><p>Minimum retention: 30 days.</p><p>Record disposal and purge remain disabled until storage-enforced retention and protected deletion receipts are implemented. Purge will require a human.</p><p class="muted">Moving this policy into Settings does not change the immutable log or allow deletion.</p></section>');
 const cards=[...main.querySelectorAll(':scope > .card')];if(!cards.length)return;
 const groups=['Profile','Company','Features','Retention','Help','About'].filter(group=>cards.some(card=>settingsCardCategory(card.querySelector(':scope > h2,:scope > h3,:scope > h1')?.textContent||'')===group));
 const nav=document.createElement('div');nav.className='settings-tabs';nav.setAttribute('role','tablist');nav.setAttribute('aria-label','Settings sections');
 const bar=document.createElement('div');bar.className='settings-tabbar';bar.append(nav);
 const help=main.querySelector(':scope > .context-help, :scope > .toolbar .context-help');if(help){help.classList.add('settings-help');bar.append(help)}
 const container=document.createElement('div');container.className='settings-panels';cards[0].before(bar,container);
 const panels={},buttons={};groups.forEach((group,index)=>{
  const panel=document.createElement('section');panel.id='settings-panel-'+index;panel.setAttribute('role','tabpanel');panel.setAttribute('aria-labelledby','settings-tab-'+index);panels[group]=panel;container.append(panel);
  const button=document.createElement('button');button.type='button';button.textContent=group;button.id='settings-tab-'+index;button.setAttribute('role','tab');button.setAttribute('aria-controls',panel.id);button.onclick=()=>activate(group);button.onkeydown=e=>{let next=index;if(e.key==='ArrowRight')next=(index+1)%groups.length;else if(e.key==='ArrowLeft')next=(index+groups.length-1)%groups.length;else if(e.key==='Home')next=0;else if(e.key==='End')next=groups.length-1;else return;e.preventDefault();activate(groups[next]);buttons[groups[next]].focus()};buttons[group]=button;nav.append(button);
 });
 function activate(group){activeSettingsCategory=group;groups.forEach(g=>{panels[g].hidden=g!==group;buttons[g].tabIndex=g===group?0:-1;buttons[g].setAttribute('aria-selected',String(g===group))})}
 cards.forEach(card=>panels[settingsCardCategory(card.querySelector(':scope > h2,:scope > h3,:scope > h1')?.textContent||'')].append(card));
 activate(groups.includes(activeSettingsCategory)?activeSettingsCategory:groups[0]);
}
const settingsTabsRender=render;render=function(){settingsTabsRender();tabbedSettings()};
