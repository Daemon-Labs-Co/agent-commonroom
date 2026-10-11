package main

import (
	"agentroom/internal/room"
	"fmt"
	"math"
	"strings"
)

func validateGraph(w room.Workflow) error {
	if strings.TrimSpace(w.Name) == "" {
		return fmt.Errorf("Workflow name required")
	}
	if len(w.Nodes) > 100 || len(w.Edges) > 300 {
		return fmt.Errorf("Workflow exceeds 100 nodes or 300 connections")
	}
	if w.MaxIterations < 1 || w.MaxIterations > 100 {
		return fmt.Errorf("Iteration limit must be 1–100")
	}
	nodes := map[string]room.WorkflowNode{}
	types := map[string]bool{"start": true, "agent": true, "review": true, "decision": true, "parallel": true, "evidence": true, "approval": true, "wait": true, "finish": true}
	for _, n := range w.Nodes {
		if n.ID == "" || nodes[n.ID].ID != "" || !types[n.Type] || strings.TrimSpace(n.Title) == "" {
			return fmt.Errorf("Node IDs, types and titles must be valid and unique")
		}
		if math.IsNaN(n.X) || math.IsNaN(n.Y) || n.X < 0 || n.Y < 0 || n.X > 5000 || n.Y > 5000 {
			return fmt.Errorf("Node position is outside the canvas")
		}
		if n.Timeout < 0 || n.Timeout > 86400 || n.Retries < 0 || n.Retries > 10 {
			return fmt.Errorf("Timeout or retries outside supported range")
		}
		nodes[n.ID] = n
	}
	edges := map[string]bool{}
	ports := map[string]bool{}
	for _, e := range w.Edges {
		from, to := nodes[e.From], nodes[e.To]
		if e.ID == "" || edges[e.ID] || from.ID == "" || to.ID == "" || e.From == e.To || from.Type == "finish" || to.Type == "start" {
			return fmt.Errorf("Invalid connection")
		}
		if !validPort(from.Type, e.Port) {
			return fmt.Errorf("Invalid output port")
		}
		key := e.From + "/" + e.Port
		if ports[key] {
			return fmt.Errorf("Output already connected; use Parallel to split work")
		}
		ports[key] = true
		edges[e.ID] = true
	}
	return nil
}
func validPort(kind, port string) bool {
	switch kind {
	case "finish":
		return false
	case "review":
		return port == "passed" || port == "changes"
	case "decision":
		return port == "yes" || port == "no"
	case "parallel":
		return port == "a" || port == "b"
	default:
		return port == "next"
	}
}
