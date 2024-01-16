package healthchecks

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/vitrvvivs/disk-io-healthcheck/diskstats"
	"github.com/vitrvvivs/disk-io-healthcheck/movingaverage"
)

type DiskstatsHealthcheck struct {
	healthy bool
	ds *diskstats.Diskstats
	read *movingaverage.MovingAverage[uint64]
	write *movingaverage.MovingAverage[uint64]
	total *movingaverage.MovingAverage[uint64]

	// Config
	DiskstatsInterval uint64 // seconds
	AverageCount uint64 // seconds
	DiskDevice string
	MaxReadKBs uint64
	MaxWriteKBs uint64
	MaxTotalKBs uint64
}

func (h *DiskstatsHealthcheck) Init() {
	flag.StringVar(&h.DiskDevice, "disk.device", "/dev/sda", "Path of device to watch")
	flag.Uint64Var(&h.DiskstatsInterval, "disk.interval", 1, "Seconds between reading /proc/diskstats")
	flag.Uint64Var(&h.AverageCount, "disk.average", 5, "Number of datapoints to average out")

	flag.Uint64Var(&h.MaxReadKBs, "disk.read", 0, "Max read KB/s to consider healthy")
	flag.Uint64Var(&h.MaxWriteKBs, "disk.write", 0, "Max write KB/s to consider healthy")
	flag.Uint64Var(&h.MaxTotalKBs, "disk.total", 0, "Max total KB/s to consider healthy")

	var err error
	h.ds, err = diskstats.New("/proc/diskstats")
	if err != nil {
		fmt.Println(err)
		return
	}
	h.read = movingaverage.New[uint64](int(h.AverageCount))
	h.write = movingaverage.New[uint64](int(h.AverageCount))
	h.total = movingaverage.New[uint64](int(h.AverageCount))
}

func (h *DiskstatsHealthcheck) Start(ctx context.Context) {
	h.ds.StartWorker(time.Duration(h.DiskstatsInterval) * time.Second, ctx)
	go (func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				h.Update()
			}
		}
	})()
}

func (h *DiskstatsHealthcheck) Update() bool {
	h.healthy = true
	delta := h.ds.GetDelta(h.DiskDevice)
	if delta == nil {
		h.healthy = false
		fmt.Printf("Could not get delta for %s (%s). Might just need to wait for more data\n", h.DiskDevice, diskstats.SanitizeDiskName(h.DiskDevice))
		return h.healthy
	}

	readB, writeB := delta.Rate()
	// convert to kb/s
	read := uint64(float64(readB / 1024) / float64(h.DiskstatsInterval))
	write := uint64(float64(writeB / 1024) / float64(h.DiskstatsInterval))
	total := read + write

	// add to moving averages
	h.read.Update(read)
	h.write.Update(write)
	h.total.Update(total)

	if (h.MaxReadKBs > 0 && uint64(h.read.Average) > h.MaxReadKBs) ||
	   (h.MaxWriteKBs > 0 && uint64(h.write.Average) > h.MaxWriteKBs) ||
	   (h.MaxTotalKBs > 0 && uint64(h.total.Average) > h.MaxTotalKBs) {
		    h.healthy = false
	}
	fmt.Printf("DiskstatsHealthcheck.Update: %t %dKB/s read %dKB/s write\n", h.healthy, uint64(h.read.Average), uint64(h.write.Average))
	return h.healthy
}

func (h *DiskstatsHealthcheck) Healthy() bool {
	return h.healthy
}
