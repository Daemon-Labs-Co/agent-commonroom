'use strict';
const {test}=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
function harness({owner=true,view='Room',enabled=true}={}){
 const nodes={},calls=[];let html='',save,resolve;
 const insert=(_,text)=>{html+=text;for(const m of text.matchAll(/id="([^"]+)"/g))nodes['#'+m[1]]={innerHTML:'',insertAdjacentHTML:insert}};
 const main={insertAdjacentHTML:insert};
 const ctx=vm.createContext({state:{currentActorId:owner?'local-owner':'member',featureFlags:{roomOutcomeReview:enabled},rooms:[{id:'engineering',name:'Synthetic room',outcome:'<outcome>',acceptanceCriteria:'Synthetic criteria'}]},roomId:'engineering',view,render(){},$:s=>s==='#main'?main:nodes[s],esc:s=>String(s??'').replaceAll('<','&lt;'),api:async(...args)=>{calls.push(args);return args[1]?{}:new Promise(r=>resolve=r)},form(_,body,fn){html+=body;save=fn},commit:async(...args)=>calls.push(args),refresh:async()=>{},note:message=>calls.push(['notice',message])});
 vm.runInContext(fs.readFileSync('web/room-outcome-review.js','utf8'),ctx);
 return{ctx,nodes,calls,html:()=>html,submit:()=>save(new Map([['note','Synthetic human rationale'],['enabled','on']])),load:async data=>{resolve(data);await new Promise(r=>setImmediate(r))}};
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
