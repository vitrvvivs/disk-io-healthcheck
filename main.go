package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/vitrvvivs/disk-io-healthcheck/diskstats"
)

type Server struct {
	// configurable
	ListenAddr string
	ListenPort uint16
	DiskstatsInterval uint64 // seconds
	DiskDevice string

	MaxReadKBs uint64
	MaxWriteKBs uint64
	MaxTotalKBs uint64

	// state
	ds *diskstats.Diskstats
}


func (server *Server) status(w http.ResponseWriter, req *http.Request) {
	delta := server.ds.GetDelta(server.DiskDevice)
	status := "healthy"
	if delta == nil {
		status := "not ready"
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "%s\n", status)
		return
	}

	readB, writeB := delta.Rate()
	read := uint64(float64(readB) / 1024 / float64(server.DiskstatsInterval))
	write := uint64(float64(writeB) / 1024 / float64(server.DiskstatsInterval))

	if (server.MaxReadKBs > 0 && read > server.MaxReadKBs) ||
	   (server.MaxWriteKBs > 0 && write > server.MaxWriteKBs) ||
	   (server.MaxTotalKBs > 0 && (read + write) > server.MaxTotalKBs) {
		    status = "unhealthy"
    		w.WriteHeader(http.StatusServiceUnavailable)
	}
	fmt.Printf("%s read %dKB/s  write %dKB/s\n", status, read, write)
	fmt.Fprintf(w, "%d %d\n", read, write)
	return
}


func (server *Server) Start() {
	ds, err := diskstats.New("/proc/diskstats")
	
	if err != nil {
		fmt.Println(err)
		return
	}
	server.ds = ds
	ds.StartWorker(time.Duration(server.DiskstatsInterval) * time.Second, context.Background())
	ds.Print()

	http.HandleFunc("/health", server.status)
	http.ListenAndServe(fmt.Sprintf("%s:%d", server.ListenAddr, server.ListenPort), nil)
}


func main() {
	var (
		ListenAddr string
		ListenPort uint
		DiskstatsInterval uint64 // seconds
		DiskDevice string

		MaxReadKBs uint64
		MaxWriteKBs uint64
		MaxTotalKBs uint64
	)

	flag.StringVar(&ListenAddr, "addr", "0.0.0.0", "Listen Address")
	flag.UintVar(&ListenPort, "port", 8010, "Listen Port")
	flag.Uint64Var(&DiskstatsInterval, "interval", 5, "Seconds between checks")
	flag.StringVar(&DiskDevice, "device", "/dev/sda", "Path of device to watch")

	flag.Uint64Var(&MaxReadKBs, "max-read", 0, "Max read KB/s to consider healthy")
	flag.Uint64Var(&MaxWriteKBs, "max-write", 0, "Max write KB/s to consider healthy")
	flag.Uint64Var(&MaxTotalKBs, "max-total", 0, "Max total KB/s to consider healthy")
	flag.Parse()

	server := &Server{
		ListenAddr: ListenAddr,
		ListenPort: uint16(ListenPort),
		DiskstatsInterval: DiskstatsInterval,
		DiskDevice: DiskDevice,
		MaxReadKBs: MaxReadKBs,
		MaxWriteKBs: MaxWriteKBs,
		MaxTotalKBs: MaxTotalKBs,
	}
	server.Start()
}
