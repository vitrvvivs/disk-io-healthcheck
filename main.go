package main

import (
	"flag"

	"github.com/vitrvvivs/disk-io-healthcheck/server"
)

func main() {
	s := &server.Server{}
	s.Init()
	flag.Parse()
	s.Start()
}
