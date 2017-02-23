package main

import (
	"fmt"
	"flag"
	"net"
	"time"
	"sort"
	"sync"
	"os"
)

var host = flag.String("h", "", "Hostname")
var loop = flag.Int("n", 1, "Number of lookups")

type report struct {
	dur time.Duration
	ips []net.IP
	err error
}

func init() {
	os.Setenv("GODEBUG", "netdns=go")
}

func main() {
	flag.Parse()

	resCh := make(chan report, *loop)
	wg := sync.WaitGroup{}
	for i := 0; i < *loop; i++ {
		wg.Add(1)
		go func() {
			now := time.Now()
			ips, err := net.LookupIP(*host)

			rep := report{
				dur: time.Since(now),
				ips: ips,
				err: err,
			}
			resCh <-rep
			wg.Done()
		}()
		if i % 100 == 0 {
			wg.Wait()
		}
	}
	var res []report
	for i := 0; i < *loop; i++ {
		res = append(res, <-resCh)
	}

	fmt.Printf("tests: %#v\n", len(res))
	fmt.Printf("errors: %d\n", numErr(res))
	fmt.Printf("min: %s\n", minDur(res))
	fmt.Printf("max: %s\n", maxDur(res))
	fmt.Printf("avg: %s\n", avgDur(res))
	fmt.Printf("median: %s\n", medianDur(res))
	fmt.Printf("top99: %s\n", topN(res, 99))
	fmt.Printf("top95: %s\n", topN(res, 95))
	fmt.Printf("top90: %s\n", topN(res, 90))
	fmt.Printf("ips: %#v\n", ips(res))
}

func minDur(res []report) time.Duration {
	if len(res) == 0 {
		return 0
	}

	m := res[0].dur
	for _, r := range res[1:] {
		if r.dur < m {
			m = r.dur
		}
	}

	return m
}

func maxDur(res []report) time.Duration {
	if len(res) == 0 {
		return 0
	}

	m := res[0].dur
	for _, r := range res[1:] {
		if r.dur > m {
			m = r.dur
		}
	}

	return m
}

func avgDur(res []report) time.Duration {
	if len(res) == 0 {
		return 0
	}

	var t int
	for _, r := range res {
		t += int(r.dur)
	}
	return time.Duration(t / len(res))
}

func medianDur(res []report) time.Duration {
	if len(res) == 0 {
		return 0
	}

	t := sorted(res)
	return t[len(t)/2]
}

func topN(res []report, n int) time.Duration {
	if len(res) == 0 {
		return 0
	}

	var dur []int
	for _, r := range res {
		dur = append(dur, int(r.dur))
	}
	sort.Ints(dur)

	idx := len(dur) * n / 100 - 1
	return time.Duration(dur[idx])
}

func sorted(res []report) []time.Duration {
	var dur []int
	for _, r := range res {
		dur = append(dur, int(r.dur))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(dur)))

	r := make([]time.Duration, len(dur))
	for i, d := range dur {
		r[i] = time.Duration(d)
	}
	return r
}

func ips(res []report) map[string]int {
	m := make(map[string]int)
	for _, r := range res {
		for _, ip := range r.ips {
			m[ip.String()] += 1
		}
	}
	return m
}

func numErr(res []report) int {
	n := 0
	for _, r := range res {
		if r.err != nil {
			n += 1
		}
	}
	return n
}
