process.env.TZ='America/Toronto';
const fs=require('node:fs'),vm=require('node:vm'),assert=require('node:assert/strict');const ctx=vm.createContext({Date,esc:s=>String(s).replaceAll('<','&lt;')});vm.runInContext(fs.readFileSync('web/history-date-groups.js','utf8'),ctx);
const bucket=(date,now)=>ctx.auditDateBucket(date,now).label;const now=new Date(2026,9,11,10);
assert.equal(bucket(new Date(2026,9,11,1),now),'Today');assert.equal(bucket(new Date(2026,9,10,23),now),'Yesterday');assert.equal(bucket(new Date(2026,9,8),now),'Earlier this week');assert.equal(bucket(new Date(2026,9,2),now),'Last week');assert.equal(bucket(new Date(2026,8,14),now),'September 2026');assert.equal(bucket('invalid',now),'Unknown date');assert.equal(bucket(new Date(2026,9,12),now),'Future dates');
const monday=new Date(2026,9,12,10);assert.equal(bucket(new Date(2026,9,10),monday),'Last week');assert.equal(bucket(new Date(2026,9,11),monday),'Yesterday');
const dst=new Date(2026,10,2,10);assert.equal(bucket(new Date(2026,10,1,0,30),dst),'Yesterday');
const january=new Date(2027,0,8);assert.equal(bucket(new Date(2026,10,1),january),'November 2026');
const events=Object.freeze([Object.freeze({at:new Date(2026,8,1),id:'older'}),Object.freeze({at:new Date(2026,9,11),id:'today'})]);const html=ctx.renderAuditDateGroups(events,e=>'<article>'+e.id+'</article>',now);assert(html.indexOf('Today')<html.indexOf('September'));assert(html.includes('<article>older</article>'));assert.equal(events[0].id,'older');assert.equal(ctx.renderAuditDateGroups([],()=>'',now),'');
console.log('Local calendar grouping, week boundaries, DST, month/year rollover and immutable event rendering passed.');
const records=[{at:new Date(2026,9,11),summary:'Agent invited',kind:'participant.created',hash:'synthetic-a'},{at:new Date(2026,9,10),summary:'Room created',kind:'room.created',hash:'synthetic-b'}];
const render=e=>'<article>'+e.summary+'</article>';let tabs=ctx.auditHistoryTabs(records,'',render,'today',now);assert.equal(tabs.count,1);assert(tabs.html.includes('role="tablist"'));assert(tabs.html.includes('Today'));assert(tabs.html.includes('Yesterday'));assert(tabs.html.includes('Agent invited'));assert(!tabs.html.includes('<article>Room created'));
tabs=ctx.auditHistoryTabs(records,'room',render,'today',now);assert.equal(tabs.count,0);assert(tabs.html.includes('No matching activity in this period.'));assert(tabs.html.includes('Yesterday'),'search must preserve date navigation');
tabs=ctx.auditHistoryTabs(records,'room',render,'all',now);assert.equal(tabs.count,1);assert(tabs.html.includes('Room created'));tabs=ctx.auditHistoryTabs(records,'',render,'missing',now);assert.equal(tabs.active,'all');assert.equal(tabs.count,2);
console.log('History tabs isolate date periods, search within the active tab and retain navigation on empty results.');

const emptyToday=ctx.auditHistoryTabs([records[1]],'',render,undefined,now);assert.equal(emptyToday.active,'today');assert.equal(emptyToday.count,0);assert(emptyToday.html.includes('Today'));
