'use strict';
const {test}=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
function harness({owner=true,view='Room',enabled=true}={}){
 const nodes={},calls=[],titles=[];let html='',save,resolve;
 const insert=(_,text)=>{html+=text;for(const m of text.matchAll(/id="([^"]+)"/g))nodes['#'+m[1]]={innerHTML:'',insertAdjacentHTML:insert}};
 const main={insertAdjacentHTML:insert};
 const ctx=vm.createContext({state:{currentActorId:owner?'local-owner':'member',featureFlags:{roomOutcomeReview:enabled},rooms:[{id:'engineering',name:'Synthetic room',outcome:'<outcome>',acceptanceCriteria:'Synthetic criteria'}]},roomId:'engineering',view,render(){},$:s=>s==='#main'?main:nodes[s],esc:s=>String(s??'').replaceAll('<','&lt;'),api:async(...args)=>{calls.push(args);return args[1]?{}:new Promise(r=>resolve=r)},form(title,body,fn){titles.push(title);html+=body;save=fn},commit:async(...args)=>calls.push(args),refresh:async()=>{},note:message=>calls.push(['notice',message])});
 vm.runInContext(fs.readFileSync('web/room-outcome-review.js','utf8'),ctx);
 return{ctx,nodes,calls,titles,html:()=>html,submit:()=>save(new Map([['note','Synthetic human rationale'],['enabled','on']])),load:async data=>{resolve(data);await new Promise(r=>setImmediate(r))}};
}
const open={enabled:true,ready:true,status:'Open',snapshot:'s'.repeat(64),review:{revision:0},definition:{name:'Current room',outcome:'<current outcome>',acceptanceCriteria:'Current criterion'},tasks:[{title:'Current step',status:'Accepted'}],evidence:[{id:'e',body:'<current evidence>',source:'Synthetic'}]};
test('owner reviews exact snapshot with rationale and never dispatches work',async()=>{
 const h=harness();h.ctx.drawRoomOutcomeReview();await h.load(open);h.nodes['#room-review-accept'].onclick();assert.match(h.html(),/&lt;current outcome>/);assert.match(h.html(),/&lt;current evidence>/);assert(!h.html().includes('&lt;outcome>'));assert.match(h.html(),/Evidence reviewed/);await h.submit();const [path,body]=h.calls[1];assert.equal(path,'room-outcome-review?room=engineering');assert.equal(body.snapshot,open.snapshot);assert.equal(body.revision,0);assert.equal(body.action,'accept');assert.equal(body.note,'Synthetic human rationale');assert(!h.calls.some(([path])=>path.includes('dispatch')));
 assert(fs.readFileSync('web/app.js','utf8').includes("'room-outcome-review'"));
});
test('members and disabled flag see history without decision controls',async()=>{
 for(const options of [{owner:false},{enabled:false}]){const h=harness(options);h.ctx.drawRoomOutcomeReview();await h.load({...open,enabled:options.enabled!==false,status:'Accepted',review:{status:'Accepted',note:'<review note>',reviewedBy:'local-owner',reviewedAt:'synthetic'}});assert.match(h.nodes['#room-review-status'].innerHTML,/&lt;review note>/);assert.equal(h.nodes['#room-review-accept'],undefined);assert.equal(h.nodes['#room-review-reopen'],undefined)}
});
test('changed work shows needs review, allows reopening, and guards stale room responses and submits',async()=>{
 const h=harness();h.ctx.drawRoomOutcomeReview();await h.load({...open,ready:false,status:'Needs review',review:{status:'Accepted',revision:1}});assert.equal(h.nodes['#room-review-accept'],undefined);h.nodes['#room-review-reopen'].onclick();h.ctx.roomId='other';await h.submit();assert.equal(h.calls.length,2);assert.equal(h.calls[1][0],'notice');
 const stale=harness();stale.ctx.drawRoomOutcomeReview();stale.ctx.roomId='other';await stale.load(open);assert.equal(stale.nodes['#room-review-accept'],undefined);
});
test('review flag defaults off, is owner-only and precedes Settings grouping',async()=>{
 const h=harness({view:'Settings',enabled:false});h.ctx.drawRoomOutcomeReview();assert.match(h.html(),/Room outcome review · Off/);h.nodes['#room-review-setting'].onclick();await h.submit();assert.equal(h.calls[0][1].name,'roomOutcomeReview');assert.equal(h.calls[0][1].enabled,true);
 const member=harness({view:'Settings',owner:false});member.ctx.drawRoomOutcomeReview();assert.equal(member.html(),'');const page=fs.readFileSync('web/index.html','utf8');assert(page.indexOf('src="room-outcome-review.js"')<page.indexOf('src="feature-flag-tabs.js"'));assert(page.indexOf('src="room-outcome-review.js"')<page.indexOf('src="settings-tabs.js"'));
});

test('accept and reopen titles escape room names as text',async()=>{
 for(const action of ['accept','reopen']){
  const h=harness();h.ctx.drawRoomOutcomeReview();await h.load({...open,status:action==='accept'?'Open':'Accepted',review:action==='accept'?{revision:0}:{revision:1,status:'Accepted'},definition:{...open.definition,name:'Build <web><img src=x onerror=alert(1)>'}});
  h.nodes['#room-review-'+action].onclick();
  assert.equal(h.titles[0],(action==='accept'?'Accept ':'Reopen ')+'Build &lt;web>&lt;img src=x onerror=alert(1)>');
  assert(!h.titles[0].includes('<web>'));assert(!h.titles[0].includes('<img'));
 }
});
test('review shows task criteria and evidence associations without treating text as markup',async()=>{
 const h=harness();h.ctx.drawRoomOutcomeReview();await h.load({...open,tasks:[{id:'task-a',title:'<Task A>',status:'Accepted',criteria:'<Criterion A>'},{id:'task-b',title:'Task B',status:'Accepted',criteria:'Criterion B'}],evidence:[{id:'e-a',taskId:'task-a',title:'First evidence',body:'<Evidence A>'},{id:'e-b',taskId:'task-b',title:'Second evidence',body:'Evidence B'},{id:'room-e',body:'Room evidence'},{id:'old-e',taskId:'<missing>',body:'Historical evidence'}]});
 h.nodes['#room-review-accept'].onclick();const markup=h.html();
 assert.match(markup,/Acceptance criteria: &lt;Criterion A>/);assert.match(markup,/Acceptance criteria: Criterion B/);
 assert.match(markup,/Task: &lt;Task A> \(task-a\)/);assert.match(markup,/Task: Task B \(task-b\)/);assert.match(markup,/Room-level evidence/);assert.match(markup,/Not in current task snapshot \(&lt;missing>\)/);
 assert.match(markup,/&lt;Evidence A>/);assert(!markup.includes('<Task A>'));assert(!markup.includes('<Criterion A>'));
});
