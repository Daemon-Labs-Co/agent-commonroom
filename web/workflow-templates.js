'use strict';
// Optional catalog. Each application creates fresh nodes and a new workflow.
const workflowTemplates=[
 {id:'idea-assessment',name:'Idea assessment',description:'Turn a human idea into a feasibility and effort assessment.',roles:'Principal engineer · Staff engineer',steps:[
  ['start','Human proposes an idea','Capture the problem, intended users and desired outcome.'],
  ['agent','Validate the idea','Act as a principal engineer. Challenge assumptions, identify value and constraints, and distinguish facts from guesses.'],
  ['agent','Explore functionality','Act as a staff engineer. Define the smallest useful version, options, dependencies, risks and estimated difficulty.'],
  ['evidence','Collect the assessment','Record the proposed scope, feasibility, effort range, assumptions and unresolved questions.'],
  ['review','Review the recommendation','Check whether the recommendation follows from the evidence. Request changes for unsupported claims.'],
  ['approval','Human chooses the direction','The human decides whether to proceed, revise or stop.'],
  ['finish','Assessment ready','Produce a documented recommendation; do not imply development has begun.']]},
 {id:'feature-development',name:'Feature development',description:'Build a feature with tests, documentation, a feature flag, and independent review.',roles:'Developer · QA engineer',steps:[
  ['start','Human requests a feature','Record the problem and acceptance criteria.'],
  ['agent','Clarify and design','Define scope, contracts, permissions, failure paths and an on/off feature flag.'],
  ['agent','Implement the feature','Build the agreed scope. Preserve history, enforce permissions and implement the documented feature flag.'],
  ['agent','Test and document','Add fast parallel-safe tests with mocked DB interactions. Document usage, configuration, flag states, API contracts and limitations.'],
  ['evidence','Collect delivery evidence','Record changes, checks, test results, documentation links, flag behavior and remaining limitations.'],
  ['review','Independent QA review','Verify acceptance criteria, tests, documentation and enabled/disabled flag behavior. Request changes when incomplete.'],
  ['approval','Human accepts the work','Review the evidence and decide acceptance. Prepare a commit message; the human signs commits.'],
  ['finish','Feature ready','Report delivery and known limitations.']]},
 {id:'ui-refinement',name:'UI refinement',description:'Refine an interface, verify interactions, and review the result visually.',roles:'UI specialist · Developer · QA engineer',steps:[
  ['start','Human gives UI feedback','Capture the current friction and desired behavior.'],
  ['agent','Propose the interaction','Act as a UI expert. Simplify hierarchy, terminology and conditional controls. Consider keyboard access and narrow screens.'],
  ['agent','Implement the refinement','Implement the agreed design behind a feature flag. Preserve input and existing permissions.'],
  ['agent','Verify and document','Test meaningful interactions and both flag states. Update usage and technical documentation. Review the UI in a browser using fictional data.'],
  ['evidence','Collect visual evidence','Attach screenshots, interaction checks, accessibility findings and documentation links.'],
  ['review','Review the experience','Check clarity, keyboard access, responsiveness, input preservation and regression risks.'],
  ['approval','Human reviews the UI','The human reviews the visual result and requests changes or accepts it.'],
  ['finish','Refinement ready','Summarize the change and verification.']]},
 {id:'bug-investigation',name:'Bug investigation',description:'Reproduce a problem, fix its cause, and prevent recurrence.',roles:'Developer · QA engineer',steps:[
  ['start','Human reports a problem','Record expected and observed behavior without disclosing credentials.'],
  ['agent','Investigate and reproduce','Find a minimal reproduction and the root cause. Use mocked DB interactions and authorized evidence.'],
  ['agent','Fix and prevent recurrence','Apply the smallest safe fix and a meaningful regression test. Document behavior and any feature-flag implications.'],
  ['evidence','Collect reproduction and results','Record the reproduction, root cause, regression result and relevant documentation.'],
  ['review','Verify the fix','Independently verify the reported problem is resolved and nearby behavior still works.'],
  ['approval','Human confirms resolution','The human reviews the evidence and decides whether the issue is resolved.'],
  ['finish','Issue resolved','State the fix and any remaining limits.']]},
 {id:'security-review',name:'Security review',description:'Assess threats, prioritize findings, and review remediation evidence.',roles:'Security engineer · Red-team specialist · Developer',steps:[
  ['start','Human authorizes review scope','Define authorized targets, accounts and constraints. No destructive testing or real credential disclosure.'],
  ['agent','Review threats and architecture','Assume constructive and malicious actors. Inspect trust boundaries, identity, room access, secrets, adapters and audit integrity.'],
  ['agent','Prioritize and remediate','Record severity, evidence, impact and actionable fixes. Implement authorized remediations with tests and documentation. Feature flags cannot bypass protections.'],
  ['evidence','Collect security evidence','Record redacted findings, reproductions, remediation checks and residual risks.'],
  ['review','Independent security verification','Verify remediations and remaining risks. Do not equate a passing test suite with security assurance.'],
  ['approval','Human accepts residual risk','The human decides follow-up priorities and acceptance.'],
  ['finish','Review ready','Publish the findings and remediation status within the authorized scope.']]},
 {id:'documenter',name:'Documenter',description:'Turn implementation details into complete, usable documentation.',roles:'Technical writer · Developer · Reviewer',steps:[
  ['start','Define the documentation scope','Capture the audience, feature or code area, and expected documentation deliverables.'],
  ['agent','Audit existing documentation','Inspect the source and existing guides. Identify outdated claims, missing usage instructions, undocumented interfaces and configuration gaps.'],
  ['agent','Write and update documentation','Document purpose, setup, usage, examples, permissions, feature flags, technical contracts and limitations. Add appropriate method/class comments and file responsibility references. Distinguish implemented behavior from plans.'],
  ['agent','Verify the instructions','Check links, examples and consistency with the implementation. Use mocked database interactions and fictional data. Never publish secrets or claim an untested installation works.'],
  ['evidence','Collect documentation evidence','Record changed guides, source references, link/example checks and anything still unverified.'],
  ['review','Review accuracy and usability','Independently check that the documentation is complete, understandable and supported by the code. Request changes for ambiguous steps or unsupported claims.'],
  ['approval','Human accepts documentation','The human confirms the guides answer the intended questions and accepts remaining limitations.'],
  ['finish','Documentation ready','Deliver documentation links and the verification record.']]},
 {id:'disaster-planning',name:'Disaster planning',description:'Plan for failures, recovery, and continuity before an incident occurs.',roles:'Reliability engineer · Security engineer · Operations reviewer',steps:[
  ['start','Define critical systems and scope','Identify the service, data, users, authorized planning scope and acceptable downtime/data loss. This workflow plans recovery; it does not authorize live destructive tests.'],
  ['agent','Assess failure scenarios','Review dependency outages, data corruption, lost credentials, compromised agents and audit-integrity failures. Identify impacts, owners and escalation triggers.'],
  ['agent','Design the recovery plan','Define recovery time and recovery point objectives, backup and restore procedures, communication responsibilities, rollback criteria and safe access recovery. Preserve immutable history and protect credentials.'],
  ['agent','Plan and assess rehearsal','Design a tabletop exercise and isolated restore verification using mocked database interactions or separately authorized disposable systems. Do not alter production data, purge records or run a live failover. Record gaps rather than claiming unperformed tests succeeded.'],
  ['evidence','Collect readiness evidence','Record the recovery runbook, dependency inventory, backup assumptions, rehearsal results or untested steps, recovery objectives and unresolved risks.'],
  ['review','Review recovery readiness','Independently challenge recovery assumptions, backup integrity, dependencies, ownership and evidence. Request changes when the plan cannot meet its stated objectives.'],
  ['approval','Human accepts the plan','The human approves responsibilities and residual risks. Schedule separately authorized rehearsals; approval of this document does not authorize destructive execution.'],
  ['finish','Recovery plan ready','Deliver the runbook, follow-up priorities and review/rehearsal schedule.']]},
 {id:'cooperative-adversarial',name:'Cooperative adversarial review',description:'Have a builder and a challenger improve a proposal together, with disagreements kept visible.',roles:'Builder · Challenger · Independent reviewer',steps:[
  ['start','Human defines the goal','Record the problem, acceptance criteria, authorized scope and constraints. Both agents share the goal of improving the outcome.'],
  ['agent','Builder proposes or revises','Act as the builder, preferably with constructive challenge level 1. Propose a solution supported by evidence, assumptions, tests and relevant documentation. On revision, address the previous critique explicitly.'],
  ['agent','Challenger tests the proposal','Act as the challenger, preferably with critical level 2 or adversarial level 3. Challenge assumptions, edge cases, security, failure modes and unsupported claims. Critique the work rather than the person. Provide evidence and constructive alternatives; do not manufacture objections merely to disagree.'],
  ['agent','Reconcile findings','Compare the proposal and critique. Record accepted criticisms, rejected criticisms with reasons, resulting changes and unresolved disagreements. Request further evidence where needed. Do not force consensus or claim authority from your title.'],
  ['evidence','Record the debate and outcome','Preserve the original proposal, critique, revision, supporting checks, documentation links and unresolved issues in the evidence record.'],
  ['review','Independent review','Use an independent reviewer to assess whether the outcome satisfies the goal and responds to substantiated criticism. Return incomplete work for revision. Keep unresolved tradeoffs visible for human decision.'],
  ['approval','Human resolves and accepts','The human decides unresolved tradeoffs and whether to accept the outcome. Agent consensus does not override a human decision or immutable history.'],
  ['finish','Reviewed recommendation ready','Deliver the agreed improvements, remaining disagreements and evidence.']]}


];
function buildWorkflowTemplate(id){
 const template=workflowTemplates.find(t=>t.id===id);if(!template)throw Error('Unknown workflow template');
 const nodes=template.steps.map(([type,title,instructions],i)=>({id:wfId(),type,title,instructions,x:60+(i%3)*290,y:60+Math.floor(i/3)*230,agentId:'',timeout:300,retries:0,evidence:['evidence','review','approval'].includes(type)?'Changes, checks and relevant documentation; record unresolved limitations.':''}));
 const edges=nodes.slice(0,-1).map((n,i)=>({id:wfId(),from:n.id,to:nodes[i+1].id,port:n.type==='review'?'passed':'next'}));
 const review=nodes.find(n=>n.type==='review');edges.push({id:wfId(),from:review.id,to:nodes[1].id,port:'changes'});
 return {name:template.name,revision:0,maxIterations:3,nodes,edges};
}
function templateScreen(){
 const actor=state.actors.find(a=>a.id===state.currentActorId),allowed=state.currentActorId==='local-owner'||actor?.role==='Developer';
 return '<div class="toolbar"><div><h1>Workflow templates</h1><p class="sub">A starting point for working with your squad.</p></div></div><p class="template-note">Apply to a room, assign its agents, then refine the steps. Templates create new saved workflows. Automatic execution is not enabled yet.</p><div class="template-grid">'+workflowTemplates.map(t=>'<article class="card template-card"><h2>'+esc(t.name)+'</h2><p>'+esc(t.description)+'</p><small>'+esc(t.roles)+'</small><p class="template-route">'+t.steps.map(s=>esc(s[1])).join(' → ')+'</p><button class="primary" data-apply-template="'+t.id+'" '+(!allowed?'disabled':'')+'>Apply to room</button></article>').join('')+'</div>'+(!allowed?'<p class="muted">An owner or Developer can apply templates to rooms they can access.</p>':'');
}
function applyWorkflowTemplate(id){
 const template=workflowTemplates.find(t=>t.id===id);if(!template)return;
 if(wfDirty&&!confirm('Applying a template opens its new workflow. Leave your unsaved draft?'))return;
 form('Apply workflow template','<p>'+esc(template.name)+' creates a new workflow. Existing workflows remain intact.</p><label class="field">Room<select name="room" required>'+state.rooms.map(r=>'<option value="'+esc(r.id)+'" '+(r.id===roomId?'selected':'')+'>'+esc(r.name)+'</option>').join('')+'</select></label><p class="muted">Assign room agents after applying. Run is not available yet.</p>',async data=>{
  const target=data.get('room'),draft=buildWorkflowTemplate(id);
  const result=await baseApi('events?room='+encodeURIComponent(target),{kind:'workflow.revised',templateId:id,payload:draft});
  roomId=target;localStorage.setItem('commonroom-room',roomId);state=await api('state');workflow=structuredClone(state.workflows.find(w=>w.id===result.workflowId));wfDirty=false;wfSelected=null;wfPort=null;wfUndo=[];wfRedo=[];view='Workflows';render();note('Template applied. Assign agents before validation.');
 });
 document.querySelector('#modal button.primary').textContent='Apply template';
}
const templateRender=render;render=function(){
 templateRender();const globals=document.querySelector('#nav .nav-global');if(globals){const b=document.createElement('button');b.textContent='Workflow templates';b.className=view==='Workflow templates'?'active':'';b.onclick=()=>{if(wfDirty&&!confirm('Leave the workflow editor with unsaved changes?'))return;view='Workflow templates';render()};globals.append(b)}
 if(view==='Workflow templates'){document.querySelector('#main').innerHTML=templateScreen();document.querySelectorAll('[data-apply-template]').forEach(b=>b.onclick=()=>applyWorkflowTemplate(b.dataset.applyTemplate));}
 if(view==='Workflows'){document.querySelector('#wf-template').textContent='Browse templates';document.querySelector('#wf-template').onclick=()=>{view='Workflow templates';render()};}
};
