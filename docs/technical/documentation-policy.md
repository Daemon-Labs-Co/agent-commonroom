# Documentation audience contract

## Input

A functionality change, user procedure, module contract or documentation update.

## Output

User guides under docs/user; active technical contracts under docs/technical.
All local links/assets must resolve and in-app Help must retain its allowlist.

## Functionality

User guides are grouped by functionality. Every page includes a relevant local
image with descriptive alt text and a small Bash setup/navigation helper. Task
instructions show ordered actions and expected results. Images must be real
prototype screenshots or explicitly labeled explanatory illustrations. Helpers
declare effects and prerequisites, reject unknown arguments and avoid credential
output, hidden configuration changes and destructive automatic actions.

Active technical pages have exactly four level-two sections: Input, Output,
Functionality, Properties and state changes. Input includes required permissions,
configuration and preconditions; Output includes success/refusal/errors;
Functionality includes validation and limits; Properties includes defaults,
mutable/persistent state, side effects, cancellation and disabled behavior.
Source links supplement these contracts. User tutorials and chronological progress
belong outside active technical reference.

Legacy mixed guides, specifications and recorded evidence remain under archive
with their historical anchors intact. Generated declaration maps remain under
reference. These retained records are excluded from the active four-section
format, and are not presented as current user tutorials or proof of readiness.
Root legal/contributor/specification documents retain their distinct purposes.

## Properties and state changes

Documentation changes do not modify product authority, flags or tier placement.
Audience and link/image checks run as part of make check. New Help documents and
images require explicit allowlist updates; arbitrary paths and symlinks stay refused.
