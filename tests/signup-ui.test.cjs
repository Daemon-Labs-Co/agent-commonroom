'use strict';
const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const source=fs.readFileSync('web/signup.js','utf8');
const html=fs.readFileSync('web/signup.html','utf8');

// The page must work under the strict policy the server sends: no inline script or style.
assert.ok(!/<script(?![^>]*\bsrc=)[^>]*>/i.test(html),'inline script');
assert.ok(!/\sstyle=|<style/i.test(html),'inline style');
assert.ok(!/\son[a-z]+=/i.test(html),'inline event handler');

function node(){return{value:'',textContent:'',hidden:false,disabled:false,href:'',handlers:{},addEventListener(n,f){this.handlers[n]=f}}}
function harness({config={domain:'rooms.example.com',codeRequired:false},signup}){
 const ids=['signup-form','signup-name','signup-id','signup-code','signup-error','signup-domain','signup-code-field','signup-submit','signup-done','signup-link','signup-expiry'];
 const els=Object.fromEntries(ids.map(i=>[i,node()]));
 els['signup-error'].hidden=true;els['signup-done'].hidden=true;
 const calls=[];
 const context={
  console,JSON,Date,String,
  document:{getElementById:i=>els[i]},
  fetch:async(url,opts)=>{calls.push([url,opts&&JSON.parse(opts.body)]);
   if(url==='/api/signup/config')return{ok:true,json:async()=>config};
   const r=signup();return{ok:r.status<400,status:r.status,text:async()=>r.body}}
 };
 vm.createContext(context);vm.runInContext(source,context);
 return{els,calls};
}
const tick=()=>new Promise(r=>setImmediate(r));
const submit=async h=>{await h.els['signup-form'].handlers.submit({preventDefault(){}});await tick()};

(async()=>{
 // The address follows the name until it is edited, and is a valid label.
 let h=harness({signup:()=>({status:201,body:'{}'})});await tick();
 assert.equal(h.els['signup-domain'].textContent,'.rooms.example.com');
 assert.equal(h.els['signup-code-field'].hidden,true);
 h.els['signup-name'].value='  Acme & Sons, Ltd! ';h.els['signup-name'].handlers.input();
 assert.equal(h.els['signup-id'].value,'acme-sons-ltd');
 h.els['signup-id'].value='mine';h.els['signup-id'].handlers.input();
 h.els['signup-name'].value='Other';h.els['signup-name'].handlers.input();
 assert.equal(h.els['signup-id'].value,'mine','an edited address must not be overwritten');

 // The code field appears only when the server asks for a code.
 h=harness({config:{domain:'rooms.example.com',codeRequired:true},signup:()=>({status:201,body:'{}'})});await tick();
 assert.equal(h.els['signup-code-field'].hidden,false);

 // Success shows the link and hides the form.
 h=harness({signup:()=>({status:201,body:JSON.stringify({setupLink:'https://acme.rooms.example.com/#setup=abc',setupExpires:'2026-11-19T10:00:00Z'})})});await tick();
 h.els['signup-name'].value='Acme';h.els['signup-name'].handlers.input();
 await submit(h);
 const sent=h.calls.find(c=>c[0]==='/api/signup');
 assert.deepEqual(sent[1],{name:'Acme',id:'acme',code:''});
 assert.equal(h.els['signup-link'].href,'https://acme.rooms.example.com/#setup=abc');
 assert.equal(h.els['signup-form'].hidden,true);assert.equal(h.els['signup-done'].hidden,false);

 // A refusal is shown and the form stays usable.
 h=harness({signup:()=>({status:409,body:'That address is taken; choose another\n'})});await tick();
 h.els['signup-name'].value='Acme';h.els['signup-name'].handlers.input();
 await submit(h);
 assert.equal(h.els['signup-error'].textContent,'That address is taken; choose another');
 assert.equal(h.els['signup-error'].hidden,false);assert.equal(h.els['signup-submit'].disabled,false);assert.equal(h.els['signup-done'].hidden,true);

 // Nothing is sent without a name and address.
 h=harness({signup:()=>({status:201,body:'{}'})});await tick();
 await submit(h);
 assert.ok(!h.calls.some(c=>c[0]==='/api/signup'));
 assert.equal(h.els['signup-error'].hidden,false);
 console.log('Sign-up page: address suggestion, code field, success, refusal and empty-form handling passed.');
})().catch(e=>{console.error(e);process.exit(1)});
