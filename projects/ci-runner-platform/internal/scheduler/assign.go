package scheduler

import (
	"time"

	"github.com/sushilkumarsb/platform-engineering-lab/projects/ci-runner-platform/internal/job"
	"github.com/sushilkumarsb/platform-engineering-lab/projects/ci-runner-platform/internal/runner"
)

// Assign selects an eligible runner and records the assignment.
// The returned values are copies; the caller must retain them.
func Assign(j job.Job, runners []runner.Runner) (job.Job, []runner.Runner, bool) {
	selected, ok := Schedule(j, runners)
	if !ok {
		return j, runners, false
	}

	// Find the selected runner in the returned slice and update it.
	for i := range runners {
		if runners[i].ID != selected.ID {
			continue
		}

		j.State = job.Assigned
		j.RunnerID = runners[i].ID
		j.AssignedAt = time.Now()
		runners[i].CurrentJobs = append(runners[i].CurrentJobs, j.ID)

		// Keep runner state consistent with its remaining concurrency capacity.
		if len(runners[i].CurrentJobs) >= runners[i].Capacity.MaxConcurrentJobs {
			runners[i].State = runner.StateBusy
		} else {
			runners[i].State = runner.StateReady
		}

		return j, runners, true
	}

	// Defensive fallback: Schedule should only return a runner from this slice.
	return j, runners, false
}
