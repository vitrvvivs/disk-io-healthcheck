package healthchecks

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type HTTPHealthcheck struct {
	healthy bool
	client *http.Client

	// Config
	URL string
	Interval uint64 // seconds
}

func (h *HTTPHealthcheck) Init() {
	flag.StringVar(&h.URL, "next.url", "", "Proxy another http healthcheck")
	flag.Uint64Var(&h.Interval, "next.interval", 1, "Seconds between checking other url")
}

func (h *HTTPHealthcheck) Start(ctx context.Context) {
	u, err := url.Parse(h.URL)
	if err != nil {
		fmt.Println("HTTPHealthcheck.Start", err)
	}
	if u.Scheme == "https" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		h.client = &http.Client{Transport: tr} 
	} else {
		h.client = &http.Client{}
	}
	go (func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second * time.Duration(h.Interval)):
				h.Update()
			}
		}
	})()
}

func (h *HTTPHealthcheck) Update() bool {
	if h.URL == "" {
		h.healthy = true
		return true
	}
	resp, err := h.client.Get(h.URL)
	if err != nil {
		h.healthy = false
		fmt.Println("HTTPHealthcheck.Update", err)
		return h.healthy
	}
	h.healthy = (resp.StatusCode == 200)
	return h.healthy
}

func (h *HTTPHealthcheck) Healthy() bool {
	return h.healthy
}
