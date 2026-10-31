# Provider-independent coordination with deterministic security enforcement

Retained detailed reference. Start with [user guides](../../../user/README.md) for illustrated
steps or [technical contracts](../../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


CommonRoom treats the selected governing agent as a coordinator and potentially
malicious participant, not the security authority. The first eligible connected
agent governs by default until an authorized human explicitly appoints another
eligible loaded agent; controller-owned state, scoped grants and authenticated
approvals enforce authority independently of every model. This preserves flexible
provider selection without trusting a user-customized model to authenticate
humans, enlarge permissions or suppress the authoritative record.

Codex-specific engine isolation remains useful adapter work. It is neither a
mandatory governance provider nor evidence that governance transfer is implemented.
See [the governance design](../governance-model.md) for implementation boundaries
and acceptance requirements. This decision is accepted; its implementation is pending.
