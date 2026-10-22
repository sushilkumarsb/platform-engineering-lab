
package scheduler

import (
	"github.com/sushilkumarsb/platform-engineering-lab/projects/ci-runner-platform/internal/job"
	"github.com/sushilkumarsb/platform-engineering-lab/projects/ci-runner-platform/internal/runner"
)

// Schedule selects the first runner that can execute the queued job.
// It returns false when no eligible runner is available.
func Schedule(j job.Job, runners []runner.Runner) (runner.Runner, bool) {
	if j.State != job.Queued {
		return runner.Runner{}, false
	}

	for _, r := range runners {
		if !matches(j, r) {
			continue
		}

		return r, true
	}

	return runner.Runner{}, false
}

func matches(j job.Job, r runner.Runner) bool {
	// Only ready or busy runners can accept additional work.
	if r.State != runner.StateReady && r.State != runner.StateBusy {
		return false
	}

	// A runner must have an available concurrency slot.
	if r.Capacity.MaxConcurrentJobs <= len(r.CurrentJobs) {
		return false
	}

	// Match required operating system and architecture.
	if j.Requirements.OS != "" &&
		j.Requirements.OS != r.Capabilities.OS {
		return false
	}

	if j.Requirements.Arch != "" &&
		j.Requirements.Arch != r.Capabilities.Architecture {
		return false
	}

	// Match required capabilities.
	if j.Requirements.Docker && !r.Capabilities.Docker {
		return false
	}

	if j.Requirements.GPU && !r.Capabilities.GPU {
		return false
	}

	// Match minimum resource requirements.
	if j.Requirements.MinCPU > r.Capacity.CPU {
		return false
	}

	// Job RAM is measured in GB; runner memory is measured in MB.
	if j.Requirements.MinRAM*1024 > r.Capacity.MemoryMB {
		return false
	}

	if j.Requirements.MinDisk > r.Capacity.DiskGB {
		return false
	}

	return true
}
