// Presentation only: gather the room's agent-team controls under one "Team"
// button so the toolbar stays readable. The controls keep their own handlers.
// "Stop discussion" stays outside the menu so it is always one click away.
const teamMenuIds=['agent-discussion','agent-discussion-limit','suggest-tasks','manager-assignment-limit'];
function teamMenuItems(buttons){return buttons.filter(b=>teamMenuIds.includes(b.id)&&!(b.id==='agent-discussion'&&b.textContent==='Stop discussion'))}
function drawTeamMenu(){
 if(view!=='Room')return;
 const actions=document.querySelector('.toolbar .actions');if(!actions||actions.querySelector('.team-menu'))return;
 const items=teamMenuItems([...actions.querySelectorAll('button')]);if(items.length<2)return;
 const menu=document.createElement('details');menu.className='team-menu';
 const summary=document.createElement('summary');summary.textContent='Team';summary.title='Agent team controls for this room';menu.append(summary);
 const panel=document.createElement('div');panel.className='team-menu-items menu-panel';for(const b of items)panel.append(b);menu.append(panel);
 panel.addEventListener('click',()=>{menu.open=false});
 actions.append(menu);
}
const teamMenuRender=render;render=function(){teamMenuRender();drawTeamMenu()};
