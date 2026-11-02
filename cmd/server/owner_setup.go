package main

import (
	"agentroom/internal/room"
	"context"
	"errors"
	"net/http"
	"os"
)

// prepareOwnerIdentity never treats missing key files on an established database
// as permission to create a replacement owner. The first public key is committed
// with an empty-history compare-and-swap; private PEM bytes never enter events.
func (s *server) prepareOwnerIdentity(ctx context.Context, path string) error {
	st, err := s.load(ctx)
	if err != nil {
		return errors.New("Unable to verify first-owner database state")
	}
	if st.EventCount != 0 {
		if _, err := readPrivateFile(path + ".public.pem"); err != nil {
			return errors.New("Existing application history requires its original owner key; automatic owner creation refused")
		}
		if _, err := os.Stat(path + ".private.pem"); err == nil {
			for _, actor := range st.Actors {
				if actor.Kind == "human" && actor.ID != ownerID {
					return errors.New("First-owner setup refused: another human account exists")
				}
			}
		}
	}
	if err := s.auth.prepareRSA(path); err != nil {
		return err
	}
	public, err := readPrivateFile(path + ".public.pem")
	if err != nil {
		return errors.New("Owner public key unavailable")
	}
	if st.EventCount == 0 {
		ctx = context.WithValue(ctx, expectedEventCountKey{}, int64(0))
		return s.append(ctx, "actor.revised", room.Actor{ID: ownerID, Name: "Owner", Kind: "human", Role: "Owner", Enabled: true, Revision: 1, PublicKey: string(public)})
	}
	if actor, ok := member(st, ownerID); ok && actor.PublicKey != "" && actor.PublicKey != string(public) {
		return errors.New("Owner public key does not match the registered database identity")
	}
	return nil
}

func (s *server) setupStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", 405)
		return
	}
	s.auth.mu.Lock()
	pending := s.auth.setupPending
	s.auth.mu.Unlock()
	w.Header().Set("Cache-Control", "no-store")
	reply(w, map[string]bool{"setupRequired": pending})
}
