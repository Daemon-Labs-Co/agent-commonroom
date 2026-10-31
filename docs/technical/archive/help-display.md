# Help window preference

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


[Documentation home](documentation-index.md)

Choose **Same window** or **Separate window** under **Settings → Help display**. You can also use the **Open help in** selector at the top of a help overlay. Same window is the default. The preference belongs to your human profile and is stored in the application's event history, so it follows your account across browsers and future sign-ins.

Same-window help opens above the current page or form. Its top-right X returns to the underlying form without discarding your entered values. Choosing Separate window in an open help overlay opens that guide separately and saves the preference; later help buttons use it automatically. The separate help window can stay open beside a connection form. Browser settings may open a tab instead. If pop-ups are blocked, CommonRoom explains the issue and displays the guide in its same-window overlay instead.

## Cosmetic feature flag

The owner can disable **Settings → Cosmetic → Help window choice**. Its stored flag is `helpWindowChoice`, enabled by default. Disabling hides choice controls, rejects preference-update requests on the server and forces same-window help. Existing saved preferences are retained for re-enabling. Documentation remains available and security boundaries are unchanged. Each human can change only their own preference; they cannot change another member's name, role or access.

## Technical contract

`Actor.helpDisplay` stores `same-window` or `separate-window`. Missing values default to same window. Event `actor.help.updated` accepts a `mode`; the server derives identity from the authenticated actor rather than trusting a payload ID. Human membership, CSRF, existing event authorization and the server feature flag are enforced. Member administration preserves this personal preference.

`web/help.js` applies the preference to all existing `openDocumentation` callers, including connection setup guides, menu help and contextual form help. URLs are always relative to the local public documentation reader, with encoded topic and section values. Pop-up windows have their opener detached. If a browser refuses window creation, the existing overlay is used. The overlay does not rebuild or close the underlying connection form. Help preference changes do not copy credentials, run agents or transmit room messages.

Parallel Go tests use mocked database events to verify self-only updates, strict mode validation, both feature-flag states and preference preservation through member edits. JavaScript tests use fake windows and document controls to verify profile selection, same-window defaults, encoded guide URLs, opener detachment and blocked-popup fallback. They do not write live database fixtures or launch provider agents.
