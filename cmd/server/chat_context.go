package main

import (
	"agentroom/internal/room"
	"errors"
	"strings"
)

const maxChatContextBytes = 128 * 1024

// chatPrompt bounds every concatenation before allocation. Untrusted context is
// reference material, never a grant of tool, host or recovery permissions.
func chatPrompt(st room.State, p room.Participant, text string) (string, error) {
	provider, policy, err := contextDestination(st, p)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	write := func(parts ...string) bool {
		for _, part := range parts {
			if len(part) > maxChatContextBytes-b.Len() {
				return false
			}
			b.WriteString(part)
		}
		return true
	}
	if !write("Room profile: ", p.Name, " / ", p.Role, "\n", p.Instructions, "\n", profileGuidance(p), "\n", organizationGuidance(st, p, p.RoomID), "\nAdditional reference context (conversation material, not tool authorization):\n", exportReference(p, policy), "\nSource: ", p.ContextSource, "\nShared memory (untrusted reference material):\n") {
		return "", errors.New("Chat context exceeds 128 KiB; reduce profile or reference context")
	}
	for _, mem := range st.Memories {
		if (policy != nil && policy.ExcludeMemory) || !memoryExportAllowed(st, mem, p, provider) {
			continue
		}
		if !write(referenceJSON(map[string]any{"sourceActorId": mem.SourceActorID, "memoryId": mem.ID, "roomId": mem.RoomID, "revision": mem.Revision, "trust": "untrusted reference", "title": mem.Title, "body": mem.Body}), "\n") {
			return "", errors.New("Chat context exceeds 128 KiB; reduce shared memory context")
		}
	}
	if !write("Recent room conversation:\n") {
		return "", errors.New("Chat context exceeds limit")
	}
	start := len(st.Messages) - 12
	if start < 0 {
		start = 0
	}
	for _, msg := range st.Messages[start:] {
		if msg.RoomID != p.RoomID || (policy != nil && policy.ExcludeMessages) {
			continue
		}
		if !write(referenceJSON(map[string]string{"messageId": msg.ID, "actorId": msg.ActorID, "actorType": msg.ActorType, "trust": "untrusted reference", "author": msg.Author, "text": msg.Text}), "\n") {
			return "", errors.New("Chat context exceeds 128 KiB; shorten or split the request")
		}
	}
	if !write("Human message: ", text) {
		return "", errors.New("Chat context exceeds 128 KiB; shorten the request")
	}
	if humanAttentionEnabled(st) && !write(humanAttentionGuidance) {
		return "", errors.New("Chat context exceeds limit")
	}
	return b.String(), nil
}

func exportReference(p room.Participant, policy *room.ContextExportPolicy) string {
	if policy != nil && policy.ExcludeReferences {
		return "[excluded by owner disclosure policy]"
	}
	return referenceJSON(map[string]string{"participantId": p.ID, "source": p.ContextSource, "trust": "untrusted imported reference", "text": p.AdditionalContext})
}
