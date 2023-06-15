package cgroups

// Manager manages a cgroup
type Manager interface {
	// Create a cgroup with name.
	Create(name string) error
	// Apply adds a process with the specified pid into the cgroup.
	// It creates a cgroup with pid as the name, if not yet created.
	Apply(pid int) error

	// Set sets cgroup resources parameters/limits.
	Set(res Resources) error

	// Destroy removes cgroup.
	Destroy()
}
