const fs=require('node:fs'),vm=require('node:vm'),assert=require('node:assert/strict');
const esc=s=>String(s??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));let rendered=0;
const ctx=vm.createContext({esc,render(){rendered++},form(){},state:{currentActorId:'local-owner',featureFlags:{companyOrgChart:false}},document:{querySelectorAll(){throw Error('disabled chart touched DOM')}}});vm.runInContext(fs.readFileSync('web/organization-chart.js','utf8'),ctx);
vm.runInContext('render()',ctx);assert.equal(rendered,1);
const data={actors:[{id:'a'},{id:'b'},{id:'c'}],organization:{nodes:[{actorId:'a',reportsTo:'missing',x:15,y:20},{actorId:'b',reportsTo:'a',x:50,y:80}]}};let nodes=ctx.organizationNodes(data);assert.equal(nodes.length,3);assert.equal(nodes[0].reportsTo,'');assert.equal(nodes[0].x,15);assert.equal(nodes[1].reportsTo,'a');assert.equal(nodes[2].reportsTo,'');
assert.equal(ctx.organizationWouldCycle(nodes,'a','b'),true);assert.equal(ctx.organizationWouldCycle(nodes,'b','c'),false);assert.equal(ctx.organizationWouldCycle(nodes,'c','c'),true);
ctx.arrangeOrganization(nodes);assert(nodes[1].y>nodes[0].y);assert.equal(nodes[0].y,nodes[2].y);
const html=ctx.organizationNodeHTML({id:'a',name:'<script>bad</script>',kind:'human',title:'Owner',enabled:false},nodes[0],'a');assert(!html.includes('<script>'));assert(html.includes('&lt;script&gt;'));assert(html.includes('Disabled'));assert(html.includes('selected'));
ctx.state.currentActorId='dev';ctx.openOrganizationChart().then(()=>console.log('Organization projection, orphan handling, cycles, layout, escaping, disabled state and non-owner opening checks passed.'));

// Editor routing must use company identities, not display names or room aliases.
const edits=[],notes=[];
ctx.connection=c=>edits.push(['connection',c.id]);ctx.memberAccessForm=p=>edits.push(['human',p.id]);ctx.note=m=>notes.push(m);
ctx.$=selector=>selector==='#edit-human-name'?{onclick:()=>edits.push(['owner'])}:null;
ctx.state={currentActorId:'local-owner',featureFlags:{companyOrgChart:true},actors:[{id:'local-owner',kind:'human'},{id:'person',kind:'human'},{id:'bot-c1',kind:'agent'},{id:'bot-gone',kind:'agent'}],connections:[{id:'c1',name:'Same name'}]};
assert.equal(ctx.editOrganizationIdentity('bot-c1'),true);assert.equal(ctx.editOrganizationIdentity('person'),true);assert.equal(ctx.editOrganizationIdentity('local-owner'),true);
assert.deepEqual(edits,[['connection','c1'],['human','person'],['owner']]);
assert.equal(ctx.editOrganizationIdentity('bot-gone'),false);assert.equal(ctx.editOrganizationIdentity('missing'),false);assert.equal(notes.length,2);
ctx.state.featureFlags.companyOrgChart=false;assert.equal(ctx.editOrganizationIdentity('bot-c1'),false);
ctx.state.featureFlags.companyOrgChart=true;ctx.state.currentActorId='person';assert.equal(ctx.editOrganizationIdentity('bot-c1'),false);assert.equal(edits.length,3);
const button={},events=[];ctx.wireOrganizationIdentity(button,'bot-c1',id=>events.push(['select',id]),id=>events.push(['edit',id]));
button.onclick();button.onclick();button.ondblclick();assert.deepEqual(events,[['select','bot-c1'],['select','bot-c1'],['edit','bot-c1']]);
console.log('Identity editor routing, missing identities, owner/flag guards and double-click dispatch passed.');
