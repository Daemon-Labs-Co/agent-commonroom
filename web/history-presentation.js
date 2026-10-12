// Read-only organization of bounded audit pages; original records stay intact.
function historyRefined(){return state?.featureFlags?.historyRefinement!==false}
const historyCategories=['All categories','Messages','Agents','Workflows','Tasks & evidence','Memory','Connections','People & company','Rooms','Security','Settings','Other'];
function auditDisplayData(event){const p=event.presentation||{};return {...p,category:p.category||'Other',tone:['info','warning','error','success'].includes(p.tone)?p.tone:'info'}}
function auditMatchesFilters(event,filters){
 const d=auditDisplayData(event);
 if(!filters.routine&&d.routine)return false;
 if(filters.category&&filters.category!=='All categories'&&d.category!==filters.category)return false;
 if(filters.outcome==='attention'&&!['warning','error'].includes(d.tone))return false;
 if(filters.outcome==='error'&&d.tone!=='error')return false;
 if(filters.outcome==='success'&&d.tone!=='success')return false;
 const haystack=[event.summary,event.kind,d.actor,d.actorId,d.subject,d.roomId,d.outcome,d.explanation,historyActorName(d),historyRoomName(d)].join(' ').toLowerCase();return haystack.includes((filters.search||'').trim().toLowerCase());
}
function historyActorName(d){return d.actor||state.actors?.find(a=>a.id===d.actorId)?.name||''}
function historyRoomName(d){return state.rooms?.find(r=>r.id===d.roomId)?.name||d.roomId||''}
function renderRefinedAuditEvent(event){
 const d=auditDisplayData(event),actor=historyActorName(d),room=historyRoomName(d),when=new Date(event.at);
 const context=[actor&&(event.kind==='workflow.run.updated'?'Run owner: ':'Actor: ')+actor,room&&'Room: '+room,d.subject&&'Subject: '+d.subject].filter(Boolean);
 return '<article class="audit-event audit-refined audit-tone-'+d.tone+'"><div class="audit-event-main"><div class="audit-event-heading"><span class="audit-category">'+esc(d.category)+'</span>'+(d.outcome?'<span class="audit-outcome">'+esc(d.outcome.replaceAll('_',' '))+'</span>':'')+'<time datetime="'+esc(event.at)+'" title="'+esc(when.toLocaleString())+'">'+esc(when.toLocaleTimeString([], {hour:'2-digit',minute:'2-digit'}))+'</time></div><strong class="audit-description">'+esc(event.summary||event.kind.replaceAll('.',' '))+'</strong>'+(context.length?'<p class="audit-context">'+context.map(esc).join(' · ')+'</p>':'<p class="audit-context">Actor and scope were not recorded for this event.</p>')+(d.explanation?'<p class="audit-explanation">'+esc(d.explanation)+'</p>':'')+(d.next?'<p class="audit-next"><strong>Next:</strong> '+esc(d.next)+'</p>':'')+'<details class="audit-technical"><summary>Technical details · record #'+esc(event.sequence)+'</summary><dl><dt>Recorded at</dt><dd>'+esc(when.toLocaleString())+'</dd><dt>Event type</dt><dd>'+esc(event.kind)+'</dd><dt>Checksum</dt><dd><code>'+esc(event.hash)+'</code></dd></dl></details></div></article>';
}
function renderRefinedAuditList(events){
 const groups=new Map();for(const event of events){const date=new Date(event.at),key=Number.isNaN(+date)?'Unknown date':date.toLocaleDateString(undefined,{weekday:'long',year:'numeric',month:'long',day:'numeric'});if(!groups.has(key))groups.set(key,[]);groups.get(key).push(event)}
 return [...groups].map(([day,rows])=>'<section class="audit-day"><h3>'+esc(day)+'</h3>'+rows.map(renderRefinedAuditEvent).join('')+'</section>').join('');
}
function auditLoadedOverview(events,shown){
 const attention=events.filter(e=>['error','warning'].includes(auditDisplayData(e).tone)).length,routine=events.filter(e=>auditDisplayData(e).routine).length;
 return '<div class="audit-overview"><span><strong>'+events.length+'</strong> loaded records</span><span><strong>'+shown.length+'</strong> shown</span><span><strong>'+attention+'</strong> errors / attention events</span><span><strong>'+routine+'</strong> routine updates</span></div>';
}
