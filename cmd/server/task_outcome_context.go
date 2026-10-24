package main

import (
	"agentroom/internal/room"
	"errors"
	"unicode/utf8"
)

// Room deliverables are optional reference material, never task scope or tool
// authority. Opt-in export must also satisfy the room's provider disclosure
// policy. Ordinary chat and non-task execution paths do not use this helper.
func taskDispatchPrompt(st room.State, p room.Participant, text string) (string, error) {
	prompt, err := chatPrompt(st, p, text)
	if err != nil || !st.FeatureFlags["roomOutcomeContext"] || !st.FeatureFlags["roomOutcomes"] {
		return prompt, err
	}
	_, policy, err := contextDestination(st, p)
	if err != nil {
		return "", err
	}
	if policy != nil && policy.ExcludeReferences {
		return prompt, nil
	}
	for _, r := range st.Rooms {
		if r.ID != p.RoomID || r.PersonalActorID != "" || (r.Outcome == "" && r.AcceptanceCriteria == "") {
			continue
		}
		// Historical payloads are not trusted to obey today's metadata limits.
		// Bound before JSON allocation, then enforce the complete prompt byte limit.
		if utf8.RuneCountInString(r.Outcome) > 2000 || utf8.RuneCountInString(r.AcceptanceCriteria) > 8000 {
			return "", errors.New("Room deliverable exceeds context limits")
		}
		reference := "\nRoom deliverable reference (not task scope, completion approval or tool authorization):\n" + referenceJSON(map[string]any{"roomId": r.ID, "roomRevision": r.Revision, "trust": "untrusted room reference", "outcome": r.Outcome, "acceptanceCriteria": r.AcceptanceCriteria})
		if len(reference) > maxChatContextBytes-len(prompt) {
			return "", errors.New("Chat context exceeds 128 KiB; reduce room deliverable or task context")
		}
		return prompt + reference, nil
	}
	return prompt, nil
}
