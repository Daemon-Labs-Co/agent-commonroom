// Display-only audit projection. Original event records and checksums stay intact.
async function loadAuditHistory(){
 const target=document.querySelector('#audit');
 const events=await api('events');
 if(!target.isConnected)return;
 target.innerHTML='<div class="audit-toolbar"><label>Search recent activity<input type="search" placeholder="Search descriptions or event types…" aria-label="Search activity"></label><span class="muted audit-count"></span></div><div class="audit-list"></div>';
 const list=target.querySelector('.audit-list'),count=target.querySelector('.audit-count');
 const draw=()=>{
  const query=target.querySelector('input').value.trim().toLowerCase();
  const shown=events.filter(e=>(e.summary+' '+e.kind).toLowerCase().includes(query));
  count.textContent=shown.length+' of '+events.length+' recent records (up to 200)';
  const renderEvent=e=>`<article class="audit-event"><div class="audit-description">${esc(e.summary||e.kind.replaceAll('.',' '))}</div><time datetime="${esc(e.at)}">${esc(new Date(e.at).toLocaleString())}</time><details><summary>Verification details · record #${e.sequence}</summary><dl><dt>Event type</dt><dd>${esc(e.kind)}</dd><dt>Checksum</dt><dd><code>${esc(e.hash)}</code></dd></dl></details></article>`;
  if(typeof drawAuditDateTabs==='function'&&state?.featureFlags?.historyDateGroups!==false){
   const visible=drawAuditDateTabs(list,events,query,renderEvent,draw);count.textContent=visible+' of '+events.length+' recent records (up to 200)';
  }else list.innerHTML=shown.map(renderEvent).join('')||'<p class="muted">No matching activity.</p>';
 };
 target.querySelector('input').addEventListener('input',draw);draw();
}
