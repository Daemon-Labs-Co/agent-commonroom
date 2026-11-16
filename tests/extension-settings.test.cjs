const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const source=fs.readFileSync('web/extension-settings.js','utf8');
let renders=0,html='';const commits=[];const button={};
const ctx={view:'Settings',state:{currentActorId:'local-owner',featureFlags:{}},render:()=>{renders++},commit:async(k,p)=>commits.push([k,p]),form:(t,h,save)=>{ctx.save=save;ctx.formHtml=h},window:{location:{reload(){}}},
 $:sel=>sel==='#main'?{insertAdjacentHTML:(_,h)=>{html+=h}}:button,document:{querySelector:()=>null}};
vm.createContext(ctx);vm.runInContext(source,ctx);
const label=vm.runInContext('extensionSwitchLabel',ctx);
assert.equal(label({}),'Extensions · Off','off unless explicitly enabled');assert.equal(label({extensions:true}),'Extensions · On');
(async()=>{
 vm.runInContext('render()',ctx);assert.equal(renders,1);assert.match(html,/Security · Extensions/);assert.match(html,/cannot approve/);
 button.onclick();assert.doesNotMatch(ctx.formHtml,/checked/);
 await ctx.save({has:()=>true});
 assert.equal(JSON.stringify(commits),JSON.stringify([['feature.flag.updated',{name:'extensions',enabled:true}]]));
 html='';ctx.state.currentActorId='someone-else';vm.runInContext('render()',ctx);assert.equal(html,'','only the owner sees the switch');
 assert.ok(fs.readFileSync('web/index.html','utf8').includes('extension-settings.js'));
 console.log('Extension settings: owner-only switch, default off and flag request passed.');
})();
