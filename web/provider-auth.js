// A saved credential is unusable while its durable custody is unresolved.
function connectionAuthUsable(result){
 if(typeof result.credentialPresent!=='boolean'||(result.recoveryRequired!==undefined&&typeof result.recoveryRequired!=='boolean'))throw Error('Invalid sign-in status');
 return result.credentialPresent&&result.recoveryRequired!==true;
}
function connectionSignOut(connection){
 form('Sign out of '+connection.name,'<p>Clear the saved worker credentials for this connection? Other connections and chat history remain intact.</p><p class="muted">Finish or stop active agents first. This removes CommonRoom’s stored login; it does not revoke the provider account token or erase existing agent session volumes.</p>',async()=>{await api('provider-auth',{action:'logout',connectionId:connection.id});render();note('Connection signed out')});$('#modal .primary').textContent='Sign out';
}
async function checkConnectionAuthButton(connection,button){
 if(!button.isConnected)return;button.disabled=true;button.textContent='Checking sign-in…';
 try{const result=await api('provider-auth',{action:'status',connectionId:connection.id});const usable=connectionAuthUsable(result);if(!button.isConnected)return;button.textContent=usable?'Sign out':'Sign in';button.title=result.recoveryRequired?'Credential recovery required: finish or reconcile prior work, then sign in again':usable?'Credentials saved; use Test connection to verify provider access':'This connection needs sign-in';button.onclick=()=>usable?connectionSignOut(connection):connectionSignIn(connection)}catch(error){if(!button.isConnected)return;button.textContent='Retry sign-in check';button.title=error.message;button.onclick=()=>checkConnectionAuthButton(connection,button)}finally{if(button.isConnected)button.disabled=false}
}
async function refreshConnectionAuthButtons(entries){let next=0;await Promise.all(Array.from({length:Math.min(4,entries.length)},async()=>{while(next<entries.length){const [connection,button]=entries[next++];await checkConnectionAuthButton(connection,button)}}))}
function validClaudeLoginURL(value){
 try{const url=new URL(value);return url.protocol==='https:'&&!url.username&&!url.password&&!url.port&&!url.hash&&['claude.ai','console.anthropic.com','platform.claude.com'].includes(url.hostname)&&url.pathname==='/oauth/authorize'&&url.searchParams.get('response_type')==='code'&&!!url.searchParams.get('code_challenge')}catch{return false}
}
async function readProviderLoginStream(response,onEvent){
 const reader=response.body.getReader(),decoder=new TextDecoder();let buffer='',complete=false;
 const consume=line=>{if(!line.trim())return;const event=JSON.parse(line);if(event.type==='error')throw Error(event.message);if(event.type==='instructions'&&(event.verificationUrl!=='https://auth.openai.com/codex/device'||!/^([A-Z0-9]{4,8}-[A-Z0-9]{4,8})$/.test(event.userCode)))throw Error('Invalid provider sign-in instructions');if(event.type==='authorization'&&(!validClaudeLoginURL(event.verificationUrl)||!/^[a-f0-9]{32}$/.test(event.loginId)))throw Error('Invalid Claude sign-in instructions');if(event.type==='complete')complete=true;onEvent(event)};
 try{for(;;){const {done,value}=await reader.read();buffer+=decoder.decode(value,{stream:!done});if(buffer.length>32768)throw Error('Sign-in response is too large');let boundary;while((boundary=buffer.indexOf('\n'))>=0){consume(buffer.slice(0,boundary));buffer=buffer.slice(boundary+1)}if(done){consume(buffer);break}}if(!complete)throw Error('Sign-in interrupted before completion')}finally{await reader.cancel().catch(()=>{});reader.releaseLock()}
}
async function runProviderBrowserLogin(connection,onEvent,signal){
 const send=()=>fetch('/api/provider-auth',{method:'POST',headers:{'Content-Type':'application/json','Accept':'application/x-ndjson','X-CSRF-Token':csrf},body:JSON.stringify({action:'login',connectionId:connection.id}),signal});
 let response=await send();if(!response.ok){const message=await response.text();if(response.status===401){showLogin();throw Error('Sign in required')}if(response.status===403&&message.trim()==='Invalid request token'){const session=await fetch('/api/session');if(!session.ok)throw Error('Please refresh the page to reconnect');csrf=(await session.json()).csrf;response=await send();if(!response.ok)throw Error(await response.text())}else throw Error(message)}
 if(!(response.headers.get('Content-Type')||'').includes('application/x-ndjson'))throw Error('The server did not return browser sign-in instructions. Refresh and try again.');
 await readProviderLoginStream(response,onEvent);
}
function connectionSignIn(connection){
 const isClaude=connectionProvider(connection)==='claude',provider=isClaude?'Claude':'Codex',website=isClaude?'Claude':'OpenAI';
 const instructions=isClaude?'Open the official Claude authorization link below. After approving access there, paste the authorization code from Claude into this dialog.':'Open the official OpenAI verification link below and enter the displayed one-time code on that page.';
 const modal=$('#modal');modal.classList.remove('evidence-drawer');
 modal.innerHTML='<button type="button" class="modal-close" aria-label="Close dialog">×</button><h2>Sign in to '+provider+'</h2><p>Connection: '+esc(connection.name)+'</p><p>'+instructions+' No terminal is needed on your browser computer.</p><p>Your password is entered only on the provider’s website. The resulting credentials are saved privately for this connection.</p><section id="provider-login-instructions" hidden><a id="provider-verification-link" target="_blank" rel="noopener noreferrer">Open '+website+' authorization ↗</a><p id="provider-user-code-row">Your one-time code: <strong id="provider-user-code"></strong></p></section>'+
 (isClaude?'<section id="provider-authorization-entry" hidden><label class="field">Authorization code from Claude<input id="provider-authorization-code" type="password" autocomplete="off" spellcheck="false" maxlength="2048"></label><button type="button" id="submit-provider-code">Submit authorization code</button></section>':'')+
 '<div class="actions"><button type="button" id="launch-provider-login" class="primary" disabled>Start sign-in</button><button type="button" id="check-provider-login">Check sign-in</button><button type="button" id="cancel-provider-login">Cancel</button></div><p id="provider-login-status" role="status">Checking connection…</p>';
 const status=modal.querySelector('#provider-login-status'),launch=modal.querySelector('#launch-provider-login'),checkButton=modal.querySelector('#check-provider-login'),close=modal.querySelector('#cancel-provider-login');
 let controller,loginID='',finished=false,running=false;
 const finish=()=>{finished=true;launch.hidden=true;checkButton.hidden=true;modal.querySelector('#provider-login-instructions').hidden=true;if(isClaude)modal.querySelector('#provider-authorization-entry').hidden=true;close.textContent='Done';modal.querySelector('h2').textContent='Signed in to '+provider;status.textContent='Sign-in saved for this connection. Use Test connection to verify provider access.';render()};
 modal.querySelector('.modal-close').onclick=()=>modal.close();close.onclick=()=>modal.close();modal.addEventListener('close',()=>controller?.abort(),{once:true});modal.showModal();
 const check=async()=>{if(running||finished)return;checkButton.disabled=true;try{const result=await api('provider-auth',{action:'status',connectionId:connection.id});const usable=connectionAuthUsable(result);if(!modal.open||finished||running)return;if(usable){finish();return}status.textContent=result.recoveryRequired?'Sign-in needs to be completed again before this connection can run.':'This connection needs sign-in. Choose Start sign-in.'}catch(e){if(modal.open&&!finished&&!running)status.textContent=e.message}finally{if(!finished&&!running){checkButton.disabled=false;launch.disabled=false}}};
 if(isClaude){
  const input=modal.querySelector('#provider-authorization-code'),submit=modal.querySelector('#submit-provider-code');
  submit.onclick=async()=>{const code=input.value.trim();input.value='';if(!code){status.textContent='Paste the authorization code from Claude first.';return}submit.disabled=true;input.disabled=true;try{await api('provider-auth',{action:'login-code',connectionId:connection.id,loginId:loginID,code});if(modal.open&&!finished){modal.querySelector('#provider-authorization-entry').hidden=true;status.textContent='Authorization code submitted. Waiting for Claude to confirm sign-in…'}}catch(error){if(modal.open&&!finished){status.textContent=error.message;submit.disabled=false;input.disabled=false}}};
  input.onkeydown=e=>{if(e.key==='Enter'&&!submit.disabled){e.preventDefault();submit.onclick()}};
 }
 checkButton.onclick=check;
 launch.onclick=async()=>{
  if(running||finished)return;controller=new AbortController();running=true;launch.disabled=true;checkButton.disabled=true;loginID='';modal.querySelector('#provider-login-instructions').hidden=true;if(isClaude)modal.querySelector('#provider-authorization-entry').hidden=true;status.textContent='Starting sign-in on the worker…';
  try{
   await runProviderBrowserLogin(connection,event=>{
    if(!modal.open)return;
    if(event.type==='starting')status.textContent='Starting the isolated worker and requesting '+website+' sign-in instructions…';
    else if(event.type==='waiting'&&modal.querySelector('#provider-login-instructions').hidden)status.textContent='Waiting for '+website+' sign-in instructions. This attempt will stop if they do not arrive within 45 seconds.';
    else if(event.type==='instructions'||event.type==='authorization'){
     modal.querySelector('#provider-verification-link').href=event.verificationUrl;modal.querySelector('#provider-login-instructions').hidden=false;modal.querySelector('#provider-user-code-row').hidden=isClaude;
     if(isClaude){loginID=event.loginId;modal.querySelector('#provider-authorization-entry').hidden=false;modal.querySelector('#provider-authorization-code').disabled=false;modal.querySelector('#submit-provider-code').disabled=false;status.textContent='Approve access on Claude’s page, then paste its authorization code below.'}
     else {modal.querySelector('#provider-user-code').textContent=event.userCode;status.textContent='Open the verification link and enter this code there. Waiting for you to finish…'}
    }else if(event.type==='complete')finish();
   },controller.signal);
  }catch(error){if(modal.open&&!finished){status.textContent=error.name==='AbortError'?'Sign-in cancelled.':error.message;modal.querySelector('#provider-login-instructions').hidden=true;if(isClaude)modal.querySelector('#provider-authorization-entry').hidden=true}}
  finally{running=false;if(!finished){launch.disabled=false;checkButton.disabled=false}}
 };
 check();
}

const providerAuthRender=render;render=function(){providerAuthRender();if(view==='Settings'&&state.currentActorId==='local-owner'){
 $('#main').insertAdjacentHTML('beforeend','<div class="card"><h3>Security · Connection sign-in</h3><p>Launch isolated provider authentication from Connections.</p><button id="edit-signin-flag">'+(state.featureFlags?.connectionSignIn===false?'Enable':'Disable')+' connection sign-in</button></div>');$('#edit-signin-flag').onclick=async()=>{await commit('feature.flag.updated',{name:'connectionSignIn',enabled:state.featureFlags?.connectionSignIn===false});};
 }if(view!=='Connections'||state.currentActorId!=='local-owner'||state.featureFlags?.connectionSignIn===false)return;const entries=[];for(const c of state.connections){if(c.protocol!=='CLI'||!/^ssh:\/\/[A-Za-z0-9_.-]+(?:\/claude)?$/.test(c.endpoint||''))continue;const edit=document.querySelector('button[data-connection="'+c.id+'"]');if(!edit)continue;const button=document.createElement('button');button.textContent='Checking sign-in…';button.disabled=true;edit.parentElement.append(button);entries.push([c,button]);}refreshConnectionAuthButtons(entries);};
