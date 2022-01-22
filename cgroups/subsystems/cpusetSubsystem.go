package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpusetSubSystem struct {
}

// Set 设置cgroupPath对应的cgroup的内存限制
func (m *CpusetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	// cgroupPath的作用是获取当前subsystem在虚拟文件系统中的路径
	if subsysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, true); err == nil {
		if res.MemoryLimit != "" {
			// 设置这个cgroup的内存限制，将限制写入到cgroup对应目录的memory.limit_in_bytes文件中
			if err = ioutil.WriteFile(path.Join(subsysCgroupPath, "cpuset.cpus"), []byte(res.CpuSet), 0644); err != nil {
				return fmt.Errorf("set cgroup cpuset fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

// Name 返回cgroup的名字
func (m *CpusetSubSystem) Name() string {
	return "cpuset"
}

// Remove 删除cgrouppath对应的cgroup
func (m *CpusetSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, false); err == nil {
		return os.Remove(subsysCgroupPath)
	} else {
		return err
	}
}

// Apply 将一个进程加入到cgroupPath对应的cgroup中
func (m *CpusetSubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, false); err == nil {
		// 把进程pid写到cgroup的虚拟文件系统对应目录下的"task"文件中
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
}
