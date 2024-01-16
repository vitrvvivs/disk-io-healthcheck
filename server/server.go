package server

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/vitrvvivs/disk-io-healthcheck/healthchecks"
)

type Server struct {
	// configurable
	ListenAddr string
	ListenPort uint

	Healthchecks []healthchecks.Healthcheck
}


func (server *Server) status(w http.ResponseWriter, req *http.Request) {
	for _, hc := range server.Healthchecks {
		if !hc.Healthy() {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}
	fmt.Fprintf(w, "\n")
	return
}


func (server *Server) Start() {
	for _, hc := range server.Healthchecks {
		hc.Start(context.Background())
	}
	http.HandleFunc("/health", server.status)
	http.ListenAndServe(fmt.Sprintf("%s:%d", server.ListenAddr, server.ListenPort), nil)
}


func (server *Server) Init() {
	flag.StringVar(&server.ListenAddr, "addr", "0.0.0.0", "Listen Address")
	flag.UintVar(&server.ListenPort, "port", 8010, "Listen Port")

	http := &healthchecks.HTTPHealthcheck{}
	http.Init()
	disk := &healthchecks.DiskstatsHealthcheck{}
	disk.Init()

	server.Healthchecks = []healthchecks.Healthcheck{disk, http}
}
