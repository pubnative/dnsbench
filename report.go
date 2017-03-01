package dnsbench

import (
	"context"
	"net"
	"time"
)

type DNSReport struct {
	ConnReps []ConnReport
	dur      time.Duration
	err      error
}

type ConnReport struct {
	IP  net.IP
	dur time.Duration
	err error
}

type Report interface {
	Dur() time.Duration
	Err() error
}

var dialer = &net.Dialer{}

var _ Report = (*DNSReport)(nil)
var _ Report = (*ConnReport)(nil)

func NewDNSReport(host string) *DNSReport {
	now := time.Now()
	ips, err := net.LookupIP(host)
	rep := &DNSReport{
		dur: time.Since(now),
		err: err,
	}
	if err == nil {
		rep.setConnReport(ips)
	}
	return rep
}

func (r *DNSReport) Dur() time.Duration {
	return r.dur
}

func (r *DNSReport) Err() error {
	return r.err
}

func (r *DNSReport) setConnReport(ips []net.IP) {
	reps := make([]ConnReport, len(ips))
	for i := range ips {
		ip := "[" + ips[i].String() + "]:80"
		ctx := context.Background()
		now := time.Now()
		conn, err := dialer.DialContext(ctx, "tcp", ip)
		reps[i] = ConnReport{
			dur: time.Since(now),
			err: err,
			IP:  ips[i],
		}
		conn.Close()
	}

	r.ConnReps = reps
}

func (r ConnReport) Dur() time.Duration {
	return r.dur
}

func (r ConnReport) Err() error {
	return r.err
}
