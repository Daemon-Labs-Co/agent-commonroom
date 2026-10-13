package main

import (
	"agentroom/internal/room"
	"strings"
	"testing"
)

func TestChatContextBoundsEverySourceAndKeepsReferenceLabels(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"profile", "memory", "message", "request"} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			st := room.State{}
			p := room.Participant{ID: "p", RoomID: "engineering"}
			text := "Synthetic request"
			large := strings.Repeat("x", maxChatContextBytes+1)
			switch source {
			case "profile":
				p.AdditionalContext = large
			case "memory":
				st.Memories = []room.Memory{{Body: large}}
			case "message":
				st.Messages = []room.Message{{Text: large}}
			case "request":
				text = large
			}
			if prompt, err := chatPrompt(st, p, text); err == nil || prompt != "" {
				t.Fatal("over-limit context accepted")
			}
		})
	}
	prompt, err := chatPrompt(room.State{Memories: []room.Memory{{Title: "Synthetic", Body: "Reference"}}}, room.Participant{RoomID: "engineering"}, "Question")
	if err != nil || !strings.Contains(prompt, "untrusted reference material") || len(prompt) > maxChatContextBytes {
		t.Fatal(prompt, err)
	}
}
