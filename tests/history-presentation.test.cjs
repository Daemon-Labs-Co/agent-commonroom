const test=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
function context(){const ctx={state:{featureFlags:{},actors:[{id:'human',name:'Demo owner'}],rooms:[{id:'engineering',name:'Engineering'}]},esc:s=>String(s??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]))};vm.createContext(ctx);vm.runInContext(fs.readFileSync('web/history-presentation.js','utf8'),ctx);return ctx}
const filters={category:'All categories',outcome:'all',routine:false,search:''};
test('routine updates can be inspected explicitly while failures remain prominent',()=>{
 const c=context(),routine={kind:'participant.session',summary:'Idle',presentation:{category:'Agents',routine:true,tone:'info'}},failed={kind:'agent.error',summary:'Turn failed',presentation:{category:'Agents',routine:false,tone:'error'}};
 assert.equal(c.auditMatchesFilters(routine,filters),false);assert.equal(c.auditMatchesFilters(routine,{...filters,routine:true}),true);assert.equal(c.auditMatchesFilters(failed,{...filters,outcome:'attention'}),true);assert.equal(c.auditMatchesFilters(failed,{...filters,category:'Memory'}),false);
});
test('search uses recorded actor and room context without claiming global search',()=>{
 const c=context(),e={summary:'Workflow execution',kind:'workflow.run.updated',presentation:{category:'Workflows',actorId:'human',roomId:'engineering',subject:'Search feature',tone:'warning'}};
 for(const query of ['Demo owner','Engineering','Search feature'])assert.equal(c.auditMatchesFilters(e,{...filters,search:query}),true);
 assert.equal(c.auditMatchesFilters(e,{...filters,search:'absent'}),false);
});
test('event cards show safe useful context with checksums tucked behind technical details',()=>{
 const c=context(),e={sequence:1,kind:'agent.error',at:'2026-10-12T15:00:00Z',summary:'Agent <script> failed',hash:'synthetic-hash',presentation:{category:'Agents',actor:'<img>',roomId:'engineering',tone:'error',outcome:'Failed',explanation:'Time limit exceeded',next:'Inspect work'}};
 const html=c.renderRefinedAuditEvent(e);assert.match(html,/audit-tone-error/);assert.match(html,/Actor: &lt;img&gt;/);assert.match(html,/Room: Engineering/);assert.match(html,/Technical details/);assert.match(html,/synthetic-hash/);assert(!html.includes('<script>'));
 assert.match(c.renderRefinedAuditEvent({...e,presentation:{}}),/were not recorded/);
});
test('daily grouping preserves source order and loaded statistics are explicit',()=>{
 const c=context(),events=[{sequence:2,kind:'agent.error',at:'2026-10-12T15:00:00Z',summary:'Failure',presentation:{tone:'error'}},{sequence:1,kind:'participant.session',at:'2026-10-11T15:00:00Z',summary:'Idle',presentation:{routine:true}}];
 const html=c.renderRefinedAuditList(events);assert.equal((html.match(/class="audit-day"/g)||[]).length,2);assert.equal(events[0].sequence,2);assert.match(c.auditLoadedOverview(events,[events[0]]),/loaded records/);
 c.state.featureFlags.historyRefinement=false;assert.equal(c.historyRefined(),false);
});
