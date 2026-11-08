const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const source=fs.readFileSync('web/features.js','utf8');
const start=source.indexOf('// Sign out stays reachable'),end=source.indexOf('// end sign-out navigation');
assert.ok(start>0&&end>start,'sign-out navigation block present');
const calls=[];let handler;let inserted=[];
const settings={insertAdjacentHTML:(where,html)=>inserted.push({where,html})};
const ctx={
 render(){calls.push('render')},
 document:{querySelector:q=>q==='#nav .nav-settings'?settings:null,getElementById:()=>inserted.length?{}:null,addEventListener:(t,f)=>{if(t==='click')handler=f}},
 api:async(path)=>{calls.push('api:'+path)},showLogin:()=>calls.push('showLogin'),note:m=>calls.push('note:'+m)
};
vm.createContext(ctx);vm.runInContext(source.slice(start,end),ctx);
ctx.render();
assert.equal(inserted.length,1);assert.equal(inserted[0].where,'afterend','button goes directly below Settings');
assert.ok(inserted[0].html.includes('id="nav-signout"')&&inserted[0].html.includes('Sign out'));
ctx.render();assert.equal(inserted.length,1,'rendering again does not duplicate the button');
(async()=>{
 calls.length=0;await handler({target:{closest:()=>null}});assert.deepEqual(calls,[],'other clicks are ignored');
 await handler({target:{closest:sel=>sel==='#nav-signout'?{}:null}});assert.deepEqual(calls,['api:logout','showLogin'],'click signs out then shows login');
 ctx.api=async()=>{throw new Error('Unable to persist sign-out; please retry')};calls.length=0;
 await handler({target:{closest:()=>({})}});assert.deepEqual(calls,['note:Unable to persist sign-out; please retry'],'failures are shown, not hidden');
 assert.ok(fs.readFileSync('web/style.css','utf8').includes('nav .nav-signout'));
 console.log('Sign out is added below Settings on every render and signs out once clicked.');
})();
