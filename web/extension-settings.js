// Extensions are separate services listed by the server administrator. This
// switch only allows or refuses them; it grants no room and adds no extension.
function extensionSwitchLabel(flags){return 'Extensions · '+(flags?.extensions===true?'On':'Off')}
function drawExtensionSettings(){
 if(view!=='Settings'||state.currentActorId!=='local-owner'||document.querySelector('#configure-extensions'))return;
 const on=state.featureFlags?.extensions===true;
 $('#main').insertAdjacentHTML('beforeend','<div class="card"><h3>Security · Extensions</h3><p>Off by default. Allow services listed in the administrator’s registry file to read the messages of the rooms they were granted and post notices there under their own name. They cannot approve, send or accept work, change settings or sign in as a person. Turning this off refuses every extension request at once; posted notices are kept.</p><button data-feature-flag="extensions" id="configure-extensions">'+extensionSwitchLabel(state.featureFlags)+'</button></div>');
 $('#configure-extensions').onclick=()=>form('Extensions','<label><input type="checkbox" name="enabled" '+(on?'checked':'')+'> Allow registered extensions</label><p class="muted">Which services are registered, and which rooms each may use, is set in the registry file on the server, not here.</p>',async d=>{await commit('feature.flag.updated',{name:'extensions',enabled:d.has('enabled')});window.location.reload()});
}
const extensionSettingsRender=render;render=function(){extensionSettingsRender();drawExtensionSettings()};
