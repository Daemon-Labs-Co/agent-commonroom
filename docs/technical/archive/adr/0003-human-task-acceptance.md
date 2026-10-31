# Bind human task acceptance to exact reviewed work

Retained detailed reference. Start with [user guides](../../../user/README.md) for illustrated
steps or [technical contracts](../../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


Human acceptance is an immutable criterion-by-criterion review tied to actual
work outputs and current input provenance, committed atomically with task state.
Its content snapshot excludes the status/revision update performed by acceptance
itself; separate exact pre/post revision fences prevent reuse after later edits.
Schema assessments remain controller-owned structural facts, while human notes
and model findings never become independently observed verification evidence.
