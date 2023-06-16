package cgroups

import (
	"os"
	"os/exec"
	"testing"
)

func TestV1fsManagerInterface(t *testing.T) {
	var _ Manager = &V1fsManager{} // this is already enough actually

	m := &V1fsManager{}
	if _, ok := interface{}(m).(Manager); !ok {
		t.Errorf("V1fsManager must implement Manager interface")
	}
}

// TODO: implement me
func TestV1fsManager_Create(t *testing.T) {
	type fields struct {
		BasePath   string
		CgroupName string
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &V1fsManager{
				BasePath:   tt.fields.BasePath,
				cgroupName: tt.fields.CgroupName,
			}
			if err := v.Create(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("V1fsManager.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestV1fsManager_subsystemDir(t *testing.T) {
	t.Log("This test needs to be run as root (to v1fsManager.Create) for the first time:\n sudo go test -timeout 30s -run ^TestV1fsManager_subsystemDir$ hind/cgroups")

	basePath := "/sys/fs/cgroup"
	testCgroupName := "hind/TestV1fsManager_subsystemDir"

	type fields struct {
		BasePath   string
		CgroupName string
	}
	type args struct {
		subsystem string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"cpu", fields{basePath, testCgroupName}, args{"cpu"}, "/sys/fs/cgroup/cpu/" + testCgroupName},
		{"cpuset", fields{basePath, testCgroupName}, args{"cpuset"}, "/sys/fs/cgroup/cpuset/" + testCgroupName},
		{"memory", fields{basePath, testCgroupName}, args{"memory"}, "/sys/fs/cgroup/memory/" + testCgroupName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &V1fsManager{
				BasePath: tt.fields.BasePath,
			}
			if err := v.Create(tt.fields.CgroupName); err != nil {
				t.Fatalf("V1fsManager.Create() error = %v", err)
			}

			got := v.subsystemDir(tt.args.subsystem)
			if got != tt.want {
				t.Errorf("V1fsManager.subsystemDir() = %v, want %v", got, tt.want)
			}

			// check if the directory exists
			if _, err := os.Stat(got); os.IsNotExist(err) {
				t.Errorf("V1fsManager.subsystemDir() = %v: path does not exist", got)
			} else if err != nil {
				t.Errorf("V1fsManager.subsystemDir() = %v: stat error: %v", got, err)
			}
		})
	}
}

func TestV1fsManager_taskFile(t *testing.T) {
	basePath := "/sys/fs/cgroup"
	cgroupName := "hind/TestV1fsManager_taskFile"

	type fields struct {
		BasePath   string
		CgroupName string
	}
	type args struct {
		subsystem string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"cpu", fields{basePath, cgroupName}, args{"cpu"}, "/sys/fs/cgroup/cpu/hind/TestV1fsManager_taskFile/tasks"},
		{"cpuset", fields{basePath, cgroupName}, args{"cpuset"}, "/sys/fs/cgroup/cpuset/hind/TestV1fsManager_taskFile/tasks"},
		{"memory", fields{basePath, cgroupName}, args{"memory"}, "/sys/fs/cgroup/memory/hind/TestV1fsManager_taskFile/tasks"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &V1fsManager{
				BasePath: tt.fields.BasePath,
			}
			if err := v.Create(tt.fields.CgroupName); err != nil {
				t.Fatalf("V1fsManager.Create() error = %v", err)
			}

			got := v.taskFile(tt.args.subsystem)
			if got != tt.want {
				t.Errorf("V1fsManager.taskFile() = %v, want %v", got, tt.want)
			}

			// check if the file exists
			if _, err := os.Stat(got); os.IsNotExist(err) {
				t.Errorf("V1fsManager.taskFile() = %v: path does not exist", got)
			} else if err != nil {
				t.Errorf("V1fsManager.taskFile() = %v: stat error: %v", got, err)
			}
		})
	}
}

// FIXME: 在一次运行（条件不明，好像是第一次运行？）中报错 No space left on device
// 重置 cpuset.mems 后才能写入 tasks：
//
//	sudo sh -c "echo 0 > /sys/fs/cgroup/cpuset/hind/TestV1fsManager_Apply/cpuset.cpus"
//	sudo sh -c "echo 0 > /sys/fs/cgroup/cpuset/hind/TestV1fsManager_Apply/cpuset.mems"
//
// 参考: https://stackoverflow.com/questions/28348627/echo-tasks-gives-no-space-left-on-device-when-trying-to-use-cpuset
func TestV1fsManager_Apply(t *testing.T) {
	basePath := "/sys/fs/cgroup"
	cgroupName := "hind/TestV1fsManager_Apply"

	pid := os.Getpid()

	type fields struct {
		BasePath   string
		CgroupName string
	}
	type args struct {
		pid int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"cpu", fields{basePath, cgroupName}, args{pid}, false},
		{"cpuset", fields{basePath, cgroupName}, args{pid}, false},
		{"memory", fields{basePath, cgroupName}, args{pid}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &V1fsManager{
				BasePath: tt.fields.BasePath,
			}
			if err := v.Create(tt.fields.CgroupName); err != nil {
				t.Fatalf("V1fsManager.Create() error = %v", err)
			}
			if err := v.Apply(tt.args.pid); (err != nil) != tt.wantErr {
				t.Errorf("V1fsManager.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: implement me
func TestV1fsManager_Set(t *testing.T) {
	basePath := "/sys/fs/cgroup"
	cgroupName := "hind/TestV1fsManager_Set"

	resources := Resources{
		CpuQuotaUs:       100000,
		CpuPeriodUs:      100000,
		CpuSetCpus:       "0",
		MemoryLimitBytes: 1000000,
	}

	type fields struct {
		BasePath   string
		CgroupName string
	}
	type args struct {
		res Resources
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"cpu", fields{basePath, cgroupName}, args{resources}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &V1fsManager{
				BasePath:   tt.fields.BasePath,
			}
			if err := v.Create(tt.fields.CgroupName); err != nil {
				t.Fatalf("V1fsManager.Create() error = %v", err)
			}
			if err := v.Set(tt.args.res); (err != nil) != tt.wantErr {
				t.Errorf("V1fsManager.Set() error = %v, wantErr %v", err, tt.wantErr)
			}

			// read back the values
			cpuQuotaUs, err := exec.Command("cat", "/sys/fs/cgroup/cpu/hind/TestV1fsManager_Set/cpu.cfs_quota_us").Output()
			if err != nil {
				t.Errorf("V1fsManager.Set() read cpu.cfs_quota_us error: %v", err)
			} else if string(cpuQuotaUs) != "100000\n" {
				t.Errorf("V1fsManager.Set() read cpu.cfs_quota_us error: got %v, want 100000", string(cpuQuotaUs))
			}

			cpuPeriodUs, err := exec.Command("cat", "/sys/fs/cgroup/cpu/hind/TestV1fsManager_Set/cpu.cfs_period_us").Output()
			if err != nil {
				t.Errorf("V1fsManager.Set() read cpu.cfs_period_us error: %v", err)
			} else if string(cpuPeriodUs) != "100000\n" {
				t.Errorf("V1fsManager.Set() read cpu.cfs_period_us error: got %v, want 100000", string(cpuPeriodUs))
			}

			cpuSetCpus, err := exec.Command("cat", "/sys/fs/cgroup/cpuset/hind/TestV1fsManager_Set/cpuset.cpus").Output()
			if err != nil {
				t.Errorf("V1fsManager.Set() read cpuset.cpus error: %v", err)
			} else if string(cpuSetCpus) != "0\n" {
				t.Errorf("V1fsManager.Set() read cpuset.cpus error: got %v, want 0", string(cpuSetCpus))
			}

			memoryLimitBytes, err := exec.Command("cat", "/sys/fs/cgroup/memory/hind/TestV1fsManager_Set/memory.limit_in_bytes").Output()
			if err != nil {
				t.Errorf("V1fsManager.Set() read memory.limit_in_bytes error: %v", err)
			} else if string(memoryLimitBytes) != "999424\n" {
				t.Errorf("V1fsManager.Set() read memory.limit_in_bytes error: got %v, want 1000000", string(memoryLimitBytes))
			}

		})
	}
}

func TestV1fsManager_setResource(t *testing.T) {
	basePath := "/sys/fs/cgroup"
	cgroupName := "hind/TestV1fsManager_setResource"

	type fields struct {
		BasePath   string
		CgroupName string
	}
	type args struct {
		r Resource
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"cpu", fields{basePath, cgroupName}, args{CpuPeriodUs(1000)}, false},
		{"cpuset", fields{basePath, cgroupName}, args{CpuSetCpus("0")}, false},
		{"memory", fields{basePath, cgroupName}, args{MemoryLimitBytes(1000000)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &V1fsManager{
				BasePath: tt.fields.BasePath,
			}
			if err := v.Create(tt.fields.CgroupName); err != nil {
				t.Fatalf("V1fsManager.Create() error = %v", err)
			}

			if err := v.setResource(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("V1fsManager.setResource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: implement me
func TestV1fsManager_Destroy(t *testing.T) {
	type fields struct {
		BasePath   string
		CgroupName string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &V1fsManager{
				BasePath:   tt.fields.BasePath,
				cgroupName: tt.fields.CgroupName,
			}
			v.Destroy()
		})
	}
}

func Test_v1fsPath(t *testing.T) {
	type args struct {
		basePath   string
		subsystem  string
		cgroupName string
		fileName   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := v1fsPath(tt.args.basePath, tt.args.subsystem, tt.args.cgroupName, tt.args.fileName); got != tt.want {
				t.Errorf("v1fsPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TODO: implement me
func TestV1fsManager_checkBasePath(t *testing.T) {
	type fields struct {
		BasePath   string
		cgroupName string
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &V1fsManager{
				BasePath:   tt.fields.BasePath,
				cgroupName: tt.fields.cgroupName,
			}
			got, err := v.checkBasePath()
			if (err != nil) != tt.wantErr {
				t.Errorf("V1fsManager.checkBasePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("V1fsManager.checkBasePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TODO: implement me
func TestV1fsManager_checkCreated(t *testing.T) {
	type fields struct {
		BasePath   string
		cgroupName string
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &V1fsManager{
				BasePath:   tt.fields.BasePath,
				cgroupName: tt.fields.cgroupName,
			}
			got, err := v.checkCreated()
			if (err != nil) != tt.wantErr {
				t.Errorf("V1fsManager.checkCreated() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("V1fsManager.checkCreated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_randCgroupName(t *testing.T) {
	times := 1023
	gots := make([]string, 0, times)
	for i := 0; i < times; i++ {
		got := randCgroupName()

		if len(got) < 6 {
			t.Errorf("randCgroupName() = %v, want %v", got, "len > 6")
		}

		for _, v := range gots {
			if v == got {
				t.Errorf("randCgroupName() = %v, want %v", got, "not repeat")
			}
		}

		gots = append(gots, got)
		// t.Logf("randCgroupName() = %v", got)
	}
}
