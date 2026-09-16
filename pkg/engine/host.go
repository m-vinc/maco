package engine

import (
	"runtime"

	"golang.org/x/sys/unix"
)

type HostInfo struct {
	CPUs        int    `json:"cpus"`
	MemoryBytes uint64 `json:"memory_bytes"`
}

func (e *Engine) HostInfo() HostInfo {
	info := HostInfo{CPUs: runtime.NumCPU()}
	if mem, err := unix.SysctlUint64("hw.memsize"); err == nil {
		info.MemoryBytes = mem
	}

	return info
}
