# hind: yet another naive container runtime

> HIND Is Not Docker.

## Usage

```
go build .
sudo ./hind run -it NOIMG /bin/sh
sudo ./hind run -it NOIMG /bin/ls -alh
```

- `NOIMG`: run image is not implemented. A placeholder.

### cgroups

宿主机：

```sh
$ stress --vm-bytes 800m --vm-keep -m 1
stress: info: [6364] dispatching hogs: 0 cpu, 0 io, 1 vm, 0 hdd
$ top
  PID USER      PR  NI    VIRT    RES    SHR S  %CPU %MEM     TIME+ COMMAND
 6365 foo       20   0  826524 819500    460 R 100.0 10.1   0:08.43 stress
```

容器：

```sh
$ sudo go run . run -ti --memory-limit-bytes=1024000 noimg stress --vm-bytes 800m --vm-keep -m 1
2023/06/17 11:39:11 INFO [cmd/run] Create and run a new container opts="{Tty:true Interactive:true Image:noimg Command:[stress --vm-bytes 800m --vm-keep -m 1] Resources:{CpuQuotaUs:0 CpuPeriodUs:0 CpuSetCpus: MemoryLimitBytes:1024000}}"
2023/06/17 11:39:11 INFO [host] Run command in container tty=true command="[stress --vm-bytes 800m --vm-keep -m 1]" resources="{CpuQuotaUs:0 CpuPeriodUs:0 CpuSetCpus: MemoryLimitBytes:1024000}"
2023/06/17 11:39:11 INFO [host] container process started pid=6719
2023/06/17 11:39:11 INFO [cmd/init] Booting container
2023/06/17 11:39:11 INFO [container] RunContainerInitProcess: bootstrapping container
2023/06/17 11:39:11 INFO [cgroups] V1fsManager setResource value=0 target=/sys/fs/cgroup/cpuset/hind/container/cpuset.cpus
2023/06/17 11:39:11 INFO [cgroups] V1fsManager setResource value=0 target=/sys/fs/cgroup/cpuset/hind/container/cpuset.mems
2023/06/17 11:39:11 INFO [host] Cgroup manager created manager="&{BasePath:/sys/fs/cgroup/ cgroupName:hind/container}"
2023/06/17 11:39:11 INFO [cgroups] V1fsManager setResource value=1024000 target=/sys/fs/cgroup/memory/hind/container/memory.limit_in_bytes
2023/06/17 11:39:11 INFO [host] Cgroup setup done
2023/06/17 11:39:11 INFO [host] Command sent, closing the pipe (w).
2023/06/17 11:39:11 INFO [container] pid 1 received command command="[stress --vm-bytes 800m --vm-keep -m 1]"
2023/06/17 11:39:11 INFO [container] pid 1 setup mount
2023/06/17 11:39:11 INFO [container] pid 1 found command in path exe=/bin/stress
2023/06/17 11:39:11 INFO [container] pid 1 ready to execve the command. Bootstrapping done. Bye. command="[stress --vm-bytes 800m --vm-keep -m 1]"
stress: info: [1] dispatching hogs: 0 cpu, 0 io, 1 vm, 0 hdd
stress: FAIL: [1] (415) <-- worker 6 got signal 9
stress: WARN: [1] (417) now reaping child worker processes
stress: FAIL: [1] (421) kill error: No such process
stress: FAIL: [1] (451) failed run completed in 11s
2023/06/17 11:39:22 INFO [host] container process exited
2023/06/17 11:39:22 WARN [cgroups] TODO: V1fsManager.Destroy()
```

内存不够用，stress 被 OOM 杀掉了。

整小一点：

```sh
$ sudo go run . run -ti --memory-limit-bytes=1024000 noimg stress --vm-bytes 1m --vm-keep -m 1
```

可以运行了，：

```sh
  PID USER      PR  NI    VIRT    RES    SHR S  %CPU %MEM     TIME+ COMMAND
 8850 root      20   0    8348    660    444 D   8.6  0.0   0:00.76 stress
```

一直交换，所以 CPU 占用也下来了。

放宽限制 CPU 就上来了：

```sh
$ sudo go run . run -ti --memory-limit-bytes=2024000 noimg stress --vm-bytes 1m --vm-keep -m 1
$ top
  PID USER      PR  NI    VIRT    RES    SHR S  %CPU %MEM     TIME+ COMMAND
10269 root      20   0    8348   1624    532 R  93.8  0.0   0:04.86 stress
```
