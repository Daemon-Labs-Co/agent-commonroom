// Presentation only: flag changes still go through authenticated, owner-only
// event validation. Active security stops and immutable history are unaffected.
const featureSwitchPending=new Set();
const featureSwitchSpecial={
 'edit-safety-flag':['agentSafetyControls','Emergency agent controls'],
 'edit-photo-flag':['humanProfileImages','Human profile photos'],
 'edit-signin-flag':['connectionSignIn','Connection sign-in'],
 'edit-testing-flag':['connectionTesting','Connection testing']
};
function featureSwitchEnabled(flag,flags){return ['automaticAgentRooms','taskDelegation','taskProposals','agentCoordination','newestMessagesFirst','workflowExecution','workspaceDisposal'].includes(flag)?flags?.[flag]===true:flags?.[flag]!==false}
function makeFeatureSwitch(original,flag,label){
 const enabled=featureSwitchEnabled(flag,state.featureFlags);
 const row=document.createElement('div');row.className='feature-switch-row';
 const text=document.createElement('span');text.className='feature-switch-label';text.textContent=label;
 const control=document.createElement('button');control.type='button';control.className='feature-switch';control.dataset.featureFlag=flag;control.setAttribute('role','switch');control.setAttribute('aria-label',label);control.setAttribute('aria-checked',String(enabled));control.disabled=featureSwitchPending.has(flag);
 const knob=document.createElement('span');knob.className='feature-switch-knob';knob.setAttribute('aria-hidden','true');control.append(knob);
 const status=document.createElement('span');status.className='feature-switch-state';status.textContent=enabled?'On':'Off';status.setAttribute('aria-live','polite');
 row.append(text,status,control);original.replaceWith(row);
 control.onclick=async()=>{
  if(featureSwitchPending.has(flag)||state.currentActorId!=='local-owner')return;
  const before=featureSwitchEnabled(flag,state.featureFlags);featureSwitchPending.add(flag);control.disabled=true;control.setAttribute('aria-busy','true');control.setAttribute('aria-checked',String(!before));status.textContent='Saving…';let saved=false;
  try{
   await api('events',{kind:'feature.flag.updated',payload:{name:flag,enabled:!before}});saved=true;status.textContent=!before?'On':'Off';
   try{sessionStorage.setItem('commonroom-feature-panel',typeof activeFeatureCategory==='undefined'?'Security':activeFeatureCategory)}catch{}
   window.location.reload();
  }catch(error){control.setAttribute('aria-checked',String(before));status.textContent=before?'On':'Off';note(error.message)}
  finally{if(!saved){featureSwitchPending.delete(flag);control.disabled=false;control.removeAttribute('aria-busy')}}
 };
 return control;
}
function drawFeatureSwitches(){
 if(view!=='Settings'||state?.currentActorId!=='local-owner'||state.featureFlags?.featureFlagSwitches===false)return;
 for(const button of document.querySelectorAll('#main button')){
  if(button.classList.contains('feature-switch'))continue;
  const special=featureSwitchSpecial[button.id];
  const flag=button.dataset.featureFlag||special?.[0];if(!flag)continue;
  const label=special?.[1]||(flag==='helpWindowChoice'?'Help window choice':button.textContent.replace(/ · (On|Off)$/,''));
  makeFeatureSwitch(button,flag,label);
 }
}
const featureSwitchRender=render;render=function(){featureSwitchRender();drawFeatureSwitches()};

drawFeatureSwitches();
