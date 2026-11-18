# Choose the workspace from the host name, one workspace per host

Status: accepted

In a multi-workspace installation each workspace is served at
`<slug>.<COMMONROOM_WORKSPACE_DOMAIN>`. The host decides the workspace for every
request, session cookies are host-only (never set for the parent domain) and the
Origin check compares against that exact host, so a session, CSRF token or setup link
for one workspace is meaningless in another.

## Considered Options

- **A path prefix (`/w/<slug>/…`).** Rejected: all workspaces would share one origin,
  so cookies, local storage and a script injected in one workspace could reach
  another's API.
- **A header or token naming the workspace.** Rejected: easy to forget on a route and
  invisible in browser isolation.

## Consequences

- Operators need wildcard DNS and a wildcard certificate for the domain.
- Slugs are lowercase DNS labels with a reserved list (`www`, `api`, `admin`,
  `platform`, `app`, …). Renaming a slug moves the address and signs everyone out.
- Custom domains per customer are a later addition that reuses the same lookup.
