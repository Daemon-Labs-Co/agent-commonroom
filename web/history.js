// Preserve the selected period across background rerenders, but reset on a new
// visit. Bounded cached pages remain scoped to this signed-in History view.
let auditHistorySession=null;
const auditHistoryBaseRender=render;
render=function(){if(view!=='History & retention')auditHistorySession=null;auditHistoryBaseRender()};
// Read-only, bounded history pages. Opening this view never downloads all events.
function auditHistoryWindow(period,now=new Date()){
 const today=new Date(now.getFullYear(),now.getMonth(),now.getDate());
 const day=n=>{const d=new Date(today);d.setDate(d.getDate()+n);return d};
 const week=day(-((today.getDay()+6)%7)),lastWeek=new Date(week);lastWeek.setDate(lastWeek.getDate()-7);
 let from,until;
 if(period==='today'){from=today;until=day(1)}
 else if(period==='yesterday'){from=day(-1);until=today}
 else if(period==='week'){from=week;until=day(-1)}
 else if(period==='last-week'){from=lastWeek;until=new Date(Math.min(+week,+day(-1)))}
 else if(period==='older'){until=lastWeek}
 else if(period.startsWith('month-')){
  const parts=period.slice(6).split('-').map(Number);from=new Date(parts[0],parts[1],1);until=new Date(Math.min(+new Date(parts[0],parts[1]+1,1),+lastWeek));
 }
 return {from:from?.toISOString(),until:until?.toISOString(),empty:!!(from&&until&&from>=until)};
}
function auditHistoryPageURL(period,before,now=new Date()){
 const range=auditHistoryWindow(period,now),params=new URLSearchParams({page:'true'});
 if(range.from)params.set('from',range.from);if(range.until)params.set('until',range.until);if(before)params.set('before',String(before));
 return 'events?'+params.toString();
}
// Stale tab/view responses never replace current content. Each request is one
// fixed server page; Load more advances its exclusive sequence cursor.
async function loadAuditHistory(){
 const target=document.querySelector('#audit');if(!target)return;
 const grouped=state?.featureFlags?.historyDateGroups!==false&&typeof auditDateBucket==='function';
 if(!auditHistorySession||auditHistorySession.grouped!==grouped||auditHistorySession.actor!==state.currentActorId)auditHistorySession={grouped,actor:state.currentActorId,now:new Date(),cache:new Map(),active:grouped?'today':'all'};
 const session=auditHistorySession,now=session.now,cache=session.cache;let active=session.active,busy=false,error='',generation=0;
 target.innerHTML='<div class="audit-toolbar"><label>Search loaded activity<input type="search" placeholder="Search loaded descriptions or event types…" aria-label="Search activity"></label><span class="muted audit-count" aria-live="polite"></span></div><div class="audit-list"></div>';
 const list=target.querySelector('.audit-list'),count=target.querySelector('.audit-count'),search=target.querySelector('input');
 const renderEvent=e=>`<article class="audit-event"><div class="audit-description">${esc(e.summary||e.kind.replaceAll('.',' '))}</div><time datetime="${esc(e.at)}">${esc(new Date(e.at).toLocaleString())}</time><details><summary>Verification details · record #${esc(e.sequence)}</summary><dl><dt>Event type</dt><dd>${esc(e.kind)}</dd><dt>Checksum</dt><dd><code>${esc(e.hash)}</code></dd></dl></details></article>`;
 function draw(){
  if(!target.isConnected)return;
  const page=cache.get(active)||{events:[],nextBefore:null},query=search.value.trim().toLowerCase();
  const shown=page.events.filter(e=>((e.summary||'')+' '+e.kind).toLowerCase().includes(query));
  const tabs=[['today','Today'],['yesterday','Yesterday'],['week','Earlier this week'],['last-week','Last week'],['older','Older activity']];
  if(grouped){const months=new Map();for(const p of cache.values())for(const e of p.events){const b=auditDateBucket(e.at,now);if(b.key.startsWith('month-'))months.set(b.key,b)}for(const b of [...months.values()].sort((a,b)=>a.order-b.order))tabs.push([b.key,b.label]);tabs.push(['all','All activity'])}
  count.textContent=(busy?'Loading… · ':'')+shown.length+' matching / '+page.events.length+' loaded records';
  list.innerHTML=(grouped?'<div class="audit-date-tabs" role="tablist" aria-label="Activity dates">'+tabs.map(([key,label],i)=>`<button type="button" id="audit-tab-${i}" role="tab" aria-selected="${key===active}" aria-controls="audit-date-panel" tabindex="${key===active?0:-1}" data-audit-date="${esc(key)}">${esc(label)}</button>`).join('')+'</div>':'')+'<div id="audit-date-panel" role="'+(grouped?'tabpanel':'region')+'" aria-busy="'+busy+'"'+(grouped?' aria-labelledby="audit-tab-'+tabs.findIndex(t=>t[0]===active)+'"':' aria-label="Activity"')+'>'+shown.map(renderEvent).join('')+(!shown.length?'<p class="muted">'+(busy?'Loading activity…':'No matching activity in this period.')+'</p>':'')+(error?'<p role="alert">'+esc(error)+'</p>':'')+((error||page.nextBefore)?'<button type="button" data-history-more '+(busy?'disabled':'')+'>'+(error?'Retry':'Load more')+'</button>':'')+'</div>';
  const buttons=[...list.querySelectorAll('[data-audit-date]')];
  buttons.forEach((button,i)=>{button.onclick=()=>select(button.dataset.auditDate);button.onkeydown=e=>{let next=i;if(e.key==='ArrowRight')next=(i+1)%buttons.length;else if(e.key==='ArrowLeft')next=(i+buttons.length-1)%buttons.length;else if(e.key==='Home')next=0;else if(e.key==='End')next=buttons.length-1;else return;e.preventDefault();select(buttons[next].dataset.auditDate);list.querySelector('[data-audit-date="'+active+'"]').focus()}});
  const more=list.querySelector('[data-history-more]');if(more)more.onclick=()=>fetchPage(true);
 }
 async function fetchPage(more=false){
  if(busy)return;
  const period=active,token=++generation,previous=cache.get(period);
  const range=auditHistoryWindow(period,now);if(range.empty){cache.set(period,{events:[],nextBefore:null});draw();return}
  busy=true;error='';draw();
  try{
   const result=await api(auditHistoryPageURL(period,more?previous?.nextBefore:null,now));
   if(!target.isConnected||token!==generation)return;
   const records=more&&previous?[...previous.events,...result.events]:result.events;
   cache.set(period,{events:records,nextBefore:result.nextBefore});
  }catch(e){if(token===generation&&target.isConnected)error='Unable to load activity. '+e.message}
  finally{if(token===generation&&target.isConnected){busy=false;draw()}}
 }
 function select(period){if(period===active)return;generation++;busy=false;error='';active=period;session.active=period;draw();if(!cache.has(period))fetchPage()}
 search.addEventListener('input',draw);draw();if(!cache.has(active))await fetchPage();
}
