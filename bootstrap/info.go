package bootstrap

import (
	"syscall"

	"github.com/capnm/sysinfo"
)

type CacheInfo struct {
	FreeMemory  uint64
	TotalMemory uint64

	FreeDisk  uint64
	TotalDisk uint64
}

func NewCacheInfo() *CacheInfo {
	return &CacheInfo{}
}

func (ci *CacheInfo) ReadMemoryStats() {
	// TODO: make sure we consider buffers as free
	si := sysinfo.Get()
	ci.FreeMemory = si.FreeRam
	ci.TotalMemory = si.TotalRam
}

func (ci *CacheInfo) ReadDiskStats(path string) error {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return err
	}

	ci.FreeDisk = stat.Bavail * uint64(stat.Bsize)
	ci.TotalDisk = stat.Blocks * uint64(stat.Bsize)

	return nil
}
