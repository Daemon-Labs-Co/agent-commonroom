const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const source=fs.readFileSync('web/chat-display.js','utf8');
function classes(){const set=new Set();return {set,toggle(n,on){on?set.add(n):set.delete(n)},contains:n=>set.has(n)}}
function row(id){const meta={items:[],append(n){this.items.push(n)}};return {dataset:{messageId:id},classList:classes(),meta,querySelector(s){return s==='.message-meta'?meta:s==='.agent-highlight-mark'?meta.items.find(n=>n.className.includes('agent-highlight-mark'))||null:null}}}
function fixture({self={},flags={},actor='local-owner',settingsButton=true}={}){
 const rows=['mine','theirs','agent','other-agent'].map(row);
 const list={classList:classes(),querySelectorAll:()=>rows};
 const commits=[],notes=[];let modalHtml='',baseRenders=0;
 const settings=settingsButton?{dataset:{},clicked:0,onclick(){this.clicked++}}:null;
 const toolbar={html:'',insertAdjacentHTML(_,h){this.html+=h}};
 const displayButton={};
 const ctx={view:'Room',console,
  state:{currentActorId:actor,featureFlags:flags,
   actors:[{id:'local-owner',kind:'human',...(actor==='local-owner'?self:{})},{id:'dev',kind:'human',...(actor==='dev'?self:{})},{id:'bot-ca',kind:'agent'}],
   participants:[{id:'p1',name:'Alex',actorId:'bot-ca',connectionId:'ca'},{id:'p2',name:'Blake',connectionId:'cb'}],
   messages:[{id:'mine',actorId:actor,actorType:'human'},{id:'theirs',actorId:actor==='dev'?'local-owner':'dev',actorType:'human'},{id:'agent',actorId:'bot-ca',actorType:'agent'},{id:'other-agent',actorId:'bot-cb',actorType:'agent'}]},
  esc:s=>String(s),note:m=>notes.push(m),commit:async(k,p)=>{commits.push([k,p])},
  form:(title)=>{ctx.formTitle=title},
  $:s=>s==='#room-settings'?settings:s==='.toolbar .actions'?toolbar:s==='#chat-display-settings'?(toolbar.html?displayButton:null):s==='#modal .actions'?{insertAdjacentHTML(_,h){modalHtml+=h}}:s==='#cancel'?(ctx.cancel={}):s==='#modal'?{querySelector:()=>null,querySelectorAll:()=>[]}:null,
  document:{querySelector:s=>s==='.message-list'?list:null,createElement:()=>({className:'',textContent:'',title:''})},
  render(){baseRenders++},openAgentModelOverlay(){ctx.overlayOpened=true}};
 vm.createContext(ctx);vm.runInContext(source,ctx);
 return {ctx,rows,list,commits,settings,toolbar,displayButton,modal:()=>modalHtml,renders:()=>baseRenders};
}
const has=(r,n)=>r.classList.contains(n);

// Default: nothing changes for anyone who has not chosen.
{const f=fixture();f.ctx.render();assert.equal(f.renders(),1);assert.equal(f.list.classList.contains('chat-style'),false);assert.ok(f.rows.every(r=>!has(r,'message-own')&&!has(r,'agent-highlighted')));}

// Chat style: only the viewer's own messages move right; other humans and agents stay.
{const f=fixture({self:{chatStyle:true}});f.ctx.render();assert.ok(f.list.classList.contains('chat-style'));assert.deepEqual(f.rows.map(r=>has(r,'message-own')),[true,false,false,false]);}
{const f=fixture({self:{chatStyle:true},actor:'dev',settingsButton:false});f.ctx.render();assert.deepEqual(f.rows.map(r=>has(r,'message-own')),[true,false,false,false]);
 assert.match(f.toolbar.html,/id="chat-display-settings"/,'a person without Room settings still reaches their display choices');
 f.displayButton.onclick();assert.equal(f.ctx.formTitle,'Display settings');assert.equal(f.ctx.cancel.textContent,'Close');assert.match(f.modal(),/id="chat-style-toggle" checked/);}

// Highlight: tint and a text marker, added once across repeated renders.
{const f=fixture({self:{highlightedAgents:['bot-ca']}});f.ctx.render();f.ctx.render();
 assert.deepEqual(f.rows.map(r=>has(r,'agent-highlighted')),[false,false,true,false]);
 assert.equal(f.rows[2].meta.items.length,1);assert.equal(f.rows[2].meta.items[0].textContent,'★ Highlighted');
 assert.equal(f.rows[0].meta.items.length,0);}

// Flag off: saved choices are ignored and no controls are offered.
{const f=fixture({self:{chatStyle:true,highlightedAgents:['bot-ca']},flags:{chatDisplayChoices:false}});f.ctx.render();
 assert.equal(f.list.classList.contains('chat-style'),false);assert.ok(f.rows.every(r=>!has(r,'message-own')&&!has(r,'agent-highlighted')));
 f.settings.onclick();assert.equal(f.modal(),'');assert.equal(f.settings.clicked,1);}

// Room settings gains the section once; fields have no name so the room form never submits them.
{const f=fixture({self:{chatStyle:true,highlightedAgents:['bot-cb']}});f.ctx.render();f.ctx.render();f.settings.onclick();
 assert.equal(f.settings.clicked,1,'the original Room settings handler still runs, once');
 const html=f.modal();assert.match(html,/id="chat-style-toggle" checked/);assert.match(html,/data-highlight-agent="bot-ca" >/);assert.match(html,/data-highlight-agent="bot-cb" checked/);
 assert.doesNotMatch(html,/name=/);assert.match(html,/applies in every room/);}

// The agent overlay offers the highlight and saving sends the personal event.
(async()=>{
 const f=fixture();f.ctx.openAgentModelOverlay({id:'p2',name:'Blake',connectionId:'cb'});
 assert.ok(f.ctx.overlayOpened);assert.match(f.modal(),/data-highlight-agent="bot-cb" >/);
 const box={checked:true,disabled:false,dataset:{highlightAgent:'bot-cb'}},toggle={checked:true,disabled:false};
 f.ctx.wireChatDisplayFields({querySelector:()=>toggle,querySelectorAll:()=>[box]});
 await toggle.onchange();await box.onchange();
 assert.equal(JSON.stringify(f.commits),JSON.stringify([['actor.chat_style.updated',{enabled:true}],['actor.agent_highlight.updated',{agentId:'bot-cb',enabled:true}]]));
 // A refused save puts the control back.
 f.ctx.commit=async()=>{throw Error('disabled')};await toggle.onchange();assert.equal(toggle.checked,false);assert.equal(toggle.disabled,false);
 const page=fs.readFileSync('web/index.html','utf8');
 assert.ok(page.indexOf('src="chat-display.js')>page.indexOf('src="squad-profiles.js'));assert.ok(page.indexOf('src="chat-display.js')<page.indexOf('src="feature-flag-tabs.js"'),'the flag card is grouped with the other cosmetic flags');
 console.log('Chat display: own messages right, agent highlights, personal saves and both flag states.');
})();
