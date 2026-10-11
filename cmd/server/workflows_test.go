package main

import (
	"agentroom/internal/room"
	"testing"
)

func TestGraphRejectsInvalidConnections(t *testing.T) {
	w := room.Workflow{Name: "Review", MaxIterations: 3, Nodes: []room.WorkflowNode{{ID: "s", Type: "start", Title: "Start"}, {ID: "r", Type: "review", Title: "Review"}, {ID: "f", Type: "finish", Title: "Finish"}}, Edges: []room.WorkflowEdge{{ID: "1", From: "s", To: "r", Port: "next"}, {ID: "2", From: "r", To: "f", Port: "passed"}}}
	if err := validateGraph(w); err != nil {
		t.Fatal(err)
	}
	w.Edges = append(w.Edges, room.WorkflowEdge{ID: "3", From: "r", To: "f", Port: "passed"})
	if validateGraph(w) == nil {
		t.Fatal("Duplicate output accepted")
	}
	w.Edges = w.Edges[:2]
	w.Edges[1].To = "missing"
	if validateGraph(w) == nil {
		t.Fatal("Dangling edge accepted")
	}
	w.Edges[1].To = "f"
	w.Edges[1].Port = "next"
	if validateGraph(w) == nil {
		t.Fatal("Invalid review port accepted")
	}
}
