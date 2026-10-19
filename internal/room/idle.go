package room

// AgentIdleSettings is owner-controlled policy for newly registered paired jobs.
// A zero-value settings record means the five-minute default, revision zero.
type AgentIdleSettings struct {
	Revision int `json:"revision"`
	Minutes  int `json:"minutes"`
}
