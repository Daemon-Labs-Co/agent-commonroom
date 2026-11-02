'use strict';
const fs=require('node:fs'),vm=require('node:vm'),assert=require('node:assert/strict');
const source=fs.readFileSync('web/app.js','utf8');
const login=source.slice(source.indexOf('let loginRequest=0;'),source.indexOf('function showOwnerSetup()'));
async function screen(setupRequired,useSavedKey=false,available=true){
 const nodes={};
 const element=s=>nodes[s]??=(s==='#modal'?{close(){}}:s==='#main'?{innerHTML:'',classList:{remove(){}}}:{innerHTML:'',textContent:''});
 const context=vm.createContext({state:null,csrf:null,$:element,esc:s=>String(s).replace(/[<>&]/g,''),fetch:async()=>({ok:available,json:async()=>({setupRequired})}),ownerSignIn:async()=>({ok:false,text:async()=> 'Invalid owner key'}),refresh:async()=>{throw Error('unexpected refresh')},showOwnerSetup(){throw Error('unexpected owner creation')}});
 vm.runInContext(login,context);
 await context.showLogin(useSavedKey);
 return {context,nodes};
}
(async()=>{
 let result=await screen(true);
 assert.match(result.nodes['#main'].innerHTML,/Set up your owner account/);
 assert.match(result.nodes['#main'].innerHTML,/You do not need an existing PEM file/);
 assert.match(result.nodes['#main'].innerHTML,/Server setup code/);
 for(const html of [result.nodes['#main'].innerHTML,fs.readFileSync('web/index.html','utf8')]){
   assert.match(html,/<a[^>]+href="\/help\/\?doc=user-getting-started"[^>]+target="_blank"[^>]+rel="noopener noreferrer"/);
   assert.match(html,/class="external-link-icon" aria-hidden="true"/);
   assert.match(html,/opens in a new tab or window/);
 }
 result=await screen(false);
 assert.match(result.nodes['#main'].innerHTML,/Sign in to CommonRoom/);
 assert.doesNotMatch(result.nodes['#main'].innerHTML,/Set up owner access/);
 const form={querySelector:()=>({disabled:false}),elements:{key:{value:'invalid synthetic credential'}}};
 await result.nodes['#owner-login'].onsubmit({preventDefault(){},target:form});
 assert.equal(result.nodes['#login-error'].textContent,'Invalid owner key');
 assert.doesNotMatch(result.nodes['#main'].innerHTML,/Set up owner access/);
 result=await screen(true,true);
 assert.match(result.nodes['#main'].innerHTML,/Sign in to CommonRoom/);
 result=await screen(false,false,false);
 assert.match(result.nodes['#main'].innerHTML,/Sign-in unavailable/);
 assert.doesNotMatch(result.nodes['#main'].innerHTML,/<form/);
 const label={textContent:''};
 vm.runInNewContext(fs.readFileSync('web/instance-label.js','utf8'),{document:{querySelector:()=>label},window:{location:{host:'agentcommonroom.home.arpa'}}});
 assert.equal(label.textContent,'Server name unavailable');
 console.log('First-owner setup, returning sign-in, failed-key refusal, setup-status failure and static-preview hostname fallback passed.');
})().catch(error=>{console.error(error);process.exitCode=1;});
