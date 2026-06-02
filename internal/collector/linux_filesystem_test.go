package collector

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/cherts/pgscv/internal/filter"
	"github.com/cherts/pgscv/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestFilesystemCollector_Update(t *testing.T) {
	var input = pipelineInput{
		required: []string{
			"node_filesystem_bytes",
			"node_filesystem_bytes_total",
			"node_filesystem_files",
			"node_filesystem_files_total",
		},
		collector:         NewFilesystemCollector,
		collectorSettings: model.CollectorSettings{Filters: filter.New()},
	}

	pipeline(t, input)
}

func Test_getFilesystemStats(t *testing.T) {
	got, err := getFilesystemStats()
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Greater(t, len(got), 0)
}

func Test_parseFilesystemStats(t *testing.T) {
	file, err := os.Open(filepath.Clean("testdata/proc/mounts.golden"))
	assert.NoError(t, err)

	stats, err := parseFilesystemStats(file)
	assert.NoError(t, err)
	assert.Greater(t, len(stats), 1)
	assert.Greater(t, stats[0].size, float64(0))
	assert.Greater(t, stats[0].free, float64(0))
	assert.Greater(t, stats[0].avail, float64(0))
	assert.Greater(t, stats[0].files, float64(0))
	assert.Greater(t, stats[0].filesfree, float64(0))

	_ = file.Close()

	// test with wrong format file
	file, err = os.Open(filepath.Clean("testdata/proc/netdev.golden"))
	assert.NoError(t, err)

	stats, err = parseFilesystemStats(file)
	assert.Error(t, err)
	assert.Nil(t, stats)
	_ = file.Close()
}

func Test_readMountpointStat(t *testing.T) {
	stat, err := readMountpointStat("/")
	assert.NoError(t, err)
	assert.Greater(t, stat.size, float64(0))
	assert.Greater(t, stat.free, float64(0))
	assert.Greater(t, stat.avail, float64(0))
	assert.Greater(t, stat.files, float64(0))
	assert.Greater(t, stat.filesfree, float64(0))

	// unknown filesystem
	stat, err = readMountpointStat("/invalid")
	assert.Error(t, err)
}

func Test_readMountpointStatWithTimeout(t *testing.T) {
	stat, err := readMountpointStatWithTimeout("/", time.Second)
	assert.NoError(t, err)
	assert.Greater(t, stat.Blocks, uint64(0))

	// unknown filesystem
	_, err = readMountpointStatWithTimeout("/invalid", time.Second)
	assert.Error(t, err)
}

func Test_readMountpointStat_timeout(t *testing.T) {
	// Save originals
	origTimeout := mountpointStatTimeout
	origFunc := readMountpointStatWithTimeoutFunc
	defer func() {
		mountpointStatTimeout = origTimeout
		readMountpointStatWithTimeoutFunc = origFunc
	}()

	// Shorten timeout for test
	mountpointStatTimeout = 10 * time.Millisecond

	// Fake function that sleeps longer than timeout to simulate stuck syscall
	readMountpointStatWithTimeoutFunc = func(mountpoint string, timeout time.Duration) (*syscall.Statfs_t, error) {
		time.Sleep(50 * time.Millisecond)
		// return a valid stat to simulate late success
		return &syscall.Statfs_t{}, nil
	}

	stat, err := readMountpointStat("/")
	assert.Error(t, err)
	assert.Equal(t, errFilesystemTimedOut, err)
	// stat should carry the error as well
	assert.Equal(t, errFilesystemTimedOut, stat.err)
}
