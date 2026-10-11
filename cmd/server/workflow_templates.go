package main

import (
	"os"
	"strconv"
)

func workflowTemplatesEnabled() (bool, error) {
	value := os.Getenv("COMMONROOM_WORKFLOW_TEMPLATES")
	if value == "" {
		return true, nil
	}
	return strconv.ParseBool(value)
}
func knownWorkflowTemplate(id string) bool {
	switch id {
	case "idea-assessment", "feature-development", "ui-refinement", "bug-investigation", "security-review", "documenter", "disaster-planning", "cooperative-adversarial":
		return true
	}
	return false
}
