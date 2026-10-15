package main

import (
	"agentroom/internal/room"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const defaultDailyTurnLimit = 200
const maximumDailyTurnLimit = 100000

type turnBudget struct {
	Limit     int       `json:"limit"`
	Used      int       `json:"used"`
	Remaining int       `json:"remaining"`
	ResetsAt  time.Time `json:"resetsAt"`
}

// parseDailyTurnLimit is an operator setting, not an optional security flag.
// Invalid/zero values fail startup rather than silently removing the budget.
func parseDailyTurnLimit(raw string) (int, error) {
	if raw == "" {
		return defaultDailyTurnLimit, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > maximumDailyTurnLimit || strings.TrimSpace(raw) != raw {
		return 0, errors.New("COMMONROOM_DAILY_TURN_LIMIT must be an integer from 1 to 100000")
	}
	return n, nil
}

// dailyTurnBudget counts unique controller-journalled intents, not lifecycle
// updates, successful completions or dollars. Restart cannot reset allocations.
func (s *server) dailyTurnBudget(st room.State, now time.Time) turnBudget {
	now = now.UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	limit := s.dailyTurnLimit
	if limit == 0 {
		limit = defaultDailyTurnLimit
	}
	budget := turnBudget{Limit: limit, ResetsAt: start.AddDate(0, 0, 1)}
	for _, job := range st.ManagedJobs {
		if !job.AllocatedAt.Before(start) && job.AllocatedAt.Before(budget.ResetsAt) {
			budget.Used++
		}
	}
	budget.Remaining = budget.Limit - budget.Used
	if budget.Remaining < 0 {
		budget.Remaining = 0
	}
	return budget
}
func (s *server) admitDailyTurns(st room.State, now time.Time, requested int) error {
	budget := s.dailyTurnBudget(st, now)
	if budget.Limit < 1 || budget.Limit > maximumDailyTurnLimit || requested < 1 || requested > budget.Remaining {
		return fmt.Errorf("Daily managed-turn budget exhausted (%d/%d); resets at %s. No model turn was started", budget.Used, budget.Limit, budget.ResetsAt.Format(time.RFC3339))
	}
	return nil
}
