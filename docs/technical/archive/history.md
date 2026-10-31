# History and retention

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


[Documentation home](documentation-index.md) · [Product guide](concepts.md)

Find recorded decisions and diagnose outcomes while preserving the original evidence.

## History date tabs

History & retention separates activity into tabs across the top: **All activity**, **Today**, **Yesterday**, **Earlier this week**, **Last week**, and month/year tabs for older records. Empty date buckets are omitted. Calendar dates use your browser's local timezone; weeks start Monday. Future timestamps and invalid dates receive explicit Future dates/Unknown date tabs rather than misleading labels. Search filters the selected period, keeping date navigation available when no result matches. Left/Right, Home and End navigate tabs; the selected period is retained during rerenders and falls back to All activity if its bucket disappears.

History opens on Today, including when today has no records. Records load asynchronously in pages of at most 100; other date periods and All activity are fetched only when selected. Load more fetches the next page without downloading the whole history. Search applies to loaded records; load more to search additional records. Earlier months appear as tabs when discovered in Older activity or All activity. Loading, retry and empty states remain visible, and responses to a closed view are discarded.

This is a read-only projection of the event history. It does not rewrite timestamps, descriptions, checksums or immutable events, and it uses exclusive sequence cursors for historical pagination. The owner-only Cosmetic flag `historyDateGroups` defaults on, applies after saving/reloading without restart, and restores an ungrouped, asynchronously paged activity list when off. The server gates `/history-date-groups.js`; record access, retention and integrity enforcement remain unchanged.

Tests cover local date boundaries, daylight-saving changes, Monday/week transitions, month/year rollover, search, empty periods, tab isolation, and both delivery states with mocked database interactions.

## Settings sections and retention

The Settings page uses **Profile**, **Company**, **Features**, **Retention**, **Help**, and **About** tabs. Tabs appear only when there is content the current human is permitted to see. The Features section contains its own Security/Cosmetic/Connections/Workflows tabs. Your own allowed profile fields and help-window preference remain editable; invited humans do not gain company, membership, feature-management or retention permissions. Existing button actions are preserved when controls move between sections. The selected settings section is retained during rerenders; Arrow keys, Home and End provide keyboard navigation.

The sidebar now calls the activity page **History**. The existing retention-policy notice moves to **Settings → Retention**. The minimum remains 30 days; disposal and purge are still disabled pending storage enforcement and protected receipts. This layout change does not implement deletion or weaken immutable history. The owner-only `settingsTabs` cosmetic flag defaults on; saving reloads the browser without requiring a restart. Disabled presentation restores the original Settings cards and History & retention layout. The server gates `/settings-tabs.js`, while all underlying authorization and retention protections remain enforced in both states.

## Organized History

History opens **Today** and keeps asynchronous 100-record pages and **Load more**.
The refined view groups visible events by calendar day, labels their category and
outcome, and shows actor/room/subject context only when recorded or resolvable from
the authorized state. A workflow run's creator is labelled **Run owner** rather
than being presented as the agent executing each step. Missing historical actor or
scope information is not invented. Original summaries, timestamps, sequences and
checksums remain available under **Technical details**.

Search covers **loaded** descriptions, event types and displayed context, not the
entire database. Category and outcome filters distinguish Messages, Agents,
Workflows, Tasks & evidence, Memory, Connections, People & company, Rooms,
Security, Settings and Other. The default view hides routine session/activity and
running-workflow updates; **Show routine updates** exposes them. Failed/blocked/
interrupted updates are never classified as routine. Loaded-record statistics
explicitly show loaded, shown, error/attention and routine counts; they are not
installation-wide totals. Filters persist during this History visit, including
background renders, and reset on a new visit.

Agent failures include allowlisted explanations and a next step without exposing
raw provider errors. Event payloads, private message bodies, endpoints, credentials
and instructions are not returned by this presentation. History remains owner-only.

The owner-controlled Cosmetic flag `historyRefinement` (**Organized event history**)
defaults **on**. Saving/reloading applies it without restart. Off restores the earlier
layout, blocks `/history-presentation.js`, and omits enriched metadata from history
responses while keeping bounded pagination and audit protections. Date tabs remain
controlled separately by `historyDateGroups`. Neither flag rewrites, deletes or
changes verification of stored events.

**Refresh activity** fetches a fresh first page for the selected period, using the
current calendar boundaries and retaining filters. It resets loaded continuation
pages. Agent-status polling does not replace the refined History page while you
are reading, filtering or opening technical details; use this explicit refresh to
see new records. The manual synthetic fixture is
`COMMONROOM_BROWSER_HISTORY_TEST=1 go test ./cmd/server -run
'^TestWorkflowBrowserHarness$' -count=1 -timeout 2m -v` (60 seconds, in-memory only).

## API and pagination

History API: authenticated owner-only `GET /api/events?page=true&before=SEQUENCE&from=RFC3339&until=RFC3339` returns `{events,nextBefore}`. Date bounds are inclusive from/exclusive until and are calculated in the browser local timezone. The server validates dates/cursors, uses parameterized SQL and reads at most 101 records to provide a 100-record page plus continuation. The legacy request without `page=true` retains its latest-200 array response. Pagination is a mandatory bounded-read implementation safeguard, not an optional bypass of record permissions, retention or integrity. The existing owner-controlled `historyDateGroups` Cosmetic flag (on by default, no restart) controls Today/date-tab presentation; its script remains server-gated, and both enabled/disabled views use bounded async pages.
