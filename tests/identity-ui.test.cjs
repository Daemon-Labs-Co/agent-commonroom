'use strict';
const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const source=fs.readFileSync('web/identity.js','utf8');

function el(registry){
 const node={innerHTML:'',textContent:'',value:'',hidden:false,disabled:false,dataset:{},children:[],handlers:{},className:'',id:''};
 node.append=c=>node.children.push(c);
 node.insertAdjacentElement=(_,c)=>{node.after=c;return c};
 node.querySelector=s=>registry[s]||null;
 node.querySelectorAll=s=>registry[s+'*']||[];
 node.select=()=>{node.selected=true};
 return node;
}
function harness(opts={}){
 const registry={},calls={fetch:[],api:[],commit:[],notes:[],refresh:0,replaced:0,navigated:null,prompts:0};
 const main=el(registry);registry['#main']=main;
 const responses=opts.responses||{};
 const context={
  console,Date,URLSearchParams,encodeURIComponent,Uint8Array,ArrayBuffer,TextEncoder,
  atob:s=>Buffer.from(s,'base64').toString('binary'),btoa:s=>Buffer.from(s,'binary').toString('base64'),
  state:opts.state||{currentActorId:'local-owner',actors:[],rooms:[],invitations:[]},view:opts.view||'Settings',csrf:null,roomId:'engineering',
  esc:s=>String(s).replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c])),
  note:t=>calls.notes.push(t),
  refresh:async()=>{calls.refresh++},
  commit:async(k,p)=>{calls.commit.push(JSON.parse(JSON.stringify([k,p])))},
  api:async(path,body)=>{calls.api.push(JSON.parse(JSON.stringify([path,body===undefined?null:body])));if(opts.apiError)throw Error(opts.apiError);return (opts.apiResponses||{})[path]||{}},
  $:s=>registry[s]||null,
  localStorage:{setItem(){}},
  location:{hash:opts.hash||'',pathname:'/',search:'',set href(v){calls.navigated=v}},
  history:{replaceState(){calls.replaced++}},
  window:{prompt:()=>{calls.prompts++;return opts.promptValue??'Work laptop'}},
  navigator:{credentials:{create:async o=>{calls.created=o;if(opts.createError)throw opts.createError;return opts.credential||fakeAttestation()},get:async o=>{calls.got=o;if(opts.getError)throw opts.getError;return fakeAssertion()}},clipboard:{writeText:async t=>{calls.copied=t}}},
  PublicKeyCredential:opts.noPasskeys?undefined:function(){},
  FormData:class{constructor(form){this.v=form.values||{}}get(k){return this.v[k]??null}getAll(k){return [].concat(this.v[k]??[])}},
  document:{querySelector:s=>registry[s]||null,createElement:()=>el(registry)},
  fetch:async(url,init)=>{
   calls.fetch.push([url,init&&init.body?JSON.parse(init.body):null,init&&init.method]);
   const r=responses[url];
   if(!r)return {ok:false,text:async()=>'not found',json:async()=>({})};
   return {ok:r.ok!==false,text:async()=>r.text||'',json:async()=>r.json||{}};
  },
  showLogin:async()=>{calls.baseLogin=(calls.baseLogin||0)+1},
  render(){calls.baseRender=(calls.baseRender||0)+1},
 };
 vm.createContext(context);vm.runInContext(source,context);
 return {ctx:context,calls,registry,main};
}
const bytes=(...n)=>new Uint8Array(n).buffer;
function fakeAttestation(){return {rawId:bytes(1,2,3),response:{attestationObject:bytes(250,251),clientDataJSON:bytes(7,8,9)}}}
function fakeAssertion(){return {rawId:bytes(1,2,3),response:{authenticatorData:bytes(4,5),clientDataJSON:bytes(7,8),signature:bytes(9,9,9),userHandle:bytes(65,66)}}}
const run=(h,code)=>vm.runInContext(code,h.ctx);

(async()=>{
 // Encoding round trips, including values whose standard base64 contains + / and padding.
 let h=harness();
 for(const raw of [[],[251,255,254],[1],[1,2],[1,2,3],[255,255,255,255],Array.from({length:70},(_,i)=>i*3%256)]){
  const encoded=h.ctx.idB64Encode(new Uint8Array(raw).buffer);
  assert.match(encoded,/^[A-Za-z0-9_-]*$/);
  assert.deepEqual([...new Uint8Array(h.ctx.idB64Decode(encoded))],raw);
 }
 const created=h.ctx.idCreationOptions({challenge:'AQID',user:{id:'BAU',name:'Ann'},excludeCredentials:[{type:'public-key',id:'AQID'}],rp:{id:'x'}});
 assert.deepEqual([...new Uint8Array(created.challenge)],[1,2,3]);assert.deepEqual([...new Uint8Array(created.user.id)],[4,5]);assert.deepEqual([...new Uint8Array(created.excludeCredentials[0].id)],[1,2,3]);assert.equal(created.rp.id,'x');
 assert.deepEqual(JSON.parse(JSON.stringify(h.ctx.idAssertion(fakeAssertion()))),{id:'AQID',authenticatorData:'BAU',clientDataJSON:'Bwg',signature:'CQkJ',userHandle:'QUI'});
 assert.equal(h.ctx.idAssertion({rawId:bytes(1),response:{authenticatorData:bytes(1),clientDataJSON:bytes(1),signature:bytes(1),userHandle:null}}).userHandle,'');
 assert.deepEqual(Object.keys(h.ctx.idAttestation(fakeAttestation())).sort(),['attestationObject','clientDataJSON']);

 // Only exact, well-formed fragments are acted on.
 const token='A'.repeat(43);
 assert.equal(h.ctx.idInviteToken('#invite='+token),token);
 for(const bad of ['#invite='+'A'.repeat(39),'#invite='+'A'.repeat(65),'#invite=<script>'+token,'#invite='+token+'!','#other='+token,'invite='+token,'',undefined,'#invite='+token+'\n'])assert.equal(h.ctx.idInviteToken(bad),null,String(bad));
 assert.equal(h.ctx.idReturnResult('#signin=unlinked'),'unlinked');
 for(const bad of ['#signin=evil','#signin=failed&x=1','#signin=','#SIGNIN=failed',''])assert.equal(h.ctx.idReturnResult(bad),null);
 assert.match(h.ctx.idFriendlyError({name:'NotAllowedError'}),/cancelled or timed out/);
 assert.match(h.ctx.idFriendlyError({name:'InvalidStateError'}),/already registered/);
 assert.equal(h.ctx.idFriendlyError(Error('Passkey sign-in failed')),'Passkey sign-in failed');

 // Passkey sign-in posts the options request then the authenticator's answer.
 h=harness({responses:{'/api/webauthn/login/options':{json:{challengeId:'c1',publicKey:{challenge:'AQID',rpId:'acme.example.com'}}},'/api/webauthn/login/verify':{json:{csrf:'tok',actorId:'member-1'}}}});
 await run(h,'idPasskeySignIn()');
 assert.deepEqual(h.calls.fetch.map(c=>c[0]),['/api/webauthn/login/options','/api/webauthn/login/verify']);
 assert.equal(h.calls.fetch[1][1].challengeId,'c1');assert.equal(h.calls.fetch[1][1].id,'AQID');
 assert.deepEqual([...new Uint8Array(h.calls.got.publicKey.challenge)],[1,2,3]);
 assert.equal(run(h,'csrf'),'tok');assert.equal(run(h,'roomId'),'personal-member-1');assert.equal(h.calls.refresh,1);
 h=harness({responses:{'/api/webauthn/login/options':{json:{challengeId:'c1',publicKey:{challenge:'AQID'}}},'/api/webauthn/login/verify':{ok:false,text:'Passkey sign-in failed'}}});
 await assert.rejects(run(h,'idPasskeySignIn()'),/Passkey sign-in failed/);assert.equal(h.calls.refresh,0);
 h=harness({getError:{name:'NotAllowedError'},responses:{'/api/webauthn/login/options':{json:{challengeId:'c1',publicKey:{challenge:'AQID'}}}}});
 await assert.rejects(run(h,'idPasskeySignIn()'));assert.equal(h.calls.fetch.length,1,'nothing is sent when the person cancels');

 // The sign-in page gains options only for a sign-in screen, only when offered, and shows return messages.
 h=harness({responses:{'/api/auth/methods':{json:{passkeys:true,oidc:true,oidcName:'Acme <SSO>'}}}});
 h.registry['#owner-login']=el(h.registry);h.registry['#main h1']={textContent:'Sign in to CommonRoom'};h.registry['#identity-error']={textContent:''};
 await run(h,'idDecorateLogin()');
 assert.ok(h.registry['#owner-login'].after,'options were added after the form');
 assert.match(h.registry['#owner-login'].after.innerHTML,/id="passkey-signin"/);assert.match(h.registry['#owner-login'].after.innerHTML,/href="\/api\/oidc\/login"/);assert.doesNotMatch(h.registry['#owner-login'].after.innerHTML,/<SSO>/,'provider name is escaped');
 for(const [methods,heading] of [[{passkeys:false,oidc:false},'Sign in to CommonRoom'],[{passkeys:true,oidc:true},'Set up your owner account']]){
  const g=harness({responses:{'/api/auth/methods':{json:methods}}});
  g.registry['#owner-login']=el(g.registry);g.registry['#main h1']={textContent:heading};
  await run(g,'idDecorateLogin()');assert.equal(g.registry['#owner-login'].after,undefined,'no options for '+heading+JSON.stringify(methods));
 }
 const unsupported=harness({noPasskeys:true,responses:{'/api/auth/methods':{json:{passkeys:true,oidc:false}}}});
 unsupported.registry['#owner-login']=el(unsupported.registry);unsupported.registry['#main h1']={textContent:'Sign in to CommonRoom'};
 await run(unsupported,'idDecorateLogin()');assert.equal(unsupported.registry['#owner-login'].after,undefined,'no passkey option in a browser without passkeys');
 const back=harness({hash:'#signin=unlinked',responses:{'/api/auth/methods':{json:{passkeys:false,oidc:false}}}});
 back.registry['#owner-login']=el(back.registry);back.registry['#main h1']={textContent:'Sign in to CommonRoom'};back.registry['#identity-error']={textContent:''};
 await run(back,'idDecorateLogin()');assert.match(back.registry['#identity-error'].textContent,/not linked to an account/);assert.equal(back.calls.replaced,1);

 // showLogin is wrapped: invitations take over the page; otherwise the normal sign-in runs.
 h=harness({hash:'#invite='+token,responses:{'/api/invitations/lookup':{ok:false,text:'This invitation is not valid'}}});
 h.registry['#modal']={close(){}};h.registry['#nav']={innerHTML:'x'};
 await run(h,'showLogin()');
 assert.match(h.main.innerHTML,/This invitation is not valid/);assert.equal(h.calls.baseLogin,undefined);assert.equal(h.calls.replaced,1);
 h=harness({responses:{'/api/auth/methods':{json:{}}}});await run(h,'showLogin()');assert.equal(h.calls.baseLogin,1);

 // Accepting an invitation with a passkey.
 const invitation={json:{valid:true,role:'<b>Developer</b>',passkeys:true,oidc:true,oidcName:'Acme SSO'}};
 h=harness({hash:'#invite='+token,responses:{'/api/invitations/lookup':invitation,'/api/invitations/passkey/options':{json:{challengeId:'c9',publicKey:{challenge:'AQID',user:{id:'BAU'}}}},'/api/invitations/passkey/complete':{json:{csrf:'tok2',actorId:'new-1'}}}});
 h.registry['#modal']={close(){}};h.registry['#nav']={innerHTML:''};
 const form=el(h.registry);form.values={name:'  Dana  '};
 h.registry['#invite-form']=form;h.registry['#invite-error']={textContent:''};h.registry['#invite-passkey']={disabled:false};h.registry['#invite-sso']={};
 await run(h,'showLogin()');
 assert.doesNotMatch(h.main.innerHTML,/<b>Developer/);assert.match(h.main.innerHTML,/&lt;b&gt;Developer/,'the role is escaped');assert.match(h.main.innerHTML,/id="invite-passkey"/);assert.match(h.main.innerHTML,/Continue with Acme SSO/);
 await form.onsubmit({preventDefault(){}});
 const posted=h.calls.fetch.filter(c=>c[0].includes('passkey'));
 assert.deepEqual(posted[0][1],{token,name:'Dana'});assert.equal(posted[1][1].challengeId,'c9');assert.ok(posted[1][1].attestationObject&&posted[1][1].clientDataJSON);
 assert.equal(run(h,'csrf'),'tok2');assert.equal(run(h,'roomId'),'personal-new-1');assert.equal(h.calls.refresh,1);
 // No name means nothing is sent; single sign-on carries the encoded token and name; a cancelled passkey can be retried.
 const empty=harness({hash:'#invite='+token,responses:{'/api/invitations/lookup':invitation}});
 empty.registry['#modal']={close(){}};empty.registry['#nav']={innerHTML:''};const eform=el(empty.registry);eform.values={name:'   '};
 empty.registry['#invite-form']=eform;empty.registry['#invite-error']={textContent:''};empty.registry['#invite-passkey']={disabled:false};empty.registry['#invite-sso']={};
 await run(empty,'showLogin()');await eform.onsubmit({preventDefault(){}});
 assert.equal(empty.registry['#invite-error'].textContent,'Enter your name.');assert.equal(empty.calls.fetch.filter(c=>c[0].includes('options')).length,0);
 empty.registry['#invite-sso'].onclick();assert.equal(empty.calls.navigated,null);
 eform.values={name:'Zoë & Co'};empty.registry['#invite-sso'].onclick();
 assert.equal(empty.calls.navigated,'/api/invitations/oidc/start?token='+token+'&name='+encodeURIComponent('Zoë & Co'));
 const cancelled=harness({hash:'#invite='+token,createError:{name:'NotAllowedError'},responses:{'/api/invitations/lookup':invitation,'/api/invitations/passkey/options':{json:{challengeId:'c',publicKey:{challenge:'AQID',user:{id:'BAU'}}}}}});
 cancelled.registry['#modal']={close(){}};cancelled.registry['#nav']={innerHTML:''};const cform=el(cancelled.registry);cform.values={name:'Dana'};
 const cbutton={disabled:false};cancelled.registry['#invite-form']=cform;cancelled.registry['#invite-error']={textContent:''};cancelled.registry['#invite-passkey']=cbutton;cancelled.registry['#invite-sso']={};
 await run(cancelled,'showLogin()');await cform.onsubmit({preventDefault(){}});
 assert.match(cancelled.registry['#invite-error'].textContent,/cancelled or timed out/);assert.equal(cbutton.disabled,false,'the button comes back so the person can retry');

 // Settings: your own methods, with names escaped.
 const me={id:'m1',kind:'human',credentials:[{id:'cred-1',name:'<img src=x onerror=alert(1)>',created:'2026-11-18T10:00:00Z'}],oidcLinks:[{email:'a@example.com'}]};
 h=harness({state:{currentActorId:'m1',actors:[me],rooms:[]},apiResponses:{'webauthn/register/options':{challengeId:'c5',publicKey:{challenge:'AQID',user:{id:'BAU',name:'Ann'}}}}});
 run(h,'idMethods={passkeys:true,oidc:true,oidcName:"Acme SSO"}');
 run(h,'idDrawOwnMethods()');
 let card=h.main.children[0];
 assert.match(card.innerHTML,/Your sign-in methods/);assert.doesNotMatch(card.innerHTML,/<img src=x/);assert.match(card.innerHTML,/data-remove-passkey="cred-1"/);assert.match(card.innerHTML,/Linked as a@example.com/);assert.match(card.innerHTML,/id="add-passkey"/);
 const remove={dataset:{removePasskey:'cred-1'},disabled:false};card.registry=h.registry;
 h.registry['[data-remove-passkey]*']=[remove];h.registry['#identity-settings-error']={textContent:''};h.registry['#add-passkey']={disabled:false};
 card=h.main.children[0];
 // The card is created before the handlers are attached, so draw again with the elements registered.
 h.main.children.length=0;run(h,'idDrawOwnMethods()');
 await remove.onclick();assert.deepEqual(h.calls.api[0],['webauthn/remove',{credentialId:'cred-1'}]);assert.equal(h.calls.refresh,1);
 await h.registry['#add-passkey'].onclick();
 assert.equal(h.calls.api[1][0],'webauthn/register/options');assert.equal(h.calls.api[2][0],'webauthn/register/verify');assert.equal(h.calls.api[2][1].name,'Work laptop');assert.equal(h.calls.api[2][1].challengeId,'c5');
 assert.ok(h.calls.prompts===1);
 // The first visit to Settings loads the methods and then renders again, so the cards are
 // drawn before the tab grouping runs rather than appended after it.
 const lazy=harness({state:{currentActorId:'m1',actors:[{id:'m1',kind:'human',credentials:[{id:'c',name:'N',created:'2026-11-18T10:00:00Z'}]}],rooms:[]},responses:{'/api/auth/methods':{json:{passkeys:true}}}});
 run(lazy,'drawIdentitySettings()');assert.equal(lazy.main.children.length,0,'nothing is drawn before the methods are known');
 await new Promise(r=>setTimeout(r,20));assert.equal(lazy.calls.baseRender,1,'a second render draws the cards in place');
 run(lazy,'drawIdentitySettings()');assert.equal(lazy.main.children.length>0,true);
 // Nothing is drawn when no method is offered and none is saved.
 const none=harness({state:{currentActorId:'m1',actors:[{id:'m1',kind:'human'}],rooms:[]}});run(none,'idMethods={}');run(none,'idDrawOwnMethods()');assert.equal(none.main.children.length,0);

 // Owner controls exist only for the owner and send exactly what was chosen.
 const rooms=[{id:'engineering',name:'Engineering'},{id:'personal-x',name:'Mine',personalActorId:'x'}];
 const future=new Date(Date.now()+3600e3).toISOString(),past=new Date(Date.now()-3600e3).toISOString();
 h=harness({state:{currentActorId:'local-owner',actors:[],rooms,invitations:[{id:'i1',role:'QA',status:'open',expires:future},{id:'i2',role:'QA',status:'open',expires:past},{id:'i3',role:'QA',status:'redeemed',expires:future}]},apiResponses:{invitations:{link:'https://acme.rooms.example.com/#invite='+token}}});
 run(h,'idMethods={passkeys:true,oidc:false,oidcConfigured:false,invitations:true}');
 run(h,'idDrawOwnerControls()');
 card=h.main.children[0];
 assert.match(card.innerHTML,/value="engineering"/);assert.doesNotMatch(card.innerHTML,/personal-x/,'personal rooms cannot be assigned');
 assert.match(card.innerHTML,/data-revoke-invitation="i1"/);assert.doesNotMatch(card.innerHTML,/data-revoke-invitation="i2"|data-revoke-invitation="i3"/,'only unused, unexpired invitations are listed');
 assert.match(card.innerHTML,/data-id-flag="oidcSignIn"[^>]*disabled/);assert.match(card.innerHTML,/not set up by the operator/);
 const member=harness({state:{currentActorId:'m1',actors:[],rooms:[],invitations:[]}});run(member,'idMethods={invitations:true}');run(member,'idDrawOwnerControls()');assert.equal(member.main.children.length,0,'members get no owner controls');
 // Wire the handlers on a second draw with registered elements.
 const create=el(h.registry);create.values={role:'QA',roomIds:['engineering'],hours:'24'};const createButton={disabled:false};create.querySelector=s=>s==='button'?createButton:null;
 const result=el(h.registry);const linkBox={select(){},value:''};const copy={};result.querySelector=s=>s==='#copy-invite'?copy:s==='#invite-link'?linkBox:null;
 const flag={dataset:{idFlag:'passkeySignIn'},checked:false,disabled:false};
 h.registry['#invite-create']=create;h.registry['#invite-result']=result;h.registry['#identity-owner-error']={textContent:''};h.registry['[data-id-flag]*']=[flag];
 const revoke={dataset:{revokeInvitation:'i1'},disabled:false};h.registry['[data-revoke-invitation]*']=[revoke];
 h.main.children.length=0;run(h,'idDrawOwnerControls()');
 await create.onsubmit({preventDefault(){}});
 assert.deepEqual(h.calls.api[0],['invitations',{role:'QA',roomIds:['engineering'],hours:24}]);assert.equal(result.hidden,false);assert.match(result.innerHTML,/#invite=/);
 await copy.onclick();assert.match(h.calls.copied,/#invite=/);assert.deepEqual(h.calls.notes,['Invitation link copied']);
 await flag.onchange();assert.deepEqual(h.calls.commit[0],['feature.flag.updated',{name:'passkeySignIn',enabled:false}]);
 await revoke.onclick();assert.deepEqual(h.calls.api[1],['invitations/revoke',{id:'i1'}]);
 // A refused change puts the checkbox back and says why.
 const refusing=harness({apiError:'Revoke an unused invitation before creating another',state:h.ctx.state});
 refusing.registry['#invite-create']=create;refusing.registry['#invite-result']=result;refusing.registry['#identity-owner-error']={textContent:''};
 run(refusing,'idMethods={invitations:true}');run(refusing,'idDrawOwnerControls()');await create.onsubmit({preventDefault(){}});
 assert.match(refusing.registry['#identity-owner-error'].textContent,/Revoke an unused invitation/);assert.equal(createButton.disabled,false);

 // Wiring in the real page: loaded after app.js and before the settings grouping that places its cards.
 const page=fs.readFileSync('web/index.html','utf8');
 assert.ok(page.indexOf('src="identity.js')>page.indexOf('src="app.js')&&page.indexOf('src="identity.js')<page.indexOf('src="settings-tabs.js'));
 const tabs=fs.readFileSync('web/settings-tabs.js','utf8');assert.match(tabs,/Your sign-in methods/);assert.match(tabs,/Sign-in options and invitations/);
 assert.ok(fs.readFileSync('Makefile','utf8').includes('tests/identity-ui.test.cjs'));
 console.log('Identity interface: encoding, sign-in page, invitations, passkey management and owner controls passed.');
})().catch(error=>{console.error(error);process.exitCode=1});
