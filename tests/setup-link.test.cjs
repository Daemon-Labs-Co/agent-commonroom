'use strict';
const fs=require('node:fs'),vm=require('node:vm'),assert=require('node:assert/strict');
const source=fs.readFileSync('web/app.js','utf8');
const login=source.slice(source.indexOf('let loginRequest=0;'),source.indexOf('function showOwnerSetup()'));
const code='ab'.repeat(32);
async function screen({setupRequired=true,expired=false,hash='',useSavedKey=false}={}){
 const nodes={},replaced=[];
 const element=s=>nodes[s]??=(s==='#modal'?{close(){}}:s==='#main'?{innerHTML:'',classList:{remove(){}}}:{innerHTML:'',textContent:'',value:''});
 const context=vm.createContext({state:null,csrf:null,$:element,esc:s=>String(s).replace(/[<>&]/g,''),
  location:{hash,pathname:'/',search:''},history:{replaceState:(...a)=>replaced.push(a)},
  fetch:async()=>({ok:true,json:async()=>({setupRequired,setupCodeExpired:expired})}),ownerSignIn:async()=>({ok:false,text:async()=>'x'}),refresh:async()=>{}});
 vm.runInContext(login,context);await context.showLogin(useSavedKey);
 return {html:nodes['#main'].innerHTML,field:nodes['#owner-login textarea'],replaced};
}
(async()=>{
 // Wording no longer points at a file on the server.
 let r=await screen();
 assert.doesNotMatch(r.html,/\.data\/owner-key|on this server/);
 assert.match(r.html,/setup link or setup code you were given/);
 assert.match(r.html,/hosted service sends you a setup link/);
 assert.match(r.html,/commonroom --setup-link/);
 assert.doesNotMatch(r.html,/has expired/);
 assert.equal(r.field?.value??'',''); // nothing prefilled without a link
 // A setup link fills the code in and removes it from the address bar.
 r=await screen({hash:'#setup='+code});
 assert.equal(r.field.value,code);
 assert.equal(r.replaced.length,1);assert.equal(r.replaced[0][2],'/');
 // Anything that is not exactly a 64-character code is ignored.
 for(const hash of ['#setup=short','#setup='+code+'0','#setup='+code.toUpperCase(),'#other='+code,'#setup=<img src=x>']){
  r=await screen({hash});assert.equal(r.field?.value??'','',hash);assert.equal(r.replaced.length,0);
 }
 // The link only applies to first-owner setup, never to a returning sign-in.
 r=await screen({setupRequired:false,hash:'#setup='+code});
 assert.match(r.html,/Sign in to CommonRoom/);assert.equal(r.replaced.length,0);
 r=await screen({hash:'#setup='+code,useSavedKey:true});
 assert.match(r.html,/Sign in to CommonRoom/);assert.equal(r.replaced.length,0);
 // Expired codes say so and tell the owner what to do.
 r=await screen({expired:true});
 assert.match(r.html,/role="alert">This setup code has expired/);assert.match(r.html,/restart the service/);
 r=await screen({setupRequired:false,expired:true});
 assert.doesNotMatch(r.html,/has expired/);
 const makefile=fs.readFileSync('Makefile','utf8');assert.ok(makefile.includes('tests/setup-link.test.cjs'));
 console.log('Setup link prefill, neutral setup wording and expired-code message passed.');
})().catch(error=>{console.error(error);process.exitCode=1});
