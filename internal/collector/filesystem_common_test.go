package collector

import (
	"testing"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/stretchr/testify/assert"
)

func Test_parseMounts(t *testing.T) {
	var diskStat = []disk.PartitionStat{
		{Device: "/dev/vda1", Mountpoint: "/", Fstype: "ext4", Opts: []string{"rw", "relatime"}},
		{Device: "/dev/vda2", Mountpoint: "/var/lib/postgres", Fstype: "ext4", Opts: []string{"rw", "relatime"}},
	}

	stats, err := parseMounts(diskStat)
	assert.NoError(t, err)

	want := []mount{
		{device: "/dev/vda1", mountpoint: "/", fstype: "ext4", options: []string{"rw", "relatime"}},
		{device: "/dev/vda2", mountpoint: "/var/lib/postgres", fstype: "ext4", options: []string{"rw", "relatime"}},
	}

	assert.Equal(t, want, stats)
}

func Test_truncateDeviceName(t *testing.T) {
	var testcases = []struct {
		name string
		path string
		want string
	}{
		{name: "valid 1", path: "testdata/dev/sda", want: "sda"},
		{name: "valid 2", path: "testdata/dev/sdb2", want: "sdb2"},
		{name: "valid 3", path: "testdata/dev/mapper/ssd-root", want: "dm-1"},
		{name: "unknown", path: "testdata/dev/unknown", want: "unknown"},
	}

	for _, tc := range testcases {
		assert.Equal(t, tc.want, truncateDeviceName(tc.path))
	}
}
