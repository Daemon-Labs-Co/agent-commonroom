// Paid access is verified server-side. These controls save preferences only;
// archive/model integrations and payment checkout are not implemented yet.
const premiumFeatureLabels={protectedAuditArchive:['Security','Protected audit archive'],managedOrchestration:['Workflows','Managed orchestration']};
function premiumControlState(info){
 const requested=info?.requested===true,entitled=info?.entitled===true;
 return {requested,disabled:!entitled&&!requested,message:entitled?(info.available?'Paid access active.':'Paid access active. Integration not implemented yet.'):(info?.status==='expired'?'Entitlement expired.':info?.status==='invalid'?'Entitlement could not be verified.':info?.status==='not_yet_valid'?'Entitlement is not active yet.':'Purchase required. Checkout and pricing are not configured.')};
}
function drawPremiumFeatures(){
 if(view!=='Settings'||state?.currentActorId!=='local-owner')return;
 for(const [flag,[category,label]] of Object.entries(premiumFeatureLabels)){
  const info=state.premiumFeatures?.[flag],controlState=premiumControlState(info);
  const card=document.createElement('section');card.className='card';
  const heading=document.createElement('h3');heading.textContent=category+' · '+label;
  const badge=document.createElement('p');badge.className='muted';badge.textContent='Premium add-on · Monthly, quarterly or yearly';
  const description=document.createElement('p');description.textContent=controlState.message;
  const row=document.createElement('div');row.className='feature-switch-row';
  const text=document.createElement('span');text.className='feature-switch-label';text.textContent='Enable when available';
  const status=document.createElement('span');status.className='feature-switch-state';status.textContent=controlState.requested?'On (saved preference)':'Off';status.setAttribute('aria-live','polite');
  const button=document.createElement('button');button.type='button';button.className='feature-switch';button.disabled=controlState.disabled;button.setAttribute('role','switch');button.setAttribute('aria-label',label);button.setAttribute('aria-checked',String(controlState.requested));
  const knob=document.createElement('span');knob.className='feature-switch-knob';knob.setAttribute('aria-hidden','true');button.append(knob);
  button.onclick=async()=>{
   if(button.disabled||state.currentActorId!=='local-owner')return;
   button.disabled=true;button.setAttribute('aria-busy','true');
   try{await api('events',{kind:'feature.flag.updated',payload:{name:flag,enabled:!controlState.requested}});try{sessionStorage.setItem('commonroom-feature-panel',category)}catch{};await refresh()}
   catch(error){note(error.message);button.disabled=controlState.disabled;button.removeAttribute('aria-busy')}
  };
  row.append(text,status,button);card.append(heading,badge,description,row);
  if(flag==='protectedAuditArchive'){const allowance=document.createElement('p');allowance.className='muted';allowance.textContent='Includes 100,000 records per month. Add-on blocks: 100,000 records each, price Y (not set).'+(info?.entitled?' Licensed allowance: '+Number(info.monthlyRecordAllowance||100000).toLocaleString()+' records per month.':'');card.append(allowance)}
  const help=document.createElement('button');help.type='button';help.className='context-help help-icon';help.textContent='?';help.setAttribute('aria-label','Help with '+label);help.onclick=()=>openDocumentation('premium-features');card.append(help);
  $('#main').append(card);
 }
}
const premiumFeatureRender=render;render=function(){premiumFeatureRender();drawPremiumFeatures()};
