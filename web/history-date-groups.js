// Calendar groups are a read-only view of the existing event history.
function auditDateBucket(timestamp,now=new Date()){
 const date=new Date(timestamp);if(Number.isNaN(date.getTime()))return {key:'unknown',label:'Unknown date',order:9999};
 const today=new Date(now.getFullYear(),now.getMonth(),now.getDate());const yesterday=new Date(today);yesterday.setDate(today.getDate()-1);
 const week=new Date(today);week.setDate(today.getDate()-(today.getDay()+6)%7);const lastWeek=new Date(week);lastWeek.setDate(week.getDate()-7);
 const nextDay=new Date(today);nextDay.setDate(today.getDate()+1);
 if(date>=nextDay)return {key:'future',label:'Future dates',order:-1};
 if(date>=today)return {key:'today',label:'Today',order:0};
 if(date>=yesterday)return {key:'yesterday',label:'Yesterday',order:1};
 if(date>=week)return {key:'week',label:'Earlier this week',order:2};
 if(date>=lastWeek)return {key:'last-week',label:'Last week',order:3};
 const month=now.getFullYear()*12+now.getMonth()-(date.getFullYear()*12+date.getMonth());
 return {key:'month-'+date.getFullYear()+'-'+date.getMonth(),label:date.toLocaleDateString(undefined,{month:'long',year:'numeric'}),order:100+month};
}
function renderAuditDateGroups(events,renderEvent,now=new Date()){
 const groups=new Map();for(const event of events){const bucket=auditDateBucket(event.at,now);if(!groups.has(bucket.key))groups.set(bucket.key,{...bucket,events:[]});groups.get(bucket.key).events.push(event)}
 return [...groups.values()].sort((a,b)=>a.order-b.order).map(group=>'<section class="audit-date-group"><h3>'+esc(group.label)+'</h3>'+group.events.map(renderEvent).join('')+'</section>').join('');
}
let selectedAuditDateGroup='today';
function auditHistoryTabs(events,query,renderEvent,selected='today',now=new Date()){
 const buckets=new Map();for(const event of events){const bucket=auditDateBucket(event.at,now);buckets.set(bucket.key,bucket)}
 buckets.set('today',{key:'today',label:'Today',order:0});
 const tabs=[{key:'all',label:'All activity',order:-2},...[...buckets.values()].sort((a,b)=>a.order-b.order)];
 const active=tabs.some(tab=>tab.key===selected)?selected:'all';
 const shown=events.filter(event=>(active==='all'||auditDateBucket(event.at,now).key===active)&&(event.summary+' '+event.kind).toLowerCase().includes(query));
 return {active,count:shown.length,html:'<div class="audit-date-tabs" role="tablist" aria-label="Activity dates">'+tabs.map((tab,index)=>'<button type="button" class="audit-date-tab" id="audit-tab-'+index+'" role="tab" aria-selected="'+(tab.key===active)+'" aria-controls="audit-date-panel" tabindex="'+(tab.key===active?'0':'-1')+'" data-audit-date="'+esc(tab.key)+'">'+esc(tab.label)+'</button>').join('')+'</div><div id="audit-date-panel" role="tabpanel" aria-labelledby="audit-tab-'+tabs.findIndex(tab=>tab.key===active)+'">'+(shown.map(renderEvent).join('')||'<p class="muted">No matching activity in this period.</p>')+'</div>'};
}
function drawAuditDateTabs(list,events,query,renderEvent,redraw){
 const result=auditHistoryTabs(events,query,renderEvent,selectedAuditDateGroup);selectedAuditDateGroup=result.active;list.innerHTML=result.html;
 const buttons=[...list.querySelectorAll('[data-audit-date]')];
 buttons.forEach((button,index)=>{button.onclick=()=>{selectedAuditDateGroup=button.dataset.auditDate;redraw()};button.onkeydown=e=>{let next=index;if(e.key==='ArrowRight')next=(index+1)%buttons.length;else if(e.key==='ArrowLeft')next=(index+buttons.length-1)%buttons.length;else if(e.key==='Home')next=0;else if(e.key==='End')next=buttons.length-1;else return;e.preventDefault();selectedAuditDateGroup=buttons[next].dataset.auditDate;redraw();list.querySelector('[data-audit-date="'+selectedAuditDateGroup+'"]').focus()}});
 return result.count;
}
