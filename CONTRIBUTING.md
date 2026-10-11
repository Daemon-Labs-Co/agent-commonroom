# Contributing to Agent CommonRoom

Agent CommonRoom is an early prototype. Read README.md for setup and current limits, FUNCTIONAL_SPEC.md for intended behavior, and SECURITY_REVIEW.md for outstanding security and architecture issues.

## Development workflow

1. Use your own database and authentication keys. Never use live room history as a test fixture.
2. Keep changes focused and describe the problem, resulting behavior, and validation.
3. Run `make check`; use `COMMONROOM_TEST_DB` with a separate database for integration checks.
4. Explain UI changes with screenshots that contain no private messages, credentials, or personal data.
5. Update the README when capabilities or setup steps change. Do not describe saved configuration as an implemented execution guarantee.

Preserve append-only history, authenticate and authorize server-side, and keep provider credentials with their official authentication tools. Never commit `.data/`, keys, connection secrets, personal transcripts, database dumps, or local logs. Retain third-party license notices when changing bundled assets.

For vulnerabilities, follow SECURITY.md rather than publishing secrets or an exploit against a live installation in an issue. Contributor license/DCO policies have not been established; discuss substantial contributions with the maintainer before submitting them.

Project contributions are intended for distribution under GPL-3.0-only, as described in README.md and LICENSE. Third-party components retain their applicable notices and licenses.

## Secret checks before committing

Install Gitleaks from its official release or package distribution, then run `make secrets`. The repository includes a staged-content hook at `.githooks/pre-commit`; configure it with `git config core.hooksPath .githooks` if you do not already use a custom hook setup. If you have existing hooks, integrate this check into them instead. The hook blocks commits if the scanner is unavailable or finds a secret and redacts scanner output. It does not change commit signing or create commits.

A local scanner can also be placed at ignored `bin/gitleaks`; the hook checks your PATH first. Ignore rules do not protect tracked files or prevent force-adding files. Git hooks can be bypassed, and scanners can miss credentials: inspect staged content and enable repository-side secret scanning/push protection when publishing. Do not add blanket scan exclusions to hide real findings. Never commit local scanner reports or copied credentials.
