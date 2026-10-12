const test=require('node:test'),assert=require('node:assert/strict'),vm=require('node:vm'),fs=require('node:fs');
function list(initialTop=1500,shift=0){
 let scroll=initialTop;const obj={scrollHeight:2000,clientHeight:500,get scrollTop(){return scroll},set scrollTop(n){scroll=Math.max(0,Math.min(n,this.scrollHeight-this.clientHeight))},getBoundingClientRect(){return {top:100}}};
 obj.rows=Array.from({length:20},(_,i)=>({dataset:{messageId:'m'+i},getBoundingClientRect(){const top=100+i*100+shift-obj.scrollTop;return {top,bottom:top+100}}}));obj.querySelectorAll=()=>obj.rows;return obj;
}
function fixture(newest=false,enabled=true){
 let current=list(),nextShift=0;const classes=new Set();
 const ctx={view:'Room',roomId:'engineering',state:{messages:[{id:'latest'}],featureFlags:{roomViewportFit:enabled,newestMessagesFirst:newest}},window:{scrollTo(){}},document:{querySelector(s){return s==='.message-list'?current:{}},body:{classList:{contains(n){return classes.has(n)},toggle(n,on){if(on)classes.add(n);else classes.delete(n)}}}},render(){
  const old=current.scrollTop;current=list(0,nextShift);current.clientHeight=550;current.scrollTop=old;current.clientHeight=500;
 }};
 vm.createContext(ctx);vm.runInContext(fs.readFileSync('web/room-viewport.js','utf8'),ctx);return {ctx,current:()=>current,setShift(n){nextShift=n}};
}
test('status refresh restores bottom after the intermediate larger panel clamps it',()=>{
 const f=fixture();f.ctx.render();assert.equal(f.current().scrollTop,1500);f.ctx.render();assert.equal(f.current().scrollTop,1500,'must not jump 50px upward');
});
test('older-message reading is anchored after layout changes instead of forced to the end',()=>{
 const f=fixture();f.ctx.render();f.current().scrollTop=920;f.setShift(60);f.ctx.render();assert.equal(f.current().scrollTop,980,'same message remains at the same visible offset');
});
test('new messages and room changes follow the chosen edge in both orders',()=>{
 for(const newest of [false,true]){const f=fixture(newest);f.ctx.render();f.current().scrollTop=700;f.ctx.state.messages=[{id:'arrived'}];f.ctx.render();assert.equal(f.current().scrollTop,newest?0:1500);f.current().scrollTop=400;f.ctx.roomId='qa';f.ctx.render();assert.equal(f.current().scrollTop,newest?0:1500)}
});
test('top-first status refresh preserves an older message and disabling restores legacy behavior',()=>{
 const f=fixture(true);f.ctx.render();f.current().scrollTop=320;f.setShift(40);f.ctx.render();assert.equal(f.current().scrollTop,360);
 const off=fixture(false,false);off.ctx.render();assert.equal(off.current().scrollTop,1450,'disabled module leaves legacy clamping unchanged');
});
