package diskstats

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

/*
	The /proc/diskstats file displays the I/O statistics
	of block devices. Each line contains the following 14
	fields:

	==  ===================================
	 1  major number
	 2  minor number
	 3  device name
	 4  reads completed successfully
	 5  reads merged
	 6  sectors read
	 7  time spent reading (ms)
	 8  writes completed
	 9  writes merged
	10  sectors written
	11  time spent writing (ms)
	12  I/Os currently in progress
	13  time spent doing I/Os (ms)
	14  weighted time spent doing I/Os (ms)
	==  ===================================

	Kernel 4.18+ appends four more fields for discard
	tracking putting the total at 18:

	==  ===================================
	15  discards completed successfully
	16  discards merged
	17  sectors discarded
	18  time spent discarding
	==  ===================================

	Kernel 5.5+ appends two more fields for flush requests:

	==  =====================================
	19  flush requests completed successfully
	20  time spent flushing
	==  =====================================
*/

const (
	Major = 0
	Minor = 1
	Name = 2

	Read = 3
	ReadMerged = 4
	ReadSectors = 5
	ReadTime = 6

	Write = 7
	WriteMerged = 8
	WriteSectors = 9
	WriteTime = 10

	InFlight = 11
	TotalTime = 12
	WeightedTime = 13
)

type Statline struct {
	Name string
	Data [14]uint64
}

type Diskstats struct {
	path string
	interval time.Duration

	stats map[string]*Statline
	delta map[string]*Statline
	fd *os.File
}

func New(path string) (*Diskstats, error) {
	d := &Diskstats{
		path: path,
		interval: 1 * time.Second,
	}
	fmt.Println("Creating Diskstats")
	fd, err := os.Open("/proc/diskstats")
	if err != nil {
		return nil, err
	}
	d.fd = fd
	d.stats = make(map[string]*Statline)
	d.delta = make(map[string]*Statline)

	err = d.Update()
	return d, err
}

func ParseLine(line string) (*Statline, error) {
	var err error
	sl := &Statline{}
	fields := strings.Fields(line)
	for i := 0; i < 14; i++ {
		field := fields[i]
		if i == 2 {
			sl.Name = field
		} else {
			sl.Data[i], err = strconv.ParseUint(field, 10, 64)
			if err != nil {
				return nil, err
			}
		}

	}
	return sl, nil
}

func (d *Diskstats) Update() (error) {
	d.fd.Seek(0, 0)
	scanner := bufio.NewScanner(d.fd)
	for scanner.Scan() {
		sl, err := ParseLine(scanner.Text())
		if err != nil {
			return err
		}
		name := sl.Name
		old, ok := d.stats[name]
		if ok {
			delta := &Statline{}
			for i, v := range old.Data {
				delta.Data[i] = sl.Data[i] - v
			}
			d.delta[name] = delta
		}
		d.stats[name] = sl
	}
	return nil
}

func (sl *Statline) Rate() (readBs uint64, writeBs uint64) {
	return sl.Data[ReadSectors] * 512, sl.Data[WriteSectors] * 512
}

func (d *Diskstats) Get(name string) (*Statline) {
	name, _ = filepath.EvalSymlinks(name)
	name = strings.Trim(name, "/dev/")
	fmt.Printf("Getting %s", name)
	return d.stats[name]
}

func (d *Diskstats) GetDelta(name string) (*Statline) {
	new_name, err := filepath.EvalSymlinks(name)
	if err == nil {
		name = new_name
		fmt.Printf("Getting Delta for %s\n", name)
		name = strings.TrimPrefix(name, "/dev/")
	}
	fmt.Printf("Getting Delta for %s\n", name)
	return d.delta[name]
}

func (d *Diskstats) Print() {
	for name, sl := range d.stats {
		fmt.Printf("%s %d %d\n", name, sl.Data[Read], sl.Data[Write])
	}
}

func (d *Diskstats) StartWorker(interval time.Duration, ctx context.Context) {
	d.interval = interval
	go (func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
				d.Update()
			}
		}
	})()
}
