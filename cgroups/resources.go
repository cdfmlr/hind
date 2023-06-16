package cgroups

// This file defines the supported Resources.

import (
	"fmt"
)

// Resources assembles the supported cgroup configs.
//
// All the fields are expected to be the implementations
// of the Resource interface.
type Resources struct {
	CpuQuotaUs       CpuQuotaUs
	CpuPeriodUs      CpuPeriodUs
	CpuSetCpus       CpuSetCpus
	MemoryLimitBytes MemoryLimitBytes
}

// ⬇️ stupid things for v1fs

const (
	SubsystemCpu    = "cpu"
	SubsystemCpuSet = "cpuset"
	SubsystemMemory = "memory"
)

func supportedSubsystem() []string {
	return []string{SubsystemCpu, SubsystemCpuSet, SubsystemMemory}
}

// ⬇️ Resources items

// CpuQuotaUs is the CPU hardcap limit (in usecs). Allowed cpu time in a given period.
type CpuQuotaUs int64

func (q CpuQuotaUs) Value() string {
	return fmt.Sprint(int64(q))
}

func (q CpuQuotaUs) V1fsPath(basePath string, containerId string) string {
	return v1fsPath(basePath, SubsystemCpu, containerId, "cpu.cfs_quota_us")
}

// CpuPeriodUs is CPU period to be used for hardcapping (in usecs). 0 to use system default.
type CpuPeriodUs uint64

func (p CpuPeriodUs) Value() string {
	return fmt.Sprint(uint64(p))
}

func (p CpuPeriodUs) V1fsPath(basePath string, containerId string) string {
	return v1fsPath(basePath, SubsystemCpu, containerId, "cpu.cfs_period_us")
}

// CpuSetCpus is the requested CPUs to be used by tasks within this cgroup: 0-4,6,8-10
type CpuSetCpus string

func (c CpuSetCpus) Value() string {
	return string(c)
}

func (c CpuSetCpus) V1fsPath(basePath string, containerId string) string {
	return v1fsPath(basePath, SubsystemCpuSet, containerId, "cpuset.cpus")
}

// MemoryLimitBytes sets memory.limit_in_bytes
type MemoryLimitBytes uint64

func (m MemoryLimitBytes) Value() string {
	return fmt.Sprint(uint64(m))
}

func (m MemoryLimitBytes) V1fsPath(basePath string, containerId string) string {
	return v1fsPath(basePath, SubsystemMemory, containerId, "memory.limit_in_bytes")
}
