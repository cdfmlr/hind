package cgroups

import (
	"fmt"
	"golang.org/x/exp/slog"
	"os"
	"path"
	"reflect"
)

// V1fsResource interface to cgroup v1 filesystem
type V1fsResource interface {
	Resource
	// V1fsPath returns the config file path to the resource.
	// basePath is the mount point of cgroup v1 filesystem ("/sys/fs/cgroup/")
	V1fsPath(basePath string, containerId string) string
}

// V1fsManager is a cgroup Manager talks to the cgroup v1 filesystem.
type V1fsManager struct {
	// BasePath is the mount point of cgroup v1 filesystem ("/sys/fs/cgroup/"):
	//  assert $(stat -fc %T /sys/fs/cgroup/) == tmpfs
	BasePath string
	// CgName is the name of the cgroup:
	//  /sys/fs/cgroup/cpu/<CgName>/cpu.shares
	CgroupName string
}

func (v *V1fsManager) Create(name string) error {
	v.CgroupName = name
	for _, subsystem := range supportedSubsystem() {
		p := v.subsystemDir(subsystem)
		if err := mkdirIfNotExists(p); err != nil {
			return err
		}
	}
	return nil
}

// subsystemDir returns /sys/fs/cgroup/<subsystem>/<CgroupName>/
func (v *V1fsManager) subsystemDir(subsystem string) string {
	return path.Join(v.BasePath, subsystem, v.CgroupName)
}

// taskFile returns /sys/fs/cgroup/<subsystem>/<CgroupName>/tasks
func (v *V1fsManager) taskFile(subsystem string) string {
	return path.Join(v.subsystemDir(subsystem), "tasks")
}

func mkdirIfNotExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
	}
	return nil
}

func (v *V1fsManager) Apply(pid int) error {
	for _, subsystem := range supportedSubsystem() {
		tasksFile := v.taskFile(subsystem)
		if err := appendFile(tasksFile, fmt.Sprint(pid)); err != nil {
			return err
		}
	}
	return nil
}

func (v *V1fsManager) Set(res Resources) error {
	resources := reflect.ValueOf(res).Elem()
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

	filePath := v1res.V1fsPath(v.BasePath, v.CgroupName)
	err := overwriteFile(filePath, v1res.Value())
	if err != nil {
		return err
	}

	return nil
}

func (v *V1fsManager) Destroy() {
	slog.Warn("TODO: Destroy a cgroup")
}

// v1fsPath: <basePath>/<subsystem>/<cgroupName>/<fileName>
func v1fsPath(basePath, subsystem, cgroupName, fileName string) string {
	return path.Join(basePath, subsystem, cgroupName, fileName)
}
