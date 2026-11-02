'use strict';
const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const app=fs.readFileSync('web/app.js','utf8');
const formSource=app.slice(app.indexOf('function form('),app.indexOf('function field('));
const safetySource=fs.readFileSync('web/agent-safety.js','utf8').split('const safetyRender=')[0];
async function run(mode){
 const button={disabled:false,textContent:'',classList:{add(){}}};
 const error={hidden:true,textContent:''};
 const dialog={open:false,innerHTML:'',close(){this.open=false},showModal(){this.open=true},querySelector(selector){return selector==='#agent-safety-error'?error:button}};
 const form={querySelector:()=>button},cancel={},close={},notices=[];
 let accepted=0,refreshes=0;
 const context=vm.createContext({document:{querySelector:s=>s==='#modal button.primary'?button:s==='#agent-safety-error'?error:dialog},$:s=>s==='#modal'?dialog:s==='#cancel'?cancel:s==='#modal .modal-close'?close:s==='#modal form'?form:button,
  FormData:function(){return new Map([['reason','synthetic test']])},note:message=>notices.push(message),
  api:async()=>{if(mode==='rejected')throw Error('Worker shutdown is unverified; execution remains stopped');accepted++;return {ok:true}},
  refresh:async()=>{refreshes++;if(mode==='refresh-failed')throw Error('Status refresh unavailable')}
 });
 vm.runInContext(formSource+safetySource,context);
 context.agentSafetyForm('application','resume');
 assert.equal(button.textContent,'Resume dispatch');
 await form.onsubmit({preventDefault(){},target:form});
 return {button,error,dialog,notices,accepted,refreshes};
}
(async()=>{
 let result=await run('refresh-failed');
 assert.equal(result.accepted,1);
 assert.equal(result.dialog.open,false,'Confirmed resume must close its dialog even when status refresh fails');
 assert.match(result.notices.join(' '),/refresh/i);
 result=await run('rejected');
 assert.equal(result.accepted,0);
 assert.equal(result.refreshes,0);
 assert.equal(result.dialog.open,true,'Rejected resume must not look successful');
 assert.match(result.dialog.innerHTML,/role="alert"/);
 assert.equal(result.error.hidden,false,'Resume error must be visible inside the dialog');
 assert.match(result.error.textContent,/Worker shutdown is unverified/);
 assert.equal(result.button.disabled,false);
 result=await run('success');
 assert.equal(result.dialog.open,false);
 assert.equal(result.refreshes,1);
 console.log('Resume success closes, refresh failure closes with warning, and rejection stays open with inline reason.');
})().catch(error=>{console.error(error);process.exitCode=1});
