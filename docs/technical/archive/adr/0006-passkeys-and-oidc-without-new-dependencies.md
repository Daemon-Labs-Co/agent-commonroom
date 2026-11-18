# Add passkeys and OpenID Connect sign-in using only the standard library

Status: accepted

People using a hosted product will not manage RSA key files. Alongside the existing
owner key and member key files, workspaces gain passkey (WebAuthn) and OpenID
Connect sign-in, written against the Go standard library so the controller takes no
new security-critical dependency. Passkey support is limited to `none` attestation
and ES256 and RS256 credentials with required user verification; OIDC is the
authorization-code flow with PKCE and nonce, ID tokens signed with RS256 or ES256,
and issuer, audience, expiry and signing keys checked against configured values.

## Considered Options

- **A WebAuthn and an OIDC library.** Less code to write, but a large transitive
  dependency tree inside the component that authenticates owners. Rejected for now;
  the project does not add dependencies without the owner agreeing.
- **Only passkeys, or only OIDC.** Rejected by the owner; each covers customers the
  other cannot.

## Consequences

- We own parsing of attacker-controlled CBOR and JWT input, so these paths get
  fixed test vectors, fuzz tests and an adversarial review before release.
- Attestation formats other than `none` are refused, so a deployment cannot require a
  particular authenticator model.
- Each workspace stores its own credentials and OIDC links; the same person in two
  workspaces has two separate accounts.
