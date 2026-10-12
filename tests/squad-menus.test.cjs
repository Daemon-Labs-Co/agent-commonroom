const test=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
function harness(){
 const winEvents={},docEvents={},classes=new Set();
 function menu(top){
  const css=new Set(),attrs={},button={focused:false,focus(){this.focused=true}},summary={isConnected:true,focused:false,setAttribute(k,v){attrs[k]=v},getBoundingClientRect(){return {right:380,top,bottom:top+28}},focus(){this.focused=true}};
  const panel={style:{},classList:{add(n){css.add(n)},remove(n){css.delete(n)}},getBoundingClientRect(){return {width:160,height:140}},querySelector(){return button},addEventListener(k,fn){this[k]=fn},contains(n){return n===this||n===button}};
  const details={open:false,querySelector(s){return s==='summary'?summary:panel},removeAttribute(){this.open=false},contains(n){return n===summary||n===panel||n===button}};
  return {details,summary,panel,button,attrs,css};
 }
 const one=menu(60),two=menu(420);
 const ctx={view:'Room',state:{featureFlags:{}},render(){},document:{body:{classList:{toggle(n,on){if(on)classes.add(n);else classes.delete(n)}}},querySelectorAll(){return [one.details,two.details]},addEventListener(k,fn){docEvents[k]=fn}},window:{innerWidth:400,innerHeight:500,addEventListener(k,fn){winEvents[k]=fn}}};
 vm.createContext(ctx);vm.runInContext(fs.readFileSync('web/squad-menus.js','utf8'),ctx);return {ctx,one,two,classes,winEvents,docEvents};
}
test('opening another menu closes the first and bottom menus open upward',()=>{
 const h=harness();h.ctx.drawSquadMenus();h.one.summary.onclick({preventDefault(){}});assert.equal(h.one.details.open,true);assert.equal(h.one.panel.style.top,'94px');
 h.two.summary.onclick({preventDefault(){}});assert.equal(h.one.details.open,false);assert.equal(h.one.attrs['aria-expanded'],'false');assert.equal(h.two.details.open,true);assert.equal(h.two.panel.style.top,'274px');assert.equal(h.two.panel.style.left,'220px');
 h.two.summary.onclick({preventDefault(){}});assert.equal(h.two.details.open,false);
});
test('keyboard, outside click, scrolling and resize dismiss without changing actions',()=>{
 const h=harness();h.ctx.drawSquadMenus();h.two.summary.onkeydown({key:'ArrowDown',preventDefault(){}});assert.equal(h.two.button.focused,true);
 h.winEvents.keydown({key:'Escape',preventDefault(){}});assert.equal(h.two.details.open,false);assert.equal(h.two.summary.focused,true);
 h.one.summary.onclick({preventDefault(){}});h.docEvents.pointerdown({target:h.one.panel});assert.equal(h.one.details.open,true);h.docEvents.pointerdown({target:{}});assert.equal(h.one.details.open,false);
 h.one.summary.onclick({preventDefault(){}});h.winEvents.scroll({target:h.one.panel});assert.equal(h.one.details.open,true);h.winEvents.scroll({target:{}});assert.equal(h.one.details.open,false);
 h.one.summary.onclick({preventDefault(){}});h.winEvents.resize();assert.equal(h.one.details.open,false);
});
test('disabled flag preserves native behavior and navigation cleans up active menus',()=>{
 const h=harness();h.ctx.state.featureFlags.squadMenuLayout=false;h.ctx.drawSquadMenus();assert.equal(h.one.summary.onclick,undefined);assert.equal(h.classes.size,0);
 h.ctx.state.featureFlags.squadMenuLayout=true;h.ctx.drawSquadMenus();h.one.summary.onclick({preventDefault(){}});h.ctx.view='Settings';h.ctx.drawSquadMenus();assert.equal(h.one.details.open,false);assert.equal(h.one.panel.style.top,'');assert.equal(h.classes.size,0);
});
