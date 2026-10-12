'use strict';
process.env.TZ='America/Toronto';
const {test}=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const source=fs.readFileSync('web/history.js','utf8');
function fixture(grouped=true,refined=false){
 const calls=[],pending=[],buttons=[];
 const list={innerHTML:'',querySelectorAll(){return buttons},querySelector(){return null}};
 const input={value:'',addEventListener(){}};const count={};const refreshButton={};const category={value:'All categories',addEventListener(k,fn){this.change=fn}},outcome={value:'all',addEventListener(k,fn){this.change=fn}},routine={checked:false,addEventListener(k,fn){this.change=fn}};
 const target={isConnected:true,innerHTML:'',querySelector(selector){return selector==='.audit-list'?list:selector==='.audit-count'?count:selector==='.audit-refresh'?refreshButton:selector==='.audit-category-filter'?category:selector==='.audit-outcome-filter'?outcome:selector==='.audit-routine-filter input'?routine:input}};
 const context=vm.createContext({Date,URLSearchParams,render(){},view:'History & retention',document:{querySelector(){return target}},state:{featureFlags:{historyDateGroups:grouped},actors:[],rooms:[]},auditDateBucket(){return {key:'today'}},esc:String,api(url){calls.push(url);return new Promise((resolve,reject)=>pending.push({resolve,reject}))}});
 if(refined)vm.runInContext(fs.readFileSync('web/history-presentation.js','utf8'),context);vm.runInContext(source,context);return {context,calls,pending,target,list,buttons,category,outcome,routine,refreshButton};
}
test('opens Today with one bounded asynchronous request and visible loading state',async()=>{
 const f=fixture();const load=f.context.loadAuditHistory();assert.equal(f.calls.length,1);assert(f.calls[0].includes('page=true'));assert(f.calls[0].includes('from='));assert(f.list.innerHTML.includes('Loading activity'));assert(f.list.innerHTML.includes('aria-selected="true"'));f.pending[0].resolve({events:[],nextBefore:null});await load;assert(f.list.innerHTML.includes('No matching activity'));assert.equal(f.calls.length,1);
});
test('disabled calendar flag uses the same bounded async loader without tabs',async()=>{
 const f=fixture(false);const load=f.context.loadAuditHistory();assert.equal(f.calls[0],'events?page=true');assert(!f.list.innerHTML.includes('tablist'));f.pending[0].resolve({events:[],nextBefore:null});await load;
});
test('failed loads offer retry and detached responses do not change another view',async()=>{
 const f=fixture();const load=f.context.loadAuditHistory();f.pending[0].reject(new Error('offline'));await load;assert(f.list.innerHTML.includes('Retry'));assert(f.list.innerHTML.includes('role="alert"'));
 const detached=fixture();const waiting=detached.context.loadAuditHistory();detached.target.isConnected=false;const before=detached.list.innerHTML;detached.pending[0].resolve({events:[],nextBefore:null});await waiting;assert.equal(detached.list.innerHTML,before);
});
test('calendar windows respect DST and cursor/date query encoding',()=>{
 const f=fixture();const now=new Date(2026,10,1,12);const range=f.context.auditHistoryWindow('today',now);assert.equal((new Date(range.until)-new Date(range.from))/3600000,25);const url=f.context.auditHistoryPageURL('all',123,now);assert.equal(url,'events?page=true&before=123');assert(f.context.auditHistoryWindow('week',new Date(2026,9,12)).empty);
});

test('All is lazy, stale Today responses are ignored, and Load more advances the cursor',async()=>{
 const f=fixture();const today={dataset:{auditDate:'today'}},all={dataset:{auditDate:'all'}};f.buttons.push(today,all);
 const more={};f.list.querySelector=()=>f.list.innerHTML.includes('data-history-more')?more:null;
 const initial=f.context.loadAuditHistory();assert.equal(f.calls.length,1);all.onclick();assert.equal(f.calls.length,2);assert.equal(f.calls[1],'events?page=true');
 f.pending[0].resolve({events:[{sequence:999,at:new Date(),summary:'stale',kind:'fixture'}],nextBefore:null});await initial;assert(!f.list.innerHTML.includes('stale'));
 f.pending[1].resolve({events:[{sequence:120,at:new Date(),summary:'first',kind:'fixture',hash:'fixture'}],nextBefore:120});await new Promise(resolve=>setImmediate(resolve));assert(f.list.innerHTML.includes('first'));
 more.onclick();assert.equal(f.calls[2],'events?page=true&before=120');more.onclick();assert.equal(f.calls.length,3);
 f.pending[2].resolve({events:[{sequence:119,at:new Date(),summary:'second',kind:'fixture',hash:'fixture'}],nextBefore:null});await new Promise(resolve=>setImmediate(resolve));assert(f.list.innerHTML.includes('first'));assert(f.list.innerHTML.includes('second'));
 await f.context.loadAuditHistory();assert.equal(f.calls.length,3,'background redraw should preserve loaded period');
 f.context.view='Room';f.context.render();f.context.view='History & retention';const reopened=f.context.loadAuditHistory();assert(f.calls[3].includes('from='));f.pending[3].resolve({events:[],nextBefore:null});await reopened;
});

test('room-scoped API preserves pagination date parameters',async()=>{
 const calls=[];const context=vm.createContext({api:async(path,body)=>{calls.push(path)},roomId:'engineering'});
 const wrapper=fs.readFileSync('web/features.js','utf8').match(/const baseApi=api;\napi=async function[^\n]+/)[0];vm.runInContext(wrapper,context);
 await context.api('events?page=true&from=2026-10-11T04%3A00%3A00.000Z');assert.equal(calls[0],'events?page=true&from=2026-10-11T04%3A00%3A00.000Z&room=engineering');
 await context.api('state');assert.equal(calls[1],'state?room=engineering');
});

test('refined loader filters loaded data and preserves filters across background rerenders',async()=>{
 const f=fixture(true,true),load=f.context.loadAuditHistory();
 const event=(sequence,summary,category,tone,routine=false)=>({sequence,summary,kind:'synthetic',hash:'synthetic',at:new Date().toISOString(),presentation:{category,tone,routine}});
 f.pending[0].resolve({events:[event(3,'Routine session entry','Agents','info',true),event(2,'Connection failure','Agents','error'),event(1,'Task accepted','Tasks & evidence','success')],nextBefore:null});await load;
 assert(!f.list.innerHTML.includes('Routine session entry'));assert(f.list.innerHTML.includes('Connection failure'));assert(f.list.innerHTML.includes('Task accepted'));
 f.category.value='Agents';f.category.change();assert(!f.list.innerHTML.includes('Task accepted'));
 f.category.value='All categories';f.category.change();f.outcome.value='success';f.outcome.change();assert(!f.list.innerHTML.includes('Connection failure'));assert(f.list.innerHTML.includes('Task accepted'));
 await f.context.loadAuditHistory();assert.equal(f.outcome.value,'success');assert.equal(f.calls.length,1);
});

test('explicit refresh resets loaded pages without clearing selected filters',async()=>{
 const f=fixture(true,true),load=f.context.loadAuditHistory();f.pending[0].resolve({events:[],nextBefore:50});await load;
 f.category.value='Agents';f.category.change();f.refreshButton.onclick();assert.equal(f.calls.length,2);assert(!f.calls[1].includes('before='));f.pending[1].resolve({events:[],nextBefore:null});await new Promise(resolve=>setImmediate(resolve));assert.equal(f.category.value,'Agents');
});
