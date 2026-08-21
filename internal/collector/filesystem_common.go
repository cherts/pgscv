// Package collector is a pgSCV collectors
package collector

import (
	"os"
	"strings"

	"github.com/cherts/pgscv/internal/log"
	"github.com/shirou/gopsutil/v4/disk"
)

// mount describes properties of mounted filesystems
type mount struct {
	device     string
	mountpoint string
	fstype     string
	options    []string
}

// parseMounts parses disk partition info and returns slice of mounted filesystems properties.
func parseMounts(r []disk.PartitionStat) ([]mount, error) {
	log.Debug("parse mounted filesystems")
	var (
		mounts []mount
	)

	// Parse line by line, split line to param and value, parse the value to float and save to store.
	for _, diskData := range r {
		s := mount{
			device:     diskData.Device,
			mountpoint: diskData.Mountpoint,
			fstype:     diskData.Fstype,
			options:    diskData.Opts,
		}
		mounts = append(mounts, s)
	}

	return mounts, nil
}

// truncateDeviceName truncates passed full path to device to short device name.
func truncateDeviceName(path string) string {
	if path == "" {
		log.Warnf("cannot truncate empty device path")
		return ""
	}
	// Make name which will be returned in case of later errors occurred.
	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]

	// Check device path exists.
	fi, err := os.Lstat(path)
	if err != nil {
		log.Debugf("%s, use default '%s'", err, name)
		return name
	}

	// If path is symlink, try dereference it.
	if fi.Mode()&os.ModeSymlink != 0 {
		resolved, err := os.Readlink(path)
		if err != nil {
			log.Warnf("%s, use name's last part '%s'", err, name)
			return name
		}
		// Swap name to dereferenced origin.
		parts := strings.Split(resolved, "/")
		name = parts[len(parts)-1]
	}

	// Return default (or dereferenced) name.
	return name
}
