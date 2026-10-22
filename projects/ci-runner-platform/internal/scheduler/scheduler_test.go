package scheduler

import (
	"testing"
	"time"

	"github.com/sushilkumarsb/platform-engineering-lab/projects/ci-runner-platform/internal/job"
	"github.com/sushilkumarsb/platform-engineering-lab/projects/ci-runner-platform/internal/runner"
)

func TestSelectsMatchingRunner(t *testing.T) {
	j := job.Job{
		ID:    "job-1",
		State: job.Queued,
		Requirements: job.Requirements{
			OS:   "linux",
			Arch: "amd64",
		},
	}

	runnerA := runner.Runner{
		ID:    "runner-a",
		State: runner.StateReady,
		Capabilities: runner.Capabilities{
			OS:           "linux",
			Architecture: "amd64",
		},
		Capacity: runner.Capacity{
			MaxConcurrentJobs: 1,
		},
		LastHeartbeat: time.Now(),
	}

	runnerB := runner.Runner{
		ID:    "runner-b",
		State: runner.StateReady,
		Capabilities: runner.Capabilities{
			OS:           "windows",
			Architecture: "amd64",
		},
		Capacity: runner.Capacity{
			MaxConcurrentJobs: 1,
		},
		LastHeartbeat: time.Now(),
	}

	runners := []runner.Runner{runnerA, runnerB}

	selected, ok := Schedule(j, runners)

	if !ok {
		t.Fatal("expected job to be scheduled")
	}

	if selected.ID != "runner-a" {
		t.Fatalf("expected runner-a, got %s", selected.ID)
	}
}

func TestRejectRunnerAtMaxConcurrency(t *testing.T) {
	j := job.Job{
		ID:    "job-1",
		State: job.Queued,
		Requirements: job.Requirements{
			OS:   "linux",
			Arch: "amd64",
		},
	}

	r := runner.Runner{
		ID:    "runner-a",
		State: runner.StateReady,
		Capabilities: runner.Capabilities{
			OS:           "linux",
			Architecture: "amd64",
		},
		Capacity: runner.Capacity{
			MaxConcurrentJobs: 1,
		},
		CurrentJobs: []string{"existing-job"},
	}

	selected, ok := Schedule(j, []runner.Runner{r})

	if ok {
		t.Fatalf("expected no runner to be selected, got %s", selected.ID)
	}
}

func TestBusyRunnerWithSpareCapacityCanBeSelected(t *testing.T) {
	j := job.Job{
		ID:    "job-1",
		State: job.Queued,
		Requirements: job.Requirements{
			OS:   "linux",
			Arch: "amd64",
		},
	}

	r := runner.Runner{
		ID:    "runner-a",
		State: runner.StateBusy,
		Capabilities: runner.Capabilities{
			OS:           "linux",
			Architecture: "amd64",
		},
		Capacity: runner.Capacity{
			MaxConcurrentJobs: 4,
		},
		CurrentJobs: []string{"job-running"},
	}

	selected, ok := Schedule(j, []runner.Runner{r})

	if !ok {
		t.Fatal("expected runner to be selected")
	}

	if selected.ID != "runner-a" {
		t.Fatalf("expected runner-a, got %s", selected.ID)
	}
}

func TestRejectRunnerWithInsufficientCPU(t *testing.T) {
	j := job.Job{
		ID:    "job-1",
		State: job.Queued,
		Requirements: job.Requirements{
			OS:     "linux",
			Arch:   "amd64",
			MinCPU: 4,
		},
	}

	r := runner.Runner{
		ID:    "runner-a",
		State: runner.StateReady,
		Capabilities: runner.Capabilities{
			OS:           "linux",
			Architecture: "amd64",
		},
		Capacity: runner.Capacity{
			CPU:               2,
			MaxConcurrentJobs: 1,
		},
	}

	selected, ok := Schedule(j, []runner.Runner{r})

	if ok {
		t.Fatalf("expected no runner to be selected, got %s", selected.ID)
	}
}

func TestRejectRunnerWithoutDockerSupport(t *testing.T) {
	j := job.Job{
		ID:    "job-1",
		State: job.Queued,
		Requirements: job.Requirements{
			OS:     "linux",
			Arch:   "amd64",
			Docker: true,
		},
	}

	r := runner.Runner{
		ID:    "runner-a",
		State: runner.StateReady,
		Capabilities: runner.Capabilities{
			OS:           "linux",
			Architecture: "amd64",
			Docker:       false,
		},
		Capacity: runner.Capacity{
			MaxConcurrentJobs: 1,
		},
	}

	selected, ok := Schedule(j, []runner.Runner{r})

	if ok {
		t.Fatalf("expected no runner to be selected, got %s", selected.ID)
	}
}

func TestRejectRunnerWithInsufficientRAM(t *testing.T) {
	j := job.Job{
		ID:    "job-1",
		State: job.Queued,
		Requirements: job.Requirements{
			OS:     "linux",
			Arch:   "amd64",
			MinRAM: 8,
		},
	}

	r := runner.Runner{
		ID:    "runner-a",
		State: runner.StateReady,
		Capabilities: runner.Capabilities{
			OS:           "linux",
			Architecture: "amd64",
		},
		Capacity: runner.Capacity{
			MemoryMB:          4096,
			MaxConcurrentJobs: 1,
		},
	}

	selected, ok := Schedule(j, []runner.Runner{r})

	if ok {
		t.Fatalf("expected no runner to be selected, got %s", selected.ID)
	}
}

func TestRejectRunnerWithoutGPU(t *testing.T) {
	j := job.Job{
		ID:    "job-1",
		State: job.Queued,
		Requirements: job.Requirements{
			OS:   "linux",
			Arch: "amd64",
			GPU:  true,
		},
	}

	r := runner.Runner{
		ID:    "runner-a",
		State: runner.StateReady,
		Capabilities: runner.Capabilities{
			OS:           "linux",
			Architecture: "amd64",
			GPU:          false,
		},
		Capacity: runner.Capacity{
			MaxConcurrentJobs: 1,
		},
	}

	selected, ok := Schedule(j, []runner.Runner{r})

	if ok {
		t.Fatalf("expected no runner to be selected, got %s", selected.ID)
	}
}

func TestAssignUpdatesJobAndRunner(t *testing.T) {
	j := job.Job{
		ID:    "job-assign-1",
		State: job.Queued,
		Requirements: job.Requirements{
			OS:   "linux",
			Arch: "amd64",
		},
	}

	r := runner.Runner{
		ID:    "runner-1",
		State: runner.StateReady,
		Capabilities: runner.Capabilities{
			OS:           "linux",
			Architecture: "amd64",
		},
		Capacity: runner.Capacity{
			MaxConcurrentJobs: 2,
		},
	}

	updatedJob, updatedRunners, ok := Assign(j, []runner.Runner{r})

	if !ok {
		t.Fatal("expected assignment to succeed")
	}

	if updatedJob.State != job.Assigned {
		t.Fatalf("expected job state %s, got %s", job.Assigned, updatedJob.State)
	}

	if updatedRunners[0].State != runner.StateReady {
		t.Fatalf("expected runner to remain READY, got %s", updatedRunners[0].State)
	}

	if updatedJob.RunnerID != "runner-1" {
		t.Fatalf("expected runner-1, got %s", updatedJob.RunnerID)
	}

	if updatedJob.AssignedAt.IsZero() {
		t.Fatal("expected AssignedAt to be set")
	}

	if len(updatedRunners) != 1 {
		t.Fatalf("expected one runner, got %d", len(updatedRunners))
	}

	if len(updatedRunners[0].CurrentJobs) != 1 ||
		updatedRunners[0].CurrentJobs[0] != "job-assign-1" {
		t.Fatalf("expected job-assign-1 in runner's CurrentJobs, got %v",
			updatedRunners[0].CurrentJobs)
	}
}

func TestAssignWithNoEligibleRunnerLeavesInputsUnchanged(t *testing.T) {
	j := job.Job{
		ID:    "job-no-match",
		State: job.Queued,
		Requirements: job.Requirements{
			OS: "linux",
		},
	}

	r := runner.Runner{
		ID:    "runner-windows",
		State: runner.StateReady,
		Capabilities: runner.Capabilities{
			OS:           "windows",
			Architecture: "amd64",
		},
		Capacity: runner.Capacity{
			MaxConcurrentJobs: 1,
		},
	}

	updatedJob, updatedRunners, ok := Assign(j, []runner.Runner{r})

	if ok {
		t.Fatal("expected assignment to fail")
	}

	if updatedJob.State != job.Queued {
		t.Fatalf("expected job to remain QUEUED, got %s", updatedJob.State)
	}

	if updatedJob.RunnerID != "" {
		t.Fatalf("expected no RunnerID, got %s", updatedJob.RunnerID)
	}

	if len(updatedRunners) != 1 || len(updatedRunners[0].CurrentJobs) != 0 {
		t.Fatal("expected runner state to remain unchanged")
	}
}

func TestAssignRejectsJobAlreadyAssigned(t *testing.T) {
	j := job.Job{
		ID:       "job-already-assigned",
		State:    job.Assigned,
		RunnerID: "runner-existing",
	}

	r := runner.Runner{
		ID:    "runner-1",
		State: runner.StateReady,
		Capabilities: runner.Capabilities{
			OS:           "linux",
			Architecture: "amd64",
		},
		Capacity: runner.Capacity{
			MaxConcurrentJobs: 2,
		},
	}

	updatedJob, updatedRunners, ok := Assign(j, []runner.Runner{r})

	if ok {
		t.Fatal("expected assignment to be rejected for an already-assigned job")
	}

	if updatedJob.State != job.Assigned {
		t.Fatalf("expected state to remain ASSIGNED, got %s", updatedJob.State)
	}

	if updatedJob.RunnerID != "runner-existing" {
		t.Fatalf("expected existing runner ID to remain unchanged, got %s", updatedJob.RunnerID)
	}

	if len(updatedRunners[0].CurrentJobs) != 0 {
		t.Fatalf("expected runner not to receive duplicate job, got %v", updatedRunners[0].CurrentJobs)
	}
}
