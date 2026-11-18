package main

import (
	"agentroom/internal/room"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"slices"
	"strings"
	"time"
)

// Sign-in methods and invitations are stored as events in the workspace's own history.
// They are written only by the server's identity handlers: the generic events endpoint
// refuses these kinds.

const (
	invitationOpen     = "open"
	invitationRedeemed = "redeemed"
	invitationRevoked  = "revoked"

	maxCredentialsPerActor = 10
	maxOpenInvitations     = 50
)

// invitationID is what is stored and logged for an invitation secret.
func invitationID(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

type credentialAdded struct {
	ActorID    string          `json:"actorId"`
	Credential room.Credential `json:"credential"`
}
type credentialUsed struct {
	ActorID      string `json:"actorId"`
	CredentialID string `json:"credentialId"`
	SignCount    uint32 `json:"signCount"`
}
type credentialRemoved struct {
	ActorID      string `json:"actorId"`
	CredentialID string `json:"credentialId"`
}
type oidcLinked struct {
	ActorID string        `json:"actorId"`
	Link    room.OIDCLink `json:"link"`
}
type invitationRedeemedEvent struct {
	ID      string `json:"id"`
	ActorID string `json:"actorId"`
}
type invitationRevokedEvent struct {
	ID string `json:"id"`
}

func projectIdentityEvent(st *room.State, kind string, data []byte) error {
	projectActors(st)
	actor := func(id string) *room.Actor {
		for i := range st.Actors {
			if st.Actors[i].ID == id && st.Actors[i].Kind == "human" {
				return &st.Actors[i]
			}
		}
		return nil
	}
	switch kind {
	case "invitation.created":
		var v room.Invitation
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		v.Status = invitationOpen
		st.Invitations = append(st.Invitations, v)
	case "invitation.revoked", "invitation.redeemed":
		var v struct {
			ID      string `json:"id"`
			ActorID string `json:"actorId"`
		}
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		for i := range st.Invitations {
			if st.Invitations[i].ID == v.ID && st.Invitations[i].Status == invitationOpen {
				if kind == "invitation.revoked" {
					st.Invitations[i].Status = invitationRevoked
				} else {
					st.Invitations[i].Status, st.Invitations[i].RedeemedBy = invitationRedeemed, v.ActorID
				}
			}
		}
	case "actor.credential.added":
		var v credentialAdded
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		if a := actor(v.ActorID); a != nil && len(a.Credentials) < maxCredentialsPerActor {
			a.Credentials = append(a.Credentials, v.Credential)
		}
	case "actor.credential.used":
		var v credentialUsed
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		if a := actor(v.ActorID); a != nil {
			for i := range a.Credentials {
				if a.Credentials[i].ID == v.CredentialID {
					a.Credentials[i].SignCount = v.SignCount
				}
			}
		}
	case "actor.credential.removed":
		var v credentialRemoved
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		if a := actor(v.ActorID); a != nil {
			a.Credentials = slices.DeleteFunc(slices.Clone(a.Credentials), func(c room.Credential) bool { return c.ID == v.CredentialID })
			if len(a.Credentials) == 0 {
				a.Credentials = nil
			}
		}
	case "actor.oidc.linked":
		var v oidcLinked
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		if a := actor(v.ActorID); a != nil && !slices.ContainsFunc(a.OIDCLinks, func(l room.OIDCLink) bool { return l.Issuer == v.Link.Issuer && l.Subject == v.Link.Subject }) {
			a.OIDCLinks = append(a.OIDCLinks, v.Link)
		}
	case "actor.oidc.unlinked":
		var v oidcLinked
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		if a := actor(v.ActorID); a != nil {
			a.OIDCLinks = slices.DeleteFunc(slices.Clone(a.OIDCLinks), func(l room.OIDCLink) bool { return l.Issuer == v.Link.Issuer && l.Subject == v.Link.Subject })
			if len(a.OIDCLinks) == 0 {
				a.OIDCLinks = nil
			}
		}
	}
	return nil
}

// openInvitation finds an invitation that can still be redeemed.
func openInvitation(st room.State, token string, now time.Time) (room.Invitation, bool) {
	id := invitationID(token)
	for _, inv := range st.Invitations {
		if inv.ID == id && inv.Status == invitationOpen && now.Before(inv.Expires) {
			return inv, true
		}
	}
	return room.Invitation{}, false
}

func openInvitationCount(st room.State, now time.Time) int {
	n := 0
	for _, inv := range st.Invitations {
		if inv.Status == invitationOpen && now.Before(inv.Expires) {
			n++
		}
	}
	return n
}

// humanMembers counts people other than the owner who are, or are about to become, members.
func humanMembers(st room.State, now time.Time) int {
	n := openInvitationCount(st, now)
	for _, a := range st.Actors {
		if a.Kind == "human" && a.ID != ownerID {
			n++
		}
	}
	return n
}

func findCredential(st room.State, credentialID string) (room.Actor, room.Credential, bool) {
	for _, a := range st.Actors {
		if a.Kind != "human" {
			continue
		}
		for _, c := range a.Credentials {
			if c.ID == credentialID {
				return a, c, true
			}
		}
	}
	return room.Actor{}, room.Credential{}, false
}

func findOIDCActor(st room.State, issuer, subject string) (room.Actor, bool) {
	for _, a := range st.Actors {
		if a.Kind != "human" {
			continue
		}
		for _, l := range a.OIDCLinks {
			if l.Issuer == issuer && l.Subject == subject {
				return a, true
			}
		}
	}
	return room.Actor{}, false
}

// hasOtherSignInMethod reports whether an actor would still be able to sign in after
// one method is removed.
func hasOtherSignInMethod(a room.Actor, removing string) bool {
	if a.ID == ownerID || a.PublicKey != "" {
		return true // the owner key file or a member key file remains
	}
	n := len(a.OIDCLinks)
	for _, c := range a.Credentials {
		if c.ID != removing {
			n++
		}
	}
	return n > 0
}

func randomChallenge() []byte {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return b
}

func stdB64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

func decodeStdB64(s string) ([]byte, bool) {
	raw, err := base64.StdEncoding.DecodeString(s)
	return raw, err == nil && len(raw) > 0 && len(raw) < 4096
}

func b64url(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func decodeB64(s string, max int) ([]byte, bool) {
	if len(s) > max*2 {
		return nil, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
	if err != nil || len(raw) > max {
		return nil, false
	}
	return raw, true
}
