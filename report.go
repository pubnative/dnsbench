package dnsbench

import (
	"context"
	"net"
	"time"
)

type DNSReport struct {
	Dur      time.Duration
	ConnReps []ConnReport
	Err      error
}

type ConnReport struct {
	Dur time.Duration
	IP  net.IP
	Err error
}

var dialer = &net.Dialer{}

func NewDNSReport(host string) *DNSReport {
	now := time.Now()
	ips, err := net.LookupIP(host)
	rep := &DNSReport{
		Dur: time.Since(now),
		Err: err,
	}
	if err == nil {
		rep.setConnReport(ips)
	}
	return rep
}

func (r *DNSReport) setConnReport(ips []net.IP) {
	reps := make([]ConnReport, len(ips))
	for i := range ips {
		ip := "[" + ips[i].String() + "]:80"
		ctx := context.Background()
		now := time.Now()
		conn, err := dialer.DialContext(ctx, "tcp", ip)
		reps[i] = ConnReport{
			Dur: time.Since(now),
			IP:  ips[i],
			Err: err,
		}
		conn.Close()
	}

	r.ConnReps = reps
}
