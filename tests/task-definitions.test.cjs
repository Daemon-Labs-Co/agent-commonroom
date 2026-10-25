'use strict';
const {test}=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
function harness({owner=true,enabled=true,view='Tasks'}={}){
 let html='',save,title='',legacyEdits=0,apiResult={executions:[],nextCursor:''};const nodes={},calls=[];
 const esc=s=>String(s??'').replaceAll('&','&amp;').replaceAll('<','&lt;').replaceAll('"','&quot;');
 const insert=(_,text)=>{html+=text;for(const m of text.matchAll(/id="([^"]+)"/g))nodes['#'+m[1]]={insertAdjacentHTML:insert,remove(){this.removed=true}}};
 const ctx=vm.createContext({state:{currentActorId:owner?'local-owner':'member',featureFlags:enabled?{taskDefinitions:true}:{},tasks:[]},roomId:'engineering',view,render(){},taskForm(){legacyEdits++},taskCard(){return '<article>Existing card</article>'},wireTasks(){},dispatchTask:async()=>{calls.push(['legacy-dispatch'])},$:s=>s==='#main'?{insertAdjacentHTML:insert}:nodes[s],document:{querySelectorAll(){return []},getElementById(){return {remove(){}}}},esc,field:(label,name,value='')=>'<label>'+label+'<input name="'+name+'" value="'+esc(value)+'" required></label>',selectField:(label,name,options,value)=>'<label>'+label+'<select name="'+name+'">'+options.map(o=>'<option '+(o===value?'selected':'')+'>'+o+'</option>').join('')+'</select></label>',memberField:()=>'',form(t,body,fn){title=t;html+=body;save=fn;insert('',body)},commit:async(...args)=>calls.push(args),api:async(...args)=>{calls.push(args);return apiResult},refresh:async()=>{},note:m=>calls.push(['notice',m])});
 vm.runInContext(fs.readFileSync('web/task-definitions.js','utf8'),ctx);
 return{ctx,nodes,calls,html:()=>html,title:()=>title,legacyEdits:()=>legacyEdits,submit:data=>save(data),setResult:value=>apiResult=value};
}
function definitionForm(){return new Map(Object.entries({'title':'Synthetic task','td-task':'Specific synthetic work','td-process-mode':'objective','td-process':'Use declared inputs','td-process-constraints':'No external effects\nNo invented evidence','td-output-0-name':'result','td-output-0-type':'value','td-output-0-required':'on','td-output-0-bytes':'1024','td-output-0-schema':'text.v1','td-criterion-0-id':'correct','td-criterion-0-description':'Result satisfies the input','td-criterion-0-output':'result','td-criterion-0-method':'human','td-op-read':'on','td-duration':'60','td-response-bytes':'2048','td-tool-calls':'2','assignee':'dev','reviewer':'reviewer','status':'Assigned'}))}
test('owner editor captures all seven definition parts without execution bindings',async()=>{
 const h=harness();h.ctx.taskForm();for(const heading of ['1. Task','2. Expected input','3. Process','4. Desired output','5. Acceptance criteria','6. Execution constraints','7. Failure contract'])assert(h.html().includes(heading));await h.submit(definitionForm());const [kind,payload]=h.calls[0];assert.equal(kind,'task.revised');const d=payload.definition;assert.equal(d.task,'Specific synthetic work');assert.equal(d.expectedInput.length,0);assert.equal(d.desiredOutput[0].name,'result');assert.equal(d.process.constraints.length,2);assert.equal(d.executionConstraints.completionAuthority,'owner');assert.equal(d.failureContract.retry,'human-only');assert.equal(d.inputs,undefined);
});
test('flag off and members retain legacy edits while definitions remain read-only',()=>{
 const h=harness({enabled:false});h.ctx.taskForm();assert.equal(h.legacyEdits(),1);const d=h.ctx.tdDefault({title:'<scope>'});h.ctx.taskForm({id:'task',definition:d,status:'Assigned'});assert(h.html().includes('&lt;scope>'));assert(!h.html().includes('name="td-task"'));const member=harness({owner:false});member.ctx.taskForm({id:'task',definition:d,status:'Assigned'});assert(!member.html().includes('name="td-task"'));
});
test('concrete object, array, file and bundle parsing preserves finite contracts',()=>{
 const h=harness(),data=definitionForm();data.set('td-output-0-schema','object.v1');data.set('td-output-0-properties','count | integer.v1 | required\nlabel | text.v1 | optional');let c=h.ctx.tdParseContract(data,'td-output-0');assert.equal(c.properties[0].name,'count');assert.equal(c.properties[0].required,true);assert.equal(c.properties[1].required,false);
 data.set('td-output-0-schema','array.v1');data.set('td-output-0-item-schema','boolean.v1');data.set('td-output-0-max-items','4');c=h.ctx.tdParseContract(data,'td-output-0');assert.equal(c.itemSchemaId,'boolean.v1');assert.equal(c.maxItems,4);
 data.set('td-output-0-type','bundle');data.set('td-output-0-members','docs/guide.md | text/markdown | 512');c=h.ctx.tdParseContract(data,'td-output-0');assert.equal(c.members[0].name,'docs/guide.md');assert.equal(c.mediaType,'application/vnd.commonroom.bundle+json');
});
test('typed input binding preserves exact JSON numbers and source versions per execution',async()=>{
 const h=harness();const d=h.ctx.tdDefault({title:'Synthetic'});d.expectedInput=[{name:'count',type:'value',required:true,schemaId:'integer.v1',maxBytes:64}];const task={id:'task',roomId:'engineering',revision:2,definition:d};await h.ctx.dispatchTask(task,'work');await h.submit(new Map([['td-bind-0','9223372036854775807'],['td-version-0','synthetic-v1']]));const [path,request]=h.calls[0];assert.equal(path,'task-dispatch');assert.equal(request.inputs[0].valueJson,'9223372036854775807');assert.equal(request.inputs[0].sourceVersion,'synthetic-v1');assert.equal(request.definition,undefined);
});
test('file input claims cannot launch, settings default off and stay owner-only',async()=>{
 const h=harness();const d=h.ctx.tdDefault({title:'Synthetic'});d.expectedInput=[{name:'source',type:'file',required:true}];await assert.rejects(()=>h.ctx.dispatchTask({definition:d},'work'),/verified immutable storage/);assert.equal(h.calls.length,0);
 const settings=harness({view:'Settings',enabled:false});settings.ctx.drawTaskDefinitionSettings();assert(settings.html().includes('Task definitions · Off'));settings.nodes['#td-setting'].onclick();await settings.submit(new Map([['enabled','on']]));assert.equal(settings.calls[0][1].name,'taskDefinitions');const member=harness({view:'Settings',owner:false});member.ctx.drawTaskDefinitionSettings();assert.equal(member.html(),'');
});
test('history displays immutable exact values and escapes task/output text',async()=>{
 const h=harness();const d=h.ctx.tdDefault({title:'<task>'});h.setResult({executions:[{id:'attempt',mode:'work',status:'validated',definition:d,inputs:[],outputs:[{name:'result',digest:'sha',valueJson:'"<value>"'}]}],nextCursor:''});await h.ctx.showTaskExecutionHistory({id:'task'});assert(h.html().includes('&lt;value>'));assert(h.html().includes('Immutable output attempt/result'));assert(!h.html().includes('<task>'));
 const page=fs.readFileSync('web/index.html','utf8');assert(page.indexOf('src="task-definitions.js"')<page.indexOf('src="feature-flag-tabs.js"'));assert(page.indexOf('src="task-definitions.js"')<page.indexOf('src="settings-tabs.js"'));
});
test('accepted task input picker uses references and never submits preview or manual replacements',async()=>{
 const h=harness();h.ctx.state.featureFlags.taskValueDependencies=true;
 const d=h.ctx.tdDefault({title:'Dependent'});d.expectedInput=[{name:'source',type:'value',schemaId:'integer.v1',required:true,maxBytes:64}];
 h.setResult({sources:[{taskId:'source',taskTitle:'Build <web>',taskRevision:4,executionId:'attempt',outputName:'count',valueJson:'9223372036854775807',digest:'hash'}],nextCursor:''});
 await h.ctx.dispatchTask({id:'target',roomId:'engineering',revision:2,definition:d},'work');
 assert(h.html().includes('Build &lt;web>'));assert(h.html().includes('Provide a new value (manual)'));assert(h.calls[0][0].startsWith('task-input-sources?'));
 await h.submit(new Map([['td-source-0','0'],['td-bind-0','spoofed'],['td-version-0','fake']]));
 const request=h.calls.find(([path])=>path==='task-dispatch')[1];assert.equal(request.inputs[0].source.executionId,'attempt');assert.equal(request.inputs[0].source.outputName,'count');assert.equal(request.inputs[0].valueJson,undefined);assert.equal(request.inputs[0].sourceVersion,undefined);
 assert.throws(()=>h.ctx.tdBoundInputs(new Map([['td-source-0','99']]),d.expectedInput,[{sources:[]}]),/unavailable/);
});
test('dependency settings default off and history marks stale provenance without erasing values',async()=>{
 const h=harness({view:'Settings'});h.ctx.drawTaskValueDependencySettings();assert(h.html().includes('Accepted task inputs · Off'));h.nodes['#td-dependency-setting'].onclick();await h.submit(new Map([['enabled','on']]));assert.equal(h.calls[0][1].name,'taskValueDependencies');
 const member=harness({view:'Settings',owner:false});member.ctx.drawTaskValueDependencySettings();assert.equal(member.html(),'');
 const history=harness(),definition=history.ctx.tdDefault({title:'Target'});history.setResult({executions:[{id:'attempt',mode:'work',status:'validated',definition,inputsCurrent:false,outputs:[],inputs:[{name:'source',valueJson:'"<stored>"',sourceVersion:'source/result',digest:'hash',provenance:{taskId:'<source>',taskRevision:4,executionId:'source',outputName:'result'}}]}],nextCursor:''});await history.ctx.showTaskExecutionHistory({id:'task'});assert(history.html().includes('Stale or disabled'));assert(history.html().includes('Verified returned value from task &lt;source>'));assert(history.html().includes('&lt;stored>'));
});
test('late source discovery cannot open a binding dialog in a different room',async()=>{
 const h=harness();h.ctx.state.featureFlags.taskValueDependencies=true;let resolve;h.ctx.api=()=>new Promise(r=>resolve=r);const d=h.ctx.tdDefault({title:'Target'});d.expectedInput=[{name:'source',type:'value',required:true,schemaId:'text.v1'}];const pending=h.ctx.dispatchTask({id:'target',roomId:'engineering',revision:1,definition:d},'work');h.ctx.roomId='elsewhere';resolve({sources:[],nextCursor:''});await pending;assert.equal(h.html(),'');
});
test('criterion review removes generic acceptance and clearly reopens accepted task edits',()=>{
 const h=harness();h.ctx.state.featureFlags.taskAcceptanceReview=true;const d=h.ctx.tdDefault({title:'Existing'});h.ctx.taskForm({id:'task',status:'Accepted',definition:d});assert(h.html().includes('Saving edits reopens this accepted task'));assert(!h.ctx.tdTaskStatuses(true).includes('Accepted'));assert.equal(h.ctx.tdEditableStatus('Accepted'),'Changes requested');
});
test('definition revision history escapes retained contracts, paginates and ignores late room responses',async()=>{
 const h=harness({enabled:false}),d=h.ctx.tdDefault({title:'<old task>'});h.setResult({definitions:[{definition:d,digest:'<hash>'}],nextCursor:'1'});await h.ctx.showTaskDefinitionHistory({id:'task'});assert(h.html().includes('&lt;old task>'));assert(h.html().includes('&lt;hash>'));assert(h.html().includes('revisions never executed'));assert(h.nodes['#td-older-definitions']);await h.nodes['#td-older-definitions'].onclick();assert(h.calls[1][0].includes('&before=1'));assert(h.ctx.taskCard({id:'task',definition:d}).includes('data-task-definition-history'));
 const late=harness();let resolve;late.ctx.api=()=>new Promise(r=>resolve=r);const pending=late.ctx.showTaskDefinitionHistory({id:'task'});late.ctx.roomId='elsewhere';resolve({definitions:[{definition:d,digest:'hash'}]});await pending;assert.equal(late.html(),'');
});
test('execution configuration history escapes identifiers and distinguishes unrecorded legacy metadata',()=>{
 const h=harness();const html=h.ctx.tdExecutionConfiguration({configurationVersion:1,participantRevision:3,connectionId:'<connection>',connectionRevision:2,participantConfigurationDigest:'<profile>',connectionConfigurationDigest:'<fingerprint>'});assert(html.includes('Profile revision 3'));assert(html.includes('&lt;connection>'));assert(html.includes('&lt;profile>'));assert(!html.includes('<fingerprint>'));assert(html.includes('not proof of worker origin'));assert(h.ctx.tdExecutionConfiguration({}).includes('completed returned values remain history'));
});
