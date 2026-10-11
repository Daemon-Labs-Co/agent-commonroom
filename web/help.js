"use strict";
function openDocumentation(topic='index',section=''){
 let dialog=document.querySelector('#help-dialog');
 if(!dialog){dialog=document.createElement('dialog');dialog.id='help-dialog';dialog.className='help-dialog';dialog.setAttribute('aria-label','Help and documentation');dialog.innerHTML='<div class="help-dialog-header"><strong>Help &amp; documentation</strong><button type="button" aria-label="Close help" class="help-close">×</button></div><iframe title="Agent CommonRoom documentation"></iframe>';document.body.append(dialog);dialog.querySelector('.help-close').onclick=()=>dialog.close();}
 dialog.querySelector('iframe').src='/help/?doc='+encodeURIComponent(topic)+(section?'#'+encodeURIComponent(section):'');if(!dialog.open)dialog.showModal();
}
function helpButton(topic,section,label='Help'){const b=document.createElement('button');b.type='button';b.className='context-help';b.textContent=label;b.setAttribute('aria-label',label);b.onclick=()=>openDocumentation(topic,section);return b}
const helpRender=render;render=function(){helpRender();const settings=document.querySelector('#nav .nav-settings');if(settings){const button=helpButton('index','','Help & documentation');button.classList.add('nav-help');settings.before(button)}
 const guides={'Room':['concepts','messages-and-context'],'Connections':['installation','connect-agents'],'Tasks':['concepts','tasks-evidence-and-workflows'],'Workflows':['concepts','tasks-evidence-and-workflows'],'Shared memory':['concepts','shared-memory'],'History & retention':['operations','integrity-and-history'],'Settings':['concepts','company-humans-and-rooms']};
 const entry=guides[view];if(entry){const bar=document.querySelector('.wf-toolbar .actions, #main .toolbar .actions');const button=helpButton(...entry);if(bar)bar.append(button);else document.querySelector('#main').prepend(button)}
};
const helpForm=form;form=function(title,fields,save){helpForm(title,fields,save);let section='';if(/connection|plug in agent/i.test(title))section='connections-versus-participants';else if(/memory/i.test(title))section='shared-memory';else if(/profile|agent/i.test(title))section='profiles-and-challenge-levels';else if(/task|evidence|reviewer/i.test(title))section='tasks-evidence-and-workflows';else if(/room|member|human|company/i.test(title))section='company-humans-and-rooms';if(section){const heading=document.querySelector('#modal h2');heading.after(helpButton('concepts',section,'Help with this form'))}};
const headerHelp=helpButton('index','','Help');document.querySelector('header .environment').append(headerHelp);
