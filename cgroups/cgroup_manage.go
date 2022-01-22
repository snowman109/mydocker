package cgroups

import (
	"fmt"
	"mdocker/cgroups/subsystems"
)

type CgroupManager struct {
	// cgroup在hierarchy中的路径，相当于创建的cgroup目录相对于各root cgroup目录的路径
	Path     string
	Resource *subsystems.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{Path: path}
}

// Apply 将进程PID加入到每个cgroup中
func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range subsystems.SubsystemIns {
		subSysIns.Apply(c.Path, pid)
	}
	return nil
}

// Set 设置各个subsystem挂在中的cgroup资源限制
func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, subSysIns := range subsystems.SubsystemIns {
		subSysIns.Set(c.Path, res)
	}
	return nil
}

// 释放各个subsystem挂载中的cgroup
func (c *CgroupManager) Destroy() error {
	for _, subSysIns := range subsystems.SubsystemIns {
		if err := subSysIns.Remove(c.Path); err != nil {
			fmt.Printf("remove cgroup fail %v\n", err)
		}
	}
	return nil
}
