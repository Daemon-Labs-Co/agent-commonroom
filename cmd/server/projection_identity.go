package main

// Event reducers for people, invitations, credentials and the company. Each reducer applies one kind of recorded event to
// the projection and returns an error only for an event that must never have been recorded.

import (
	"agentroom/internal/room"
	"encoding/json"
	"strings"
)

// reduceInvitationCreated reduces invitation.created, invitation.revoked, invitation.redeemed, actor.credential.added, actor.credential.used, actor.credential.removed, actor.oidc.linked, actor.oidc.unlinked.
func reduceInvitationCreated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectIdentityEvent(st, kind, data); err != nil {
		return err
	}
	return nil
}

// reduceActorChatStyleUpdated reduces actor.chat_style.updated.
func reduceActorChatStyleUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectChatStyle(st, data); err != nil {
		return err
	}
	return nil
}

// reduceActorAgentHighlightUpdated reduces actor.agent_highlight.updated.
func reduceActorAgentHighlightUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectAgentHighlight(st, data); err != nil {
		return err
	}
	return nil
}

// reduceActorHelpUpdated reduces actor.help.updated.
func reduceActorHelpUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v struct {
		ID   string `json:"id"`
		Mode string `json:"mode"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	projectActors(st)
	for i := range st.Actors {
		if st.Actors[i].ID == v.ID {
			st.Actors[i].HelpDisplay = v.Mode
		}
	}
	return nil
}

// reduceActorPhotoSizeUpdated reduces actor.photo_size.updated.
func reduceActorPhotoSizeUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v struct {
		ID   string `json:"id"`
		Size string `json:"size"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	projectActors(st)
	for i := range st.Actors {
		if st.Actors[i].ID == v.ID {
			st.Actors[i].PhotoSize = v.Size
		}
	}
	return nil
}

// reduceActorImageUpdated reduces actor.image.updated.
func reduceActorImageUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v struct {
		ID    string `json:"id"`
		Image string `json:"image"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	projectActors(st)
	for i := range st.Actors {
		if st.Actors[i].ID == v.ID {
			st.Actors[i].Image = v.Image
		}
	}
	return nil
}

// reduceActorNameUpdated reduces actor.name.updated.
func reduceActorNameUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if e := json.Unmarshal(data, &v); e != nil {
		return e
	}
	for i, a := range st.Actors {
		if a.ID == v.ID {
			st.Actors[i].Name = v.Name
		}
	}
	return nil
}

// reduceActorRevised reduces actor.revised.
func reduceActorRevised(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v room.Actor
	if e := json.Unmarshal(data, &v); e != nil {
		return e
	}
	found := false
	for i, a := range st.Actors {
		if a.ID == v.ID {
			v.HelpDisplay = a.HelpDisplay
			v.PhotoSize = a.PhotoSize
			v.ChatStyle, v.HighlightedAgents = a.ChatStyle, a.HighlightedAgents
			v.Credentials, v.OIDCLinks = a.Credentials, a.OIDCLinks
			st.Actors[i] = v
			found = true
		}
	}
	if !found {
		st.Actors = append(st.Actors, v)
	}
	return nil
}

// reduceCompanyOrganizationUpdated reduces company.organization.updated.
func reduceCompanyOrganizationUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := json.Unmarshal(data, &st.Organization); err != nil {
		return err
	}
	return nil
}

// reduceCompanyUpdated reduces company.updated.
func reduceCompanyUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	json.Unmarshal(data, &st.Company)
	return nil
}

// reduceHumanProfileUpdated reduces human.profile.updated.
func reduceHumanProfileUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v struct {
		Name  string  `json:"name"`
		Title *string `json:"title"`
	}
	json.Unmarshal(data, &v)
	st.HumanRevision++
	previous := strings.TrimSpace(st.HumanName)
	st.HumanName = v.Name
	// One name everywhere: a rename releases every room-only name.
	if previous != strings.TrimSpace(v.Name) {
		for i := range st.Rooms {
			st.Rooms[i].HumanName = ""
		}
	}
	if v.Title != nil {
		st.HumanTitle = *v.Title
	}
	return nil
}
