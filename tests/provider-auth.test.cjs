const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const ctx=vm.createContext({render(){},state:{currentActorId:'local-owner',featureFlags:{connectionSignIn:false}},view:'Connections',document:{querySelector(){throw Error('Disabled UI touched DOM')}},api:async()=>({credentialPresent:true})});
vm.runInContext(fs.readFileSync('web/provider-auth.js','utf8'),ctx);
(async()=>{
 ctx.render();ctx.state.featureFlags.connectionSignIn=true;ctx.state.currentActorId='developer';ctx.render();
 const called=[];ctx.connectionSignOut=c=>called.push('out:'+c.id);ctx.connectionSignIn=c=>called.push('in:'+c.id);
 const b={isConnected:true};await ctx.checkConnectionAuthButton({id:'c'},b);assert.equal(b.textContent,'Sign out');assert.equal(b.disabled,false);b.onclick();assert.deepEqual(called,['out:c']);
 ctx.api=async()=>({credentialPresent:false});await ctx.checkConnectionAuthButton({id:'c'},b);assert.equal(b.textContent,'Sign in');b.onclick();assert.equal(called[1],'in:c');
 ctx.api=async()=>({credentialPresent:true,recoveryRequired:true});await ctx.checkConnectionAuthButton({id:'c'},b);assert.equal(b.textContent,'Sign in');assert.match(b.title,/recovery required/);b.onclick();assert.equal(called.at(-1),'in:c');
 ctx.api=async()=>({credentialPresent:true,recoveryRequired:'invalid'});await ctx.checkConnectionAuthButton({id:'c'},b);assert.equal(b.textContent,'Retry sign-in check');
 ctx.api=async()=>{throw Error('Worker unavailable')};await ctx.checkConnectionAuthButton({id:'c'},b);assert.equal(b.textContent,'Retry sign-in check');assert.match(b.title,/unavailable/);assert.equal(b.disabled,false);
 ctx.api=async()=>({});await ctx.checkConnectionAuthButton({id:'c'},b);assert.equal(b.textContent,'Retry sign-in check');
 let calledAPI=0;ctx.api=async()=>{calledAPI++;return {credentialPresent:true}};const gone={isConnected:false};await ctx.checkConnectionAuthButton({id:'c'},gone);assert.equal(calledAPI,0);
 ctx.api=async()=>{b.isConnected=false;return {credentialPresent:true}};b.isConnected=true;b.textContent='Starting';await ctx.checkConnectionAuthButton({id:'c'},b);assert.equal(b.textContent,'Checking sign-in…');
 let active=0,max=0;ctx.api=async()=>{active++;max=Math.max(max,active);await Promise.resolve();active--;return {credentialPresent:true}};
 await ctx.refreshConnectionAuthButtons(Array.from({length:10},(_,i)=>[{id:String(i)},{isConnected:true}]));assert.ok(max<=4);assert.ok(max>1);
 console.log('Sign-in/out selection, unknown/error states, detached views, flag/owner presentation and bounded concurrency passed.');
})().catch(e=>{console.error(e);process.exitCode=1});
