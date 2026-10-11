const fs=require('node:fs'),vm=require('node:vm'),assert=require('node:assert/strict');
const node=(text,classes=[])=>({textContent:text,title:'',attributes:{},classList:{values:new Set(classes),contains(v){return this.values.has(v)},add(v){this.values.add(v)}},getAttribute(k){return this.attributes[k]||null},setAttribute(k,v){this.attributes[k]=v}});
const help=node('Help with this form'),field=node('?', ['seniority-help']),nav=node('Help & documentation',['nav-help']);help.attributes['aria-label']='Help with this form';let clicks=0;help.onclick=()=>{clicks++};field.attributes['aria-label']='Help with decision seniority';
const modal=node('');modal.querySelector=()=>help;modal.querySelectorAll=()=>[];
const ctx=vm.createContext({state:{featureFlags:{}},form(){},render(){},document:{querySelectorAll(){return [help,field,nav]},querySelector(){return modal}}});
vm.runInContext(fs.readFileSync('web/help-icons.js','utf8'),ctx);
assert.equal(help.textContent,'?');assert.equal(help.title,'Help with this form');assert.equal(help.attributes['aria-label'],'Help with this form');assert(help.classList.contains('overlay-help'));assert(field.classList.contains('help-icon'));assert.equal(nav.textContent,'Help & documentation');help.onclick();assert.equal(clicks,1,'help routing must stay intact');
ctx.state.featureFlags.standardHelpIcons=false;help.textContent='Legacy help';vm.runInContext('standardizeHelpIcons()',ctx);assert.equal(help.textContent,'Legacy help');
console.log('Consistent overlay/field help icons preserve accessible labels, routing and disabled behavior.');

ctx.state=undefined;vm.runInContext('standardizeHelpIcons()',ctx);console.log('Help icons also tolerate the initial signed-out loading state.');

let wrapper;const info=node('Field guidance',['muted']);info.tagName='P';const fieldLabel={nextElementSibling:info,before(row){wrapper=row}};modal.querySelectorAll=()=>[fieldLabel];ctx.document.createElement=()=>({append(...items){this.items=items}});vm.runInContext('rightAlignFieldGuidance(document.querySelector())',ctx);assert.equal(wrapper.className,'field-guidance-row');assert.equal(wrapper.items[0],fieldLabel);assert.equal(wrapper.items[1],info);console.log('Field guidance is placed alongside its field.');
