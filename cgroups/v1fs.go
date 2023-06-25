package cgroups

// This file defined the V1fsManager, a Manager implementation
// based on the cgroup v1 filesystem.

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

// V1fsManager is a cgroup Manager talks to the cgroup v1 filesystem.
//
// The Resources items should implement the V1fsResource interface to
// make them V1fsManager compatible.
type V1fsManager struct {
	// BasePath is the mount point of cgroup v1 filesystem ("/sys/fs/cgroup/"):
	//  assert $(stat -fc %T /sys/fs/cgroup/) == tmpfs
	BasePath string
	// CgName is the name of the cgroup:
	//  /sys/fs/cgroup/cpu/<CgName>/cpu.shares
	cgroupName string
}

// NewV1fsManager is a shortcut of:
//
//	m := &V1fsManager{BasePath}
//	m.Create(cgroupName)
func NewV1fsManager(basePath string, cgroupName string) (Manager, error) {
	m := &V1fsManager{
		BasePath: basePath,
	}

	err := m.Create(cgroupName)

	return m, err
}

// -- implement Manager interface --

func (v *V1fsManager) Create(name string) error {
	if _, err := v.checkBasePath(); err != nil {
		return err
	}
	if v.cgroupName != "" {
		return fmt.Errorf("cgroup name already set: %s", v.cgroupName)
	}
	if name == "" {
		return fmt.Errorf("cgroup name cannot be empty")
	}

	v.cgroupName = name
	for _, subsystem := range supportedSubsystem() {
		p := v.subsystemDir(subsystem)
		if err := mkdirIfNotExists(p); err != nil {
			return err
		}
	}

	// In my practice, the cpuset.cpus and cpuset.mems are required
	// to be set first, or any other resources will fail to set.
	if err := v.setResource(CpuSetCpus("0")); err != nil {
		return fmt.Errorf("error init cpuset.cpus: %w", err)
	}
	if err := v.setResource(CpuSetMems("0")); err != nil {
		return fmt.Errorf("error init cpuset.mems: %w", err)
	}

	return nil
}

func (v *V1fsManager) Apply(pid int) error {
	if _, err := v.checkBasePath(); err != nil {
		return err
	}
	if _, err := v.checkCreated(); err != nil {
		return err
	}

	for _, subsystem := range supportedSubsystem() {
		tasksFile := v.taskFile(subsystem)
		if err := appendFile(tasksFile, fmt.Sprint(pid)); err != nil {
			return err
		}
	}
	return nil
}

func (v *V1fsManager) Set(res Resources) error {
	if _, err := v.checkBasePath(); err != nil {
		return err
	}
	if _, err := v.checkCreated(); err != nil {
		return err
	}

	resources := reflect.ValueOf(res)
	for i := 0; i < resources.NumField(); i++ {
		field := resources.Field(i)
		if field.IsZero() {
			continue
		}
		err := v.setResource(field.Interface().(Resource))
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *V1fsManager) setResource(r Resource) error {
	if _, ok := r.(V1fsResource); !ok {
		return fmt.Errorf("res is not V1fsResource")
	}
	v1res := r.(V1fsResource)

	slog.Info("[cgroups] V1fsManager setResource", "value", v1res.Value(), "target", v1res.V1fsPath(v.BasePath, v.cgroupName))

	filePath := v1res.V1fsPath(v.BasePath, v.cgroupName)
	err := overwriteFile(filePath, v1res.Value())

	return err
}

// Destroy destroys the cgroup by calling cgdelete(1):
//
//	cgdelete --recursive <controllers>:<cgroupName>
func (v *V1fsManager) Destroy() {
	slog.Info("[cgroups] Destroy: do cgdelete.", "cgroupName", v.cgroupName)

	subsustems := strings.Join(supportedSubsystem(), ",")

	delete := exec.Command("cgdelete", "--recursive", subsustems+":"+v.cgroupName)
	if err := delete.Run(); err != nil {
		slog.Error("[cgroups] Destroy: cgdelete failed.", "err", err)
	}

	v.cgroupName = ""
}

// -- path methods --

// subsystemDir returns /sys/fs/cgroup/<subsystem>/<CgroupName>/
func (v *V1fsManager) subsystemDir(subsystem string) string {
	return path.Join(v.BasePath, subsystem, v.cgroupName)
}

// taskFile returns /sys/fs/cgroup/<subsystem>/<CgroupName>/tasks
func (v *V1fsManager) taskFile(subsystem string) string {
	return path.Join(v.subsystemDir(subsystem), "cgroup.procs")
}

// -- check methods --

const defaultCgroupV1FsBasePath = "/sys/fs/cgroup/"

// checkBasePath checks if BasePath is set, if not, set it to default.
//
// Return:
//   - true if BasePath is set
//   - false if BasePath is not set and set it to default
//   - error if BasePath is not a cgroup v1 filesystem.
func (v *V1fsManager) checkBasePath() (bool, error) {
	set := v.BasePath != ""

	if !set {
		slog.Warn("cgroup v1 base path not set, using default",
			"defaultCgroupV1FsBasePath", defaultCgroupV1FsBasePath)
		v.BasePath = defaultCgroupV1FsBasePath
	}

	if stat, err := os.Stat(v.BasePath); err != nil {
		return set, fmt.Errorf("error checking cgroup v1 base path: %w", err)
	} else if !stat.IsDir() {
		return set, fmt.Errorf("cgroup v1 base path is not a directory: %s", v.BasePath)
	}
	if err := assertFsType(v.BasePath, "tmpfs"); err != nil {
		return set, err
	}

	return set, nil
}

// checkCreated checks if the cgroup is created.
// If not, create it with a random name.
//
// Return:
//   - true, nil -> already created
//   - false, nil -> created here with random name successfully
//   - false, err -> error creating
func (v *V1fsManager) checkCreated() (bool, error) {
	if v.cgroupName != "" {
		return true, nil
	}

	randName := randCgroupName()

	slog.Warn("cgroup name not set, do Create() with random name", "randName", randName)

	err := v.Create(randName)
	return false, err
}

// -- helper functions --
// TODO: move to utils.go if can be reused

// v1fsPath: <basePath>/<subsystem>/<cgroupName>/<fileName>
func v1fsPath(basePath, subsystem, cgroupName, fileName string) string {
	return path.Join(basePath, subsystem, cgroupName, fileName)
}

// randCgroupName() returns a random cgroup name:
//
//	"hind_{uuid}"
func randCgroupName() string {
	return "hind_" + uuid.NewString()
}
