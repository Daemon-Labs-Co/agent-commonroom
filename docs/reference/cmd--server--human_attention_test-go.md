# cmd/server/human_attention_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/human_attention_test.go)

Automated regression checks for human_attention; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestHumanRequestDetectionAndProjection`** — Regression contract: TestHumanRequestDetectionAndProjection.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `humanRequestKind`, `mockServer`, `s.load`, `t.Fatal`, `t.Fatalf`, `t.Parallel`.
- **`TestHumanAttentionAcknowledgmentPreservesMessageAndIsIdempotent`** — Regression contract: TestHumanAttentionAcknowledgmentPreservesMessageAndIsIdempotent.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `db.snapshot`, `len`, `mockRequest`, `mockServer`, `s.load`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
- **`TestHumanAttentionAcknowledgmentRoomAndAccountBoundaries`** — Regression contract: TestHumanAttentionAcknowledgmentRoomAndAccountBoundaries.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `len`, `mockRequest`, `mockServer`, `s.load`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `visibleState`, `w.Body.String`.
- **`TestAgentAccountCannotAcknowledgeHumanRequest`** — Regression contract: TestAgentAccountCannotAcknowledgeHumanRequest.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `mockRequest`, `mockServer`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`.
- **`TestPublicMessageCannotForgeHumanAttention`** — Regression contract: TestPublicMessageCannotForgeHumanAttention.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `mockRequest`, `mockServer`, `s.load`, `t.Fatal`, `t.Parallel`.
