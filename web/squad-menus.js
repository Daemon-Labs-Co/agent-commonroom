// Preserve existing action handlers and permissions while bounding one menu.
let activeSquadMenu=null;
function squadMenuPosition(anchor,box,width,height){
 const margin=8,gap=6;
 const left=Math.max(margin,Math.min(anchor.right-box.width,width-box.width-margin));
 const below=anchor.bottom+gap;
 const top=below+box.height<=height-margin?below:Math.max(margin,anchor.top-box.height-gap);
 return {left,top};
}
function closeSquadMenu(returnFocus=false){
 const active=activeSquadMenu;if(!active)return;activeSquadMenu=null;active.details.removeAttribute('open');active.summary.setAttribute('aria-expanded','false');
 active.panel.classList.remove('squad-menu-floating');
 for(const key of ['left','right','top','maxHeight','maxWidth','visibility'])active.panel.style[key]='';
 if(returnFocus&&active.summary.isConnected)active.summary.focus();
}
function openSquadMenu(details,summary,panel,focusFirst=false){
 if(activeSquadMenu?.details===details){closeSquadMenu();return}
 closeSquadMenu();details.open=true;summary.setAttribute('aria-expanded','true');panel.classList.add('squad-menu-floating');panel.style.visibility='hidden';panel.style.maxHeight=Math.max(80,window.innerHeight-16)+'px';panel.style.maxWidth=Math.max(80,window.innerWidth-16)+'px';
 const position=squadMenuPosition(summary.getBoundingClientRect(),panel.getBoundingClientRect(),window.innerWidth,window.innerHeight);
 panel.style.left=position.left+'px';panel.style.right='auto';panel.style.top=position.top+'px';panel.style.visibility='visible';activeSquadMenu={details,summary,panel};
 if(focusFirst)panel.querySelector('button:not(:disabled)')?.focus();
}
function drawSquadMenus(){
 closeSquadMenu();const enabled=view==='Room'&&state?.featureFlags?.squadMenuLayout!==false;document.body.classList.toggle('squad-menu-layout',enabled);if(!enabled)return;
 for(const details of document.querySelectorAll('.participant-menu')){
  const summary=details.querySelector('summary'),panel=details.querySelector('.menu-panel');if(!summary||!panel)continue;
  details.removeAttribute('open');summary.setAttribute('aria-expanded','false');
  summary.onclick=e=>{e.preventDefault();openSquadMenu(details,summary,panel)};
  summary.onkeydown=e=>{if(e.key==='ArrowDown'){e.preventDefault();openSquadMenu(details,summary,panel,true)}};
  panel.addEventListener('click',e=>{if(e.target.closest('button'))closeSquadMenu()});
 }
}
window.addEventListener('keydown',e=>{if(e.key==='Escape'&&activeSquadMenu){e.preventDefault();closeSquadMenu(true)}});
document.addEventListener('pointerdown',e=>{if(activeSquadMenu&&!activeSquadMenu.details.contains(e.target))closeSquadMenu()});
window.addEventListener('scroll',e=>{if(activeSquadMenu&&!activeSquadMenu.panel.contains(e.target))closeSquadMenu()},true);
window.addEventListener('resize',()=>closeSquadMenu());
const squadMenusRender=render;render=function(){squadMenusRender();drawSquadMenus()};
