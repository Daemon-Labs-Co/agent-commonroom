"use strict";
function helpWindowChoiceEnabled(){return typeof state!=='undefined'&&state?.featureFlags?.helpWindowChoice!==false;}
function helpDisplayMode(){if(!helpWindowChoiceEnabled())return 'same-window';return state.actors?.find(a=>a.id===state.currentActorId)?.helpDisplay==='separate-window'?'separate-window':'same-window';}
function helpURL(topic,section){return '/help/?doc='+encodeURIComponent(topic)+(section?'#'+encodeURIComponent(section):'');}
function openHelpWindow(url){const popup=window.open(url,'commonroom-help','popup,width=1100,height=800');if(popup){try{popup.opener=null;}catch(e){}return true;}note('The browser blocked the help window. Allow pop-ups for CommonRoom, or use same-window help.');return false;}
function openDocumentation(topic='index',section='',forceSame=false){
 const url=helpURL(topic,section);
 if(!forceSame&&helpDisplayMode()==='separate-window'&&openHelpWindow(url))return;
 let dialog=document.querySelector('#help-dialog');
 if(!dialog){dialog=document.createElement('dialog');dialog.id='help-dialog';dialog.className='help-dialog';dialog.setAttribute('aria-label','Help and documentation');dialog.innerHTML='<div class="help-dialog-header"><strong>Help &amp; documentation</strong><label class="help-display-control">Open help in<select aria-label="Help window preference"><option value="same-window">Same window</option><option value="separate-window">Separate window</option></select></label><button type="button" aria-label="Close help" class="help-close">×</button></div><iframe title="Agent CommonRoom documentation"></iframe>';document.body.append(dialog);dialog.querySelector('.help-close').onclick=()=>dialog.close();}
 const control=dialog.querySelector('.help-display-control'),select=control.querySelector('select');control.hidden=!helpWindowChoiceEnabled();select.value=helpDisplayMode();
 select.onchange=()=>{const mode=select.value;const opened=mode==='separate-window'&&openHelpWindow(url);commit('actor.help.updated',{mode}).then(()=>{if(opened)dialog.close();}).catch(e=>{note(e.message);select.value=helpDisplayMode();});};
 dialog.querySelector('iframe').src=url;if(!dialog.open)dialog.showModal();
}
function helpButton(topic,section,label='Help'){const b=document.createElement('button');b.type='button';b.className='context-help';b.textContent=label;b.setAttribute('aria-label',label);b.onclick=()=>openDocumentation(topic,section);return b}
const helpRender=render;render=function(){helpRender();const settings=document.querySelector('#nav .nav-settings');if(settings){const button=helpButton('index','','Help & documentation');button.classList.add('nav-help');settings.before(button)}
 const guides={'Room':['concepts','messages-and-context'],'Connections':['installation','connect-agents'],'Tasks':['concepts','tasks-evidence-and-workflows'],'Workflow templates':['concepts','workflow-templates'],'Workflows':['workflow-execution',''],'Shared memory':['concepts','shared-memory'],'History & retention':['operations','integrity-and-history'],'Settings':['concepts','company-humans-and-rooms']};
 const entry=guides[view];if(entry){const bar=document.querySelector('.wf-toolbar .actions, #main .toolbar .actions');const button=helpButton(...entry);if(bar)bar.append(button);else document.querySelector('#main').prepend(button)}
};
const helpForm=form;form=function(title,fields,save){helpForm(title,fields,save);let section='';if(/workflow|interrupted work/i.test(title)){const heading=document.querySelector('#modal h2');heading.after(helpButton('workflow-execution','','Workflow help'));return}if(title==='Invite agent')section='squad-profiles-and-invitations';else if(/stop|resume|emergency agent controls/i.test(title))section='emergency-agent-cutoff';else if(/template/i.test(title))section='workflow-templates';else if(/connection|plug in agent/i.test(title))section='connections-versus-participants';else if(/memory/i.test(title))section='shared-memory';else if(/profile|agent/i.test(title))section='profiles-and-challenge-levels';else if(/task|evidence|reviewer/i.test(title))section='tasks-evidence-and-workflows';else if(/room|member|human|company/i.test(title))section='company-humans-and-rooms';if(section){const heading=document.querySelector('#modal h2');heading.after(helpButton('concepts',section,'Help with this form'))}};
const headerHelp=helpButton('index','','Help');document.querySelector('header .environment').append(headerHelp);

const helpPreferenceRender=render;render=function(){helpPreferenceRender();if(view!=='Settings')return;
 if(helpWindowChoiceEnabled()){
 const card=document.createElement('section');card.className='card';card.innerHTML='<h2>Help display</h2><p>Choose where help opens. This preference belongs to your human profile.</p><label class="field">Open help in<select id="settings-help-display"><option value="same-window">Same window</option><option value="separate-window">Separate window</option></select></label><p class="muted">Your browser may open a tab instead of a window. If pop-ups are blocked, help opens here.</p>';
 document.querySelector('#main').append(card);const select=card.querySelector('select');select.value=helpDisplayMode();select.onchange=()=>commit('actor.help.updated',{mode:select.value}).catch(e=>note(e.message));
 }
 if(state.currentActorId==='local-owner'){
 const card=document.createElement('section');card.className='card';card.innerHTML='<h3>Cosmetic · Help window choice</h3><p>Let humans choose same-window or separate-window help.</p><button>Toggle help window choice</button>';card.querySelector('button').dataset.featureFlag='helpWindowChoice';card.querySelector('button').onclick=()=>commit('feature.flag.updated',{name:'helpWindowChoice',enabled:state.featureFlags?.helpWindowChoice===false});document.querySelector('#main').append(card);
 }
};
