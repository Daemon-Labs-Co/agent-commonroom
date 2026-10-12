const assert=require('node:assert/strict');const vm=require('node:vm');const fs=require('node:fs');const crypto=require('node:crypto');
const context={setInterval:()=>{},crypto,render:()=>{},window:{addEventListener:()=>{}},state:{participants:[{id:'reviewer'}]},structuredClone};vm.createContext(context);
vm.runInContext(fs.readFileSync('web/workflows.js','utf8'),context);vm.runInContext(fs.readFileSync('web/workflow-templates.js','utf8'),context);
const catalog=vm.runInContext('workflowTemplates',context);assert.equal(catalog.length,8);assert.equal(new Set(catalog.map(t=>t.id)).size,8);
for(const template of catalog){
 const first=context.buildWorkflowTemplate(template.id),second=context.buildWorkflowTemplate(template.id);
 assert.equal(first.id,undefined);assert.equal(first.revision,0);assert.equal(first.nodes.filter(n=>n.type==='start').length,1);assert.equal(first.nodes.filter(n=>n.type==='approval').length,1);
 assert.ok(first.nodes.every(n=>n.instructions.length>20));assert.ok(first.nodes.filter(n=>['agent','review'].includes(n.type)).every(n=>n.agentId===''));
 assert.ok(first.nodes.every(n=>!second.nodes.some(other=>other.id===n.id)));assert.ok(first.edges.every(e=>!second.edges.some(other=>other.id===e.id)));
 const ids=new Set(first.nodes.map(n=>n.id));for(const edge of first.edges){assert.ok(ids.has(edge.from)&&ids.has(edge.to));assert.ok(context.wfPorts(first.nodes.find(n=>n.id===edge.from)).includes(edge.port))}
 const review=first.nodes.find(n=>n.type==='review');assert.ok(first.edges.some(e=>e.from===review.id&&e.port==='changes'&&e.to===first.nodes[1].id));
 assert.ok(context.wfValidate(first).some(e=>e.includes('choose a room agent')));
 first.nodes.filter(n=>['agent','review'].includes(n.type)).forEach(n=>n.agentId='reviewer');assert.equal(context.wfValidate(first).length,0);
 first.nodes[0].title='modified';assert.notEqual(second.nodes[0].title,'modified');
}
const feature=context.buildWorkflowTemplate('feature-development');const instructions=feature.nodes.map(n=>n.instructions).join(' ');assert.match(instructions,/tests/);assert.match(instructions,/documentation/);assert.match(instructions,/feature flag/);
assert.throws(()=>context.buildWorkflowTemplate('missing'),/Unknown workflow template/);
console.log('All eight templates: fresh IDs, review loops, required assignments, valid graphs and delivery requirements passed.');

const documenter=context.buildWorkflowTemplate('documenter');const docText=documenter.nodes.map(n=>n.instructions).join(' ');assert.match(docText,/source/);assert.match(docText,/method\/class comments/);assert.match(docText,/Check links/);assert.match(docText,/Never publish secrets/);
const disaster=context.buildWorkflowTemplate('disaster-planning');const disasterText=disaster.nodes.map(n=>n.instructions).join(' ');assert.match(disasterText,/recovery time and recovery point objectives/);assert.match(disasterText,/tabletop exercise/);assert.match(disasterText,/Do not alter production data/);assert.match(disasterText,/Preserve immutable history/);assert.match(disasterText,/unperformed tests/);
console.log('Documentation accuracy and non-destructive disaster-planning guidance passed.');

const debate=context.buildWorkflowTemplate('cooperative-adversarial');
assert.equal(debate.nodes[1].title,'Builder proposes or revises');assert.equal(debate.nodes[2].title,'Challenger tests the proposal');assert.equal(debate.nodes[3].title,'Reconcile findings');
assert.match(debate.nodes[2].instructions,/Provide evidence and constructive alternatives/);assert.match(debate.nodes[3].instructions,/Do not force consensus/);assert.match(debate.nodes.find(n=>n.type==='approval').instructions,/human decides/i);
assert.ok(debate.nodes.filter(n=>['agent','review'].includes(n.type)).every(n=>n.agentId===''));
console.log('Builder/challenger sequencing, evidence-based critique and unresolved-disagreement guidance passed.');
