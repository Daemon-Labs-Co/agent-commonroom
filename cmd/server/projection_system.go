package main

// Event reducers for feature flags, stop and incident state and governance. Each reducer applies one kind of recorded event to
// the projection and returns an error only for an event that must never have been recorded.

import (
	"encoding/json"
	"errors"
)

// reduceGovernanceAppointed reduces governance.appointed.
func reduceGovernanceAppointed(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectGovernance(st, data); err != nil {
		return err
	}
	return nil
}

// reduceAgentsIdlePolicyUpdated reduces agents.idle_policy.updated.
func reduceAgentsIdlePolicyUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectIdleSettings(st, data); err != nil {
		return err
	}
	return nil
}

// reduceFeatureFlagUpdated reduces feature.flag.updated.
func reduceFeatureFlagUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v struct {
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
	}
	if e := json.Unmarshal(data, &v); e != nil {
		return e
	}
	if st.FeatureFlags == nil {
		st.FeatureFlags = map[string]bool{}
	}
	if v.Name == "agentCoordination" && !v.Enabled {
		if st.CoordinationEpoch == int(^uint(0)>>1) {
			return errors.New("coordination generation exhausted")
		}
		st.CoordinationEpoch++
	}
	if v.Name == "taskAcceptanceReview" && !v.Enabled {
		if st.TaskAcceptanceEpoch == int(^uint(0)>>1) {
			return errors.New("task acceptance epoch exhausted")
		}
		st.TaskAcceptanceEpoch++
	}
	if v.Name == "taskValueDependencies" && !v.Enabled {
		if st.ValueDependencyEpoch == int(^uint(0)>>1) {
			return errors.New("value dependency epoch exhausted")
		}
		st.ValueDependencyEpoch++
	}
	if v.Name == "taskDefinitions" && !v.Enabled {
		if st.TaskDefinitionEpoch == int(^uint(0)>>1) {
			return errors.New("task definition epoch exhausted")
		}
		st.TaskDefinitionEpoch++
	}
	if v.Name == "taskDelegation" && !v.Enabled {
		if st.DelegationEpoch == int(^uint(0)>>1) {
			return errors.New("delegation generation exhausted")
		}
		st.DelegationEpoch++
	}
	st.FeatureFlags[v.Name] = v.Enabled
	return nil
}

// reduceAgentsStop reduces agents.stop, agents.resume.
func reduceAgentsStop(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v struct {
		Scope  string `json:"scope"`
		RoomID string `json:"roomId"`
	}
	if e := json.Unmarshal(data, &v); e != nil {
		return e
	}
	if v.Scope == "application" {
		st.AgentsStopped = kind == "agents.stop"
	} else if v.Scope == "room" {
		kept := []string{}
		for _, rid := range st.AgentStopRooms {
			if rid != v.RoomID {
				kept = append(kept, rid)
			}
		}
		if kind == "agents.stop" {
			kept = append(kept, v.RoomID)
		}
		st.AgentStopRooms = kept
	}
	return nil
}

// reduceIncidentOpened reduces incident.opened.
func reduceIncidentOpened(proj *projection, kind string, data []byte) error {
	st := &proj.st
	st.Incident = true
	return nil
}

// reduceIncidentRecovered reduces incident.recovered.
func reduceIncidentRecovered(proj *projection, kind string, data []byte) error {
	st := &proj.st
	st.Incident = false
	return nil
}
