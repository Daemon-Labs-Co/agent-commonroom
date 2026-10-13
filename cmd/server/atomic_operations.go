package main

import (
	"agentroom/internal/room"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

var errAuditConflict = errors.New("State changed during this operation; reload before retrying")

type expectedEventCountKey struct{}
type operationReceiptKey struct{}

// Operation receipts contain a request fingerprint and the already committed
// response. A replay acknowledges existing work and never dispatches it again.
func prepareOperation(w http.ResponseWriter, r *http.Request, st room.State, input any) (context.Context, bool) {
	ctx := context.WithValue(r.Context(), expectedEventCountKey{}, st.EventCount)
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		return ctx, false
	}
	if len(key) > 80 || len(key) < 8 || strings.ContainsAny(key, "\r\n\t ") {
		http.Error(w, "Invalid idempotency key", 400)
		return ctx, true
	}
	scope := requestActor(r) + "\n" + requestedRoom(r) + "\n" + r.URL.Path + "\n" + key
	digest := sha256.Sum256([]byte(scope))
	binding := hex.EncodeToString(digest[:])
	raw, _ := json.Marshal(input)
	fingerprint := sha256.Sum256(raw)
	receipt := room.OperationReceipt{Key: binding, Fingerprint: hex.EncodeToString(fingerprint[:])}
	if old, exists := st.OperationReceipts[binding]; exists {
		if old.Fingerprint != receipt.Fingerprint {
			http.Error(w, "Idempotency key already used for different input", 409)
		} else {
			w.Header().Set("Idempotent-Replay", "true")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(old.Status)
			_, _ = w.Write(old.Response)
		}
		return ctx, true
	}
	return context.WithValue(ctx, operationReceiptKey{}, receipt), false
}
func withOperationResponse(ctx context.Context, status int, response any) context.Context {
	receipt, ok := ctx.Value(operationReceiptKey{}).(room.OperationReceipt)
	if !ok {
		return ctx
	}
	receipt.Status = status
	receipt.Response, _ = json.Marshal(response)
	return context.WithValue(ctx, operationReceiptKey{}, receipt)
}
func operationError(w http.ResponseWriter, err error) {
	if errors.Is(err, errAuditConflict) {
		http.Error(w, err.Error(), 409)
	} else {
		http.Error(w, "Commit failed; no part of the operation was committed", 500)
	}
}
