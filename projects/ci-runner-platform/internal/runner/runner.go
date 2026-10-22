package runner

import "time"

type State string

const (
	StateProvisioning State = "provisioning"
	StateReady        State = "ready"
	StateBusy         State = "busy"
	StateDraining     State = "draining"
	StateOffline      State = "offline"
)

type Capabilities struct {
	OS           string
	Architecture string
	Docker       bool
	GPU          bool
}

type Capacity struct {
	CPU               int
	MemoryMB          int
	DiskGB            int
	MaxConcurrentJobs int
}

type Runner struct {
	ID            string
	State         State
	Capabilities  Capabilities
	Capacity      Capacity
	LastHeartbeat time.Time
	CurrentJobs   []string
}
