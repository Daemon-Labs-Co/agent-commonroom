const fs=require('node:fs'),vm=require('node:vm'),assert=require('node:assert/strict');
let baseCalls=0;const ctx=vm.createContext({view:'Room',state:{currentActorId:'local-owner',featureFlags:{}},render(){baseCalls++},$(){throw Error('disabled/non-settings layout must not touch DOM')},document:{}});
vm.runInContext(fs.readFileSync('web/feature-flag-tabs.js','utf8'),ctx);
for(const [title,expected] of [['Security · Connection sign-in','Security'],['Cosmetic · Help window choice','Cosmetic'],['Connections · Connection testing','Connections'],['Room presentation','Cosmetic'],['Your profile',null]])assert.equal(vm.runInContext(`featureFlagCategory(${JSON.stringify(title)})`,ctx),expected);
vm.runInContext('render()',ctx);assert.equal(baseCalls,1);
ctx.view='Settings';ctx.state.featureFlags.featureFlagTabs=false;vm.runInContext('render()',ctx);assert.equal(baseCalls,2);
ctx.state.featureFlags.featureFlagTabs=true;ctx.state.currentActorId='developer';vm.runInContext('render()',ctx);assert.equal(baseCalls,3);
console.log('Feature category mapping, legacy disabled fallback and owner-only presentation passed.');
function element(tag,text=''){
 return {tagName:tag.toUpperCase(),textContent:text,children:[],attributes:{},classList:{add(){}},setAttribute(k,v){this.attributes[k]=v},append(child){this.children.push(child)},focus(){this.focused=true},querySelector(){return this.heading||null},replaceWith(node){this.replacement=node}};
}
const base=element('div');base.heading=element('h2','Feature flags');const securityControl=element('button','Configure emergency controls');let invoked=0;securityControl.onclick=()=>invoked++;
base.children=[base.heading,element('h3','Security'),securityControl,element('h3','Cosmetic'),element('button','Configure photos'),element('h3','Connections'),element('p','Connection editor'),element('h3','Workflows'),element('p','Workflow templates')];
const roomCard=element('div');roomCard.heading=element('h2','Room presentation');const signin=element('div');signin.heading=element('h3','Security · Connection sign-in');const profile=element('div');profile.heading=element('h2','Your profile');
ctx.state.currentActorId='local-owner';ctx.$=()=>({querySelectorAll(){return [base,roomCard,signin,profile]}});ctx.document.createElement=tag=>element(tag);vm.runInContext('render()',ctx);
const grouped=base.replacement;assert(grouped);const nav=grouped.children[0];assert.deepEqual(nav.children.map(b=>b.textContent),['Security','Cosmetic','Connections','Workflows']);
const security=grouped.children[1],cosmetic=grouped.children[2];assert.equal(security.hidden,false);assert.equal(cosmetic.hidden,true);assert(security.children.includes(securityControl));assert(security.children.includes(signin));assert(cosmetic.children.includes(roomCard));securityControl.onclick();assert.equal(invoked,1,'moving controls must preserve their handlers');nav.children[1].onclick();assert.equal(cosmetic.hidden,false);assert.equal(security.hidden,true);nav.children[1].onkeydown({key:'End',preventDefault(){}});assert.equal(grouped.children[4].hidden,false);assert.equal(nav.children[3].focused,true);assert(!grouped.children.some(p=>p.children.includes(profile)));
console.log('Category tabs preserve controls/handlers, switch panels and support keyboard navigation.');
