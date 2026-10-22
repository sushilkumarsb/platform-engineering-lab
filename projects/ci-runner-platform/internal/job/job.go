package job

import "time"

type State string

const (
	Queued    State = "QUEUED"
	Assigned  State = "ASSIGNED"
	Running   State = "RUNNING"
	Succeeded State = "SUCCEEDED"
	Failed    State = "FAILED"
	Cancelled State = "CANCELLED"
	TimedOut  State = "TIMED_OUT"
)

type Requirements struct {
	// Capability requirements.
	OS     string
	Arch   string
	Docker bool
	GPU    bool

	// Minimum resource requirements.
	MinCPU  int
	MinRAM  int // GB
	MinDisk int // GB
}

type Job struct {
	// Unique job execution ID.
	ID string

	// Source context.
	Repository string
	Ref        string

	// Current lifecycle state.
	State State

	// Requirements used by the scheduler.
	Requirements Requirements

	// Empty until assigned.
	RunnerID string

	// Execution attempt number.
	Attempt int

	CreatedAt   time.Time
	AssignedAt  time.Time
	StartedAt   time.Time
	CompletedAt time.Time
}
