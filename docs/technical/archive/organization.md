# Company organization

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


[Documentation home](documentation-index.md) · [Product guide](concepts.md)

Use the chart to make reporting relationships visible while keeping account
permissions and human decisions explicit.

## Company organization and decision authority

The chart supplies prompt guidance, not execution or override permission. The
current implementation compares reporting ancestry; automated consensus and
conflict arbitration are not implemented. Under the free specification,
unresolved disagreements require a human decision and block affected work.
Structured blocking/delegation remains unfinished; see [release readiness](release-readiness.md).

Legacy numeric rank is not authority. Titles, challenge levels, reporting lines
and payment cannot override authentication, isolation, acceptance permissions or
immutable history. Paid voting and escalation are separate future work.

## Edit the chart

1. As the owner, open **Settings → Company → Organization chart**.
2. Select an identity in the roster, canvas or accessible **List view**.
3. Choose **Reports to**. Drag nodes or use arrow keys; Shift moves farther.
4. Use zoom or **Auto-arrange** as needed, then choose **Save chart**.

Double-click an identity, or use the properties panel, to open its authorized
connection/profile editor. The chart stays open and retains unsaved positions;
those reporting changes still need an explicit save. Closing a dirty chart asks
whether to discard the draft.

A stale save returns a revision conflict. Reload before reapplying changes.
Only the owner may save; the server rejects cycles, missing/duplicate identities,
self-reporting and invalid positions. The chart supports up to 500 identities.

## Identities and privacy

One registered connection has one company identity even when several rooms use
it with different agent profiles. Humans and disabled identities are shown to
the owner. Removed identities disappear from the projection; a missing manager
becomes unassigned until the owner corrects the chart.

The API exposes chart metadata and room-profile identities, not conversations,
instructions, keys, provider credentials or connection endpoints. Other humans
do not gain company-wide chart access from their room membership.

## Configuration

The owner-controlled `companyOrgChart` flag defaults **on** under
**Settings → Features → Cosmetic**. Saving/reloading applies it without a restart.
Off hides the editor, blocks chart reads/writes, omits hierarchy guidance from new
turns and restores legacy rank presentation. Records remain preserved; active
prompts are not recalled and old ranks are not converted into reporting lines.

Ancestry comparisons and available-peer guidance are tested with synthetic
fixtures and mocked persistence. They are guidance, not proof that a provider
will follow it. Live saving/provider behavior remains a separate acceptance check.
