const fs=require('node:fs'),vm=require('node:vm'),assert=require('node:assert/strict');
const esc=s=>String(s??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
let renders=0,opened=0;
const person={id:'dev',name:'Dev',role:'Developer',enabled:true,roomIds:['dev-room'],publicKey:'DO-NOT-RENDER-KEY'};
const ctx=vm.createContext({esc,view:'Room',render(){renders++},form(){},openSquadHumanProfile(){opened++},state:{currentActorId:'dev',featureFlags:{companyOrgChart:false},company:{name:'A <company>'},actors:[person],rooms:[{id:'dev-room',name:'Development'},{id:'qa-room',name:'Restricted QA'},{id:'personal',name:'My space',personalActorId:'dev'},{id:'other',name:'Private other',personalActorId:'other'}]}});
vm.runInContext(fs.readFileSync('web/human-profile-details.js','utf8'),ctx);
let html=ctx.humanAccountDetails(person,{name:'Development',humanRank:3});
assert(html.includes('A &lt;company&gt;'));assert(html.includes('Developer'));assert(html.includes('Development'));assert(html.includes('My space'));
assert(!html.includes('Restricted QA'));assert(!html.includes('Private other'));assert(!html.includes('DO-NOT-RENDER-KEY'));assert(html.includes('Decision seniority</dt><dd>11'));
html=ctx.humanAccountDetails({...person,id:'local-owner'}, {name:'Dev',humanPresent:false,humanRank:4});assert(html.includes('Restricted QA'));assert(html.includes('Away'));assert(html.includes('Decision seniority</dt><dd>4'));assert(!html.includes('Private other'));
let inserted='';ctx.appendHumanAccountDetails({insertAdjacentHTML(position,value){inserted=value}},person);assert(inserted.includes('Account and rooms'));
ctx.state.featureFlags.humanProfileDetails=false;inserted='';ctx.appendHumanAccountDetails({insertAdjacentHTML(){throw Error('disabled wrote DOM')}},person);vm.runInContext('render()',ctx);assert.equal(renders,1);
ctx.state.featureFlags.humanProfileDetails=true;ctx.view='Settings';ctx.document={querySelectorAll(){return [{querySelector(selector){return selector==='h1,h2'?{textContent:'Your profile'}:null},insertAdjacentHTML(position,value){inserted=value}}]}};vm.runInContext('render()',ctx);assert(inserted.includes('Account status'));
ctx.$=()=>({querySelector(){return null},insertAdjacentHTML(position,value){inserted=value}});ctx.openSquadHumanProfile(person,{name:'Development'},null,null);assert.equal(opened,1);assert(inserted.includes('Current room'));
console.log('Human details expose authorized account/room information, escape text, omit keys, preserve profile actions and honor the flag.');
const independent=vm.createContext({render(){},form(){}});vm.runInContext(fs.readFileSync('web/human-profile-details.js','utf8'),independent);assert.equal(typeof independent.humanAccountDetails,'function');

ctx.state.featureFlags.companyOrgChart=true;html=ctx.humanAccountDetails(person,{name:'Dev'});assert(html.includes('Company reporting hierarchy'));assert(!html.includes('Decision seniority'));
