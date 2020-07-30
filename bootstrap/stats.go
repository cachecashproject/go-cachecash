package bootstrap

import (
	"syscall"

	"github.com/capnm/sysinfo"
)

type CacheStats struct {
	FreeMemory  uint64
	TotalMemory uint64

	FreeDisk  uint64
	TotalDisk uint64
}

func NewCacheStats() *CacheStats {
	return &CacheStats{}
}

func (cs *CacheStats) ReadMemoryStats() {
	// TODO: make sure we consider buffers as free
	si := sysinfo.Get()
	cs.FreeMemory = si.FreeRam
	cs.TotalMemory = si.TotalRam
}

func (cs *CacheStats) ReadDiskStats(path string) error {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return err
	}

	cs.FreeDisk = stat.Bavail * uint64(stat.Bsize)
	cs.TotalDisk = stat.Blocks * uint64(stat.Bsize)

	return nil
}
