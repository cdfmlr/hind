package cgroups

import (
	"reflect"
	"testing"
)

// make sure all fields of the Resources struct implement the
// {Resource, V1fsResource} interface.
func Test_ResourcesFields(t *testing.T) {
	interfaceTypes := []reflect.Type{
		// reflect.TypeOf((*Xxx)(nil)).Elem() is the reflect.Type of an interface Xxx
		reflect.TypeOf((*Resource)(nil)).Elem(),
		reflect.TypeOf((*V1fsResource)(nil)).Elem(),
	}

	for _, interfaceType := range interfaceTypes {
		if interfaceType.Kind() != reflect.Interface {
			t.Fatalf("type %v is not an interface", interfaceType.Name())
		}

		t.Run("impl_"+interfaceType.Name(), func(t *testing.T) {
			test_ResourcesFieldsImplInterfacet(t, interfaceType)
		})
	}
}

func test_ResourcesFieldsImplInterfacet(t *testing.T, targetInterface reflect.Type) {
	resources := Resources{}

	resourcesValue := reflect.ValueOf(resources)
	resourcesType := reflect.TypeOf(resources)

	for i := 0; i < resourcesValue.NumField(); i++ {
		fieldType := resourcesType.Field(i) // fieldType is not the type of the field. it's to get the field name in the struct.

		t.Run("field_"+fieldType.Name, func(t *testing.T) {
			fieldValue := resourcesValue.Field(i)

			ok := fieldValue.Type().Implements(targetInterface)
			if !ok {
				t.Errorf("Resources field[%v] %v (type %v) does not implement interface %v",
					i, fieldType.Name, fieldType.Type.Name(), targetInterface.Name())
			}
		})
	}
}

func TestCpuQuotaUs_Value(t *testing.T) {
	tests := []struct {
		name string
		q    CpuQuotaUs
		want string
	}{
		{"CpuQuotaUs", CpuQuotaUs(42), "42"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.q.Value(); got != tt.want {
				t.Errorf("CpuQuotaUs.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}



func TestCpuQuotaUs_V1fsPath(t *testing.T) {
	type args struct {
		basePath    string
		containerId string
	}
	tests := []struct {
		name string
		q    CpuQuotaUs
		args args
		want string
	}{
		{name: "CpuQuotaUs", q: CpuQuotaUs(42), args: args{
			basePath: "/sys/fs/cgroup", containerId: "testgroup",
		}, want: "/sys/fs/cgroup/cpu/testgroup/cpu.cfs_quota_us"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.q.V1fsPath(tt.args.basePath, tt.args.containerId); got != tt.want {
				t.Errorf("CpuQuotaUs.V1fsPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TODO: 👇

func TestCpuPeriodUs_Value(t *testing.T) {
	tests := []struct {
		name string
		p    CpuPeriodUs
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Value(); got != tt.want {
				t.Errorf("CpuPeriodUs.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCpuPeriodUs_V1fsPath(t *testing.T) {
	type args struct {
		basePath    string
		containerId string
	}
	tests := []struct {
		name string
		p    CpuPeriodUs
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.V1fsPath(tt.args.basePath, tt.args.containerId); got != tt.want {
				t.Errorf("CpuPeriodUs.V1fsPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCpuSetCpus_Value(t *testing.T) {
	tests := []struct {
		name string
		c    CpuSetCpus
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Value(); got != tt.want {
				t.Errorf("CpuSetCpus.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCpuSetCpus_V1fsPath(t *testing.T) {
	type args struct {
		basePath    string
		containerId string
	}
	tests := []struct {
		name string
		c    CpuSetCpus
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.V1fsPath(tt.args.basePath, tt.args.containerId); got != tt.want {
				t.Errorf("CpuSetCpus.V1fsPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryLimitBytes_Value(t *testing.T) {
	tests := []struct {
		name string
		m    MemoryLimitBytes
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Value(); got != tt.want {
				t.Errorf("MemoryLimitBytes.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryLimitBytes_V1fsPath(t *testing.T) {
	type args struct {
		basePath    string
		containerId string
	}
	tests := []struct {
		name string
		m    MemoryLimitBytes
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.V1fsPath(tt.args.basePath, tt.args.containerId); got != tt.want {
				t.Errorf("MemoryLimitBytes.V1fsPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
