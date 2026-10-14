package main

import (
	"agentroom/internal/room"
	"context"
	"log"
	"time"
)

// persistWorkflowTransitionLocked commits run and participant outcomes together.
// Callers hold s.mu. A storage failure latches the mandatory execution stop;
// neither optional flags nor a later successful write may clear that latch.
func (s *server) persistWorkflowTransitionLocked(run *room.WorkflowRun, effects ...auditChange) error {
	next := *run
	next.Revision++
	next.Updated = time.Now().UTC()
	changes := append([]auditChange{{Kind: "workflow.run.updated", Payload: next}}, effects...)
	if err := s.appendBatch(context.Background(), changes); err != nil {
		// Do not log provider output, prompts, credentials or database error details.
		log.Print("Workflow transition could not be persisted; execution stopped, owner recovery required")
		s.tripIntegrityLocked()
		return err
	}
	*run = next
	return nil
}
