package cgroups

// This file defines the Manager and Resource interfaces.

// Manager creates a cgroup, sets resources limitions and applies it to processes.
//
// A Manager implementation is tied with a specific cgroup interface:
//   - cgroup v1 filesystem: V1fsManager
//   - cgconfig.service (deprecated, won't support)
//   - TODO: cgroup v2 filesystem (not implemented)
//   - TODO: systemd (not implemneted)
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

// Resource is a cgroup config item. For example `memory.limit_in_bytes`.
// A resource has a value. For example `1024000`.
//
// Implements the Resource interface to support a resouece. Any
// Resource can be struct, int, string, or any types. Just implements
// the Value() method to offer its value that will be write into the
// cgroup config file.
//
// Additional methods may be required for specific resource Managers.
// For example, to work with the V1fsManager, the V1fsResource interface
// is required, which requires an additional V1fsPath method.
type Resource interface {
	Value() string
}

// V1fsResource the Resource for cgroup v1 filesystem.
// Required by the V1fsManager
type V1fsResource interface {
	Resource
	// V1fsPath returns the config file path to the resource.
	// basePath is the mount point of cgroup v1 filesystem ("/sys/fs/cgroup/")
	V1fsPath(basePath string, containerId string) string
}
