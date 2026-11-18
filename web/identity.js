// Passkey, single sign-on and invitation sign-in (ADR 0006). This page only forwards what
// the browser's authenticator returns and navigates to the identity provider; the server
// verifies everything. Nothing here is a permission: the server enforces who may do what.
let idMethods=null;
function idB64Decode(text){let s=String(text).replace(/-/g,'+').replace(/_/g,'/');while(s.length%4)s+='=';const bin=atob(s);const out=new Uint8Array(bin.length);for(let i=0;i<bin.length;i++)out[i]=bin.charCodeAt(i);return out.buffer}
function idB64Encode(buffer){const bytes=new Uint8Array(buffer);let s='';for(const b of bytes)s+=String.fromCharCode(b);return btoa(s).replace(/\+/g,'-').replace(/\//g,'_').replace(/=+$/,'')}
function idCreationOptions(o){return {...o,challenge:idB64Decode(o.challenge),user:{...o.user,id:idB64Decode(o.user.id)},excludeCredentials:(o.excludeCredentials||[]).map(c=>({...c,id:idB64Decode(c.id)}))}}
function idRequestOptions(o){return {...o,challenge:idB64Decode(o.challenge)}}
function idAttestation(credential){const r=credential.response;return {attestationObject:idB64Encode(r.attestationObject),clientDataJSON:idB64Encode(r.clientDataJSON)}}
function idAssertion(credential){const r=credential.response;return {id:idB64Encode(credential.rawId),authenticatorData:idB64Encode(r.authenticatorData),clientDataJSON:idB64Encode(r.clientDataJSON),signature:idB64Encode(r.signature),userHandle:r.userHandle?idB64Encode(r.userHandle):''}}
function idPasskeysSupported(){return typeof PublicKeyCredential!=='undefined'&&!!navigator?.credentials?.create}
function idFriendlyError(error){
 if(error?.name==='NotAllowedError'||error?.name==='AbortError')return 'The passkey request was cancelled or timed out.';
 if(error?.name==='InvalidStateError')return 'This passkey is already registered.';
 return error?.message||'Something went wrong. Try again.';
}
const idReturnMessages={failed:'Single sign-on did not complete. Try again.',unlinked:'That identity is not linked to an account here. Ask the owner for an invitation, or sign in another way.',taken:'That identity is already linked to another account.',linked:'Single sign-on is now linked to your account.',invalid:'That invitation is no longer valid.'};
function idReturnResult(hash){const m=/^#signin=(failed|unlinked|taken|linked|invalid)$/.exec(hash||'');return m?m[1]:null}
function idInviteToken(hash){const m=/^#invite=([A-Za-z0-9_-]{40,64})$/.exec(hash||'');return m?m[1]:null}
function idClearHash(){try{history.replaceState(null,'',location.pathname+location.search)}catch(e){}}
// Unauthenticated calls use fetch directly: the shared api() treats a 401 as "sign in".
async function idPost(path,body){
 const r=await fetch('/api/'+path,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body||{})});
 if(!r.ok)throw Error((await r.text()).trim()||'Request failed');
 return r.json();
}
async function idLoadMethods(){
 try{const r=await fetch('/api/auth/methods',{cache:'no-store'});idMethods=r.ok?await r.json():{}}catch(e){idMethods={}}
 return idMethods;
}
async function idFinishSignIn(result){
 csrf=result.csrf;
 if(result.actorId!=='local-owner'){roomId='personal-'+result.actorId;try{localStorage.setItem('commonroom-room',roomId)}catch(e){}}
 await refresh();
}
async function idPasskeySignIn(){
 const options=await idPost('webauthn/login/options',{});
 const credential=await navigator.credentials.get({publicKey:idRequestOptions(options.publicKey)});
 if(!credential)throw Error('No passkey was provided.');
 const result=await idPost('webauthn/login/verify',{challengeId:options.challengeId,...idAssertion(credential)});
 await idFinishSignIn(result);
}
async function idDecorateLogin(){
 const form=document.querySelector('#owner-login');
 if(!form||document.querySelector('#identity-signin')||document.querySelector('#main h1')?.textContent!=='Sign in to CommonRoom')return;
 const methods=await idLoadMethods();
 const passkeys=methods.passkeys&&idPasskeysSupported(),sso=methods.oidc;
 const result=idReturnResult(location.hash);
 if(!passkeys&&!sso&&!result)return;
 const box=document.createElement('div');box.id='identity-signin';box.className='identity-signin';
 box.innerHTML=(passkeys||sso?'<p class="muted">Or sign in with</p><div class="actions">'+(passkeys?'<button type="button" id="passkey-signin">Passkey</button>':'')+(sso?'<a class="button" id="sso-signin" href="/api/oidc/login">'+esc(methods.oidcName||'Single sign-on')+'</a>':'')+'</div>':'')+'<p id="identity-error" role="status"></p>';
 form.insertAdjacentElement('afterend',box);
 if(result){document.querySelector('#identity-error').textContent=idReturnMessages[result];idClearHash()}
 const button=document.querySelector('#passkey-signin');
 if(button)button.onclick=async()=>{button.disabled=true;document.querySelector('#identity-error').textContent='';try{await idPasskeySignIn()}catch(e){document.querySelector('#identity-error').textContent=idFriendlyError(e)}finally{button.disabled=false}};
}
async function idShowInvitation(token){
 const main=$('#main');
 try{$('#modal').close()}catch(e){}
 $('#nav').innerHTML='';main.innerHTML='<div class="card login-card"><h1>Checking your invitation…</h1></div>';
 let info;
 try{info=await idPost('invitations/lookup',{token})}catch(e){
  main.innerHTML='<div class="card login-card"><h1>This invitation is not valid</h1><p>It has expired, was already used or was withdrawn. Ask the person who invited you for a new link.</p><p><a href="/" id="invite-back">Go to sign in</a></p></div>';
  idClearHash();return;
 }
 const passkeys=info.passkeys&&idPasskeysSupported(),sso=info.oidc;
 main.innerHTML='<div class="card login-card"><h1>Join this workspace</h1><p>You have been invited as <strong>'+esc(info.role)+'</strong>. Enter your name and choose how you will sign in. You can add more ways later.</p>'
  +'<form id="invite-form"><label class="field">Your name<input name="name" required maxlength="100" autocomplete="name"></label><div class="actions">'
  +(passkeys?'<button class="primary" id="invite-passkey">Create a passkey</button>':'')
  +(sso?'<button type="button" id="invite-sso">Continue with '+esc(info.oidcName||'single sign-on')+'</button>':'')+'</div>'
  +(!passkeys&&!sso?'<p role="alert">This browser cannot create a passkey and no other sign-in method is set up. Try a different browser or ask the owner.</p>':'')
  +'<p id="invite-error" role="status"></p></form></div>';
 const form=$('#invite-form'),error=$('#invite-error');
 const nameOf=()=>String(new FormData(form).get('name')||'').trim();
 form.onsubmit=async e=>{
  e.preventDefault();
  const name=nameOf();if(!name){error.textContent='Enter your name.';return}
  const button=$('#invite-passkey');button.disabled=true;error.textContent='';
  try{
   const options=await idPost('invitations/passkey/options',{token,name});
   const credential=await navigator.credentials.create({publicKey:idCreationOptions(options.publicKey)});
   if(!credential)throw Error('No passkey was created.');
   const result=await idPost('invitations/passkey/complete',{challengeId:options.challengeId,...idAttestation(credential)});
   idClearHash();await idFinishSignIn(result);
  }catch(err){error.textContent=idFriendlyError(err);button.disabled=false}
 };
 const ssoButton=$('#invite-sso');
 if(ssoButton)ssoButton.onclick=()=>{const name=nameOf();if(!name){error.textContent='Enter your name.';return}idClearHash();location.href='/api/invitations/oidc/start?token='+encodeURIComponent(token)+'&name='+encodeURIComponent(name)};
}
if(typeof showLogin==='function'){
 const idShowLogin=showLogin;
 showLogin=async function(...args){
  const token=idInviteToken(location.hash);
  if(token)return idShowInvitation(token);
  await idShowLogin(...args);
  await idDecorateLogin();
 };
}
// Settings: your own sign-in methods, and for the owner, invitations and options.
function idOwnActor(){return state?.actors?.find(a=>a.id===state.currentActorId)||null}
function idCard(html){const card=document.createElement('section');card.className='card identity-card';card.innerHTML=html;document.querySelector('#main').append(card);return card}
async function idAddPasskey(){
 const options=await api('webauthn/register/options',{});
 const credential=await navigator.credentials.create({publicKey:idCreationOptions(options.publicKey)});
 if(!credential)throw Error('No passkey was created.');
 const name=(window.prompt?window.prompt('Name this passkey (for example "Work laptop")','Passkey'):'Passkey')||'Passkey';
 await api('webauthn/register/verify',{challengeId:options.challengeId,...idAttestation(credential),name:name.trim().slice(0,60)||'Passkey'});
 await refresh();
}
function idDrawOwnMethods(){
 const methods=idMethods||{},me=idOwnActor();
 if(!me||(!methods.passkeys&&!methods.oidc&&!(me.credentials||[]).length))return;
 const creds=me.credentials||[],links=me.oidcLinks||[];
 const card=idCard('<h2>Your sign-in methods</h2><p class="muted">Ways you can sign in to this workspace. Only you see this list.</p>'
  +(methods.passkeys?'<h3>Passkeys</h3>'+(creds.length?'<ul class="id-list">'+creds.map(c=>'<li data-credential="'+esc(c.id)+'"><span>'+esc(c.name)+' <span class="muted">added '+esc(new Date(c.created).toLocaleDateString())+'</span></span> <button type="button" data-remove-passkey="'+esc(c.id)+'">Remove</button></li>').join('')+'</ul>':'<p class="muted">No passkeys yet.</p>')+(idPasskeysSupported()?'<button type="button" id="add-passkey">Add a passkey</button>':'<p class="muted">This browser cannot create passkeys.</p>'):'')
  +(methods.oidc?'<h3>'+esc(methods.oidcName||'Single sign-on')+'</h3>'+(links.length?'<p>Linked'+(links[0].email?' as '+esc(links[0].email):'')+'.</p>':'<p class="muted">Not linked.</p><a class="button" id="link-sso" href="/api/oidc/link">Link '+esc(methods.oidcName||'single sign-on')+'</a>'):'')
  +'<p id="identity-settings-error" role="status"></p>');
 const error=card.querySelector('#identity-settings-error');
 const add=card.querySelector('#add-passkey');
 if(add)add.onclick=async()=>{add.disabled=true;error.textContent='';try{await idAddPasskey()}catch(e){error.textContent=idFriendlyError(e)}finally{add.disabled=false}};
 card.querySelectorAll('[data-remove-passkey]').forEach(button=>button.onclick=async()=>{button.disabled=true;error.textContent='';try{await api('webauthn/remove',{credentialId:button.dataset.removePasskey});await refresh()}catch(e){error.textContent=idFriendlyError(e);button.disabled=false}});
}
function idDrawOwnerControls(){
 if(state.currentActorId!=='local-owner')return;
 const methods=idMethods||{};
 const rooms=(state.rooms||[]).filter(r=>!r.personalActorId);
 const open=(state.invitations||[]).filter(i=>i.status==='open'&&new Date(i.expires)>new Date());
 const card=idCard('<h2>Sign-in options and invitations</h2>'
  +'<h3>Sign-in options</h3><label class="field"><input type="checkbox" data-id-flag="passkeySignIn" '+(methods.passkeys?'checked':'')+'> Passkeys</label>'
  +'<label class="field"><input type="checkbox" data-id-flag="oidcSignIn" '+(methods.oidc?'checked':'')+(methods.oidcConfigured?'':' disabled')+'> '+esc(methods.oidcName||'Single sign-on')+(methods.oidcConfigured?'':' <span class="muted">(not set up by the operator of this installation)</span>')+'</label>'
  +'<label class="field"><input type="checkbox" data-id-flag="memberInvitations" '+(methods.invitations?'checked':'')+'> Invitation links</label>'
  +'<p class="muted">Turning a method off hides it and refuses it on the server; saved passkeys and links are kept. The key file always works for the owner.</p>'
  +(methods.invitations?'<h3>Invite someone</h3><form id="invite-create"><label class="field">Role<select name="role"><option>Developer</option><option>QA</option></select></label>'
   +'<fieldset class="field"><legend>Rooms they can enter</legend>'+(rooms.map(r=>'<label><input type="checkbox" name="roomIds" value="'+esc(r.id)+'" checked> '+esc(r.name)+'</label>').join('')||'<span class="muted">No shared rooms yet.</span>')+'</fieldset>'
   +'<label class="field">Link valid for<select name="hours"><option value="24">1 day</option><option value="72" selected>3 days</option><option value="168">7 days</option></select></label>'
   +'<button class="primary">Create invitation link</button></form><div id="invite-result" hidden></div>'
   +(open.length?'<h3>Unused invitations</h3><ul class="id-list">'+open.map(i=>'<li><span>'+esc(i.role)+' <span class="muted">expires '+esc(new Date(i.expires).toLocaleString())+'</span></span> <button type="button" data-revoke-invitation="'+esc(i.id)+'">Revoke</button></li>').join('')+'</ul>':'')
   :'')
  +'<p id="identity-owner-error" role="status"></p>');
 const error=card.querySelector('#identity-owner-error');
 card.querySelectorAll('[data-id-flag]').forEach(box=>box.onchange=async()=>{box.disabled=true;error.textContent='';try{await commit('feature.flag.updated',{name:box.dataset.idFlag,enabled:box.checked});await idLoadMethods();await refresh()}catch(e){error.textContent=idFriendlyError(e);box.checked=!box.checked}finally{box.disabled=false}});
 const form=card.querySelector('#invite-create');
 if(form)form.onsubmit=async e=>{
  e.preventDefault();const data=new FormData(form);const button=form.querySelector('button');button.disabled=true;error.textContent='';
  try{
   const made=await api('invitations',{role:data.get('role'),roomIds:data.getAll('roomIds'),hours:Number(data.get('hours'))});
   const result=card.querySelector('#invite-result');result.hidden=false;
   result.innerHTML='<p>Send this link to the person. It works once and anyone who has it can use it, so share it privately.</p><input id="invite-link" readonly value="'+esc(made.link)+'"> <button type="button" id="copy-invite">Copy</button>';
   result.querySelector('#copy-invite').onclick=async()=>{try{await navigator.clipboard.writeText(made.link);note('Invitation link copied')}catch(err){result.querySelector('#invite-link').select()}};
  }catch(err){error.textContent=idFriendlyError(err)}finally{button.disabled=false}
 };
 card.querySelectorAll('[data-revoke-invitation]').forEach(button=>button.onclick=async()=>{button.disabled=true;error.textContent='';try{await api('invitations/revoke',{id:button.dataset.revokeInvitation});await refresh()}catch(e){error.textContent=idFriendlyError(e);button.disabled=false}});
}
function drawIdentitySettings(){
 if(view!=='Settings'||!state?.currentActorId||document.querySelector('.identity-card'))return;
 // The first time, fetch what this workspace offers and render again so the cards are placed
 // with the other settings (the tab grouping runs after this wrapper).
 if(!idMethods){idLoadMethods().then(()=>{if(view==='Settings'&&!document.querySelector('.identity-card'))render()});return}
 idDrawOwnMethods();idDrawOwnerControls();
}
{
 const idRender=render;
 render=function(){idRender();drawIdentitySettings()};
 const result=idReturnResult(location.hash);
 if(result==='linked'&&typeof note==='function'){note(idReturnMessages.linked);idClearHash()}
}
