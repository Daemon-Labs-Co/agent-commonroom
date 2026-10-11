package main

import (
	"agentroom/internal/room"
	"testing"
)

func TestGraphRejectsInvalidConnections(t *testing.T) {
	t.Parallel()
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

func TestWorkflowValidationBoundaries(t *testing.T) {
	t.Parallel()
	valid := func() room.Workflow {
		return room.Workflow{Name: "Review", MaxIterations: 3, Nodes: []room.WorkflowNode{{ID: "s", Type: "start", Title: "Start"}, {ID: "a", Type: "agent", Title: "Engineer"}, {ID: "f", Type: "finish", Title: "Done"}}, Edges: []room.WorkflowEdge{{ID: "1", From: "s", To: "a", Port: "next"}, {ID: "2", From: "a", To: "f", Port: "next"}}}
	}
	cases := []struct {
		name   string
		change func(*room.Workflow)
	}{
		{"empty name", func(w *room.Workflow) { w.Name = " " }},
		{"unbounded iterations", func(w *room.Workflow) { w.MaxIterations = 101 }},
		{"too many nodes", func(w *room.Workflow) { w.Nodes = make([]room.WorkflowNode, 101) }},
		{"too many edges", func(w *room.Workflow) { w.Edges = make([]room.WorkflowEdge, 301) }},
		{"duplicate node", func(w *room.Workflow) { w.Nodes[1].ID = "s" }},
		{"unknown type", func(w *room.Workflow) { w.Nodes[1].Type = "execute-shell" }},
		{"missing title", func(w *room.Workflow) { w.Nodes[1].Title = " " }},
		{"outside canvas", func(w *room.Workflow) { w.Nodes[1].X = 5001 }},
		{"invalid timeout", func(w *room.Workflow) { w.Nodes[1].Timeout = -1 }},
		{"invalid retries", func(w *room.Workflow) { w.Nodes[1].Retries = 11 }},
		{"self edge", func(w *room.Workflow) { w.Edges[0].To = "s" }},
		{"finish output", func(w *room.Workflow) { w.Edges[1].From = "f"; w.Edges[1].To = "a" }},
		{"start input", func(w *room.Workflow) { w.Edges[1].To = "s" }},
		{"duplicate edge id", func(w *room.Workflow) { w.Edges[1].ID = "1" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			w := valid()
			c.change(&w)
			if validateGraph(w) == nil {
				t.Fatal("invalid graph accepted")
			}
		})
	}
	for _, c := range []struct{ kind, allowed, rejected string }{{"review", "changes", "next"}, {"decision", "yes", "passed"}, {"parallel", "b", "next"}, {"finish", "", "next"}, {"agent", "next", "yes"}} {
		t.Run("port "+c.kind, func(t *testing.T) {
			t.Parallel()
			if c.kind != "finish" && !validPort(c.kind, c.allowed) {
				t.Fatal("valid output rejected")
			}
			if validPort(c.kind, c.rejected) {
				t.Fatal("invalid output accepted")
			}
		})
	}
}
