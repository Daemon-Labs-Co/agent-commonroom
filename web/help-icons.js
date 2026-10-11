// Cosmetic only; destinations and the person's help-window preference stay intact.
function standardHelpIcon(button){
 if(!button||button.classList.contains('nav-help'))return;
 const label=button.getAttribute('aria-label')||button.textContent||'Help';
 button.textContent='?';button.classList.add('help-icon');button.setAttribute('aria-label',label);button.title=label;
}
function rightAlignFieldGuidance(modal){
 if(!modal)return;
 modal.querySelectorAll('label.field').forEach(field=>{
  if(field.parentElement?.classList.contains('field-guidance-row'))return;
  const info=field.nextElementSibling;
  if(!info||info.tagName!=='P'||!info.classList.contains('muted'))return;
  const row=document.createElement('div');row.className='field-guidance-row';field.before(row);row.append(field,info);info.classList.add('field-guidance');
 });
}
function standardizeHelpIcons(){
 if(state?.featureFlags?.standardHelpIcons===false)return;
 document.querySelectorAll('.context-help,.seniority-help,#connection-ssh-help').forEach(standardHelpIcon);
 const modal=document.querySelector('#modal');const help=modal?.querySelector(':scope > .context-help');
 if(help){help.classList.add('overlay-help');modal.classList.add('has-overlay-help');}
 rightAlignFieldGuidance(modal);
}
const iconHelpForm=form;
form=function(title,fields,save){iconHelpForm(title,fields,save);standardizeHelpIcons()};
const iconHelpRender=render;
render=function(){iconHelpRender();standardizeHelpIcons()};
standardizeHelpIcons();
