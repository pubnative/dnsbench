package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/pubnative/dnsbench"
)

var host = flag.String("h", "", "Hostname")
var loop = flag.Int("n", 1, "Number of lookups")

func init() {
	os.Setenv("GODEBUG", "netdns=go")
}

func main() {
	flag.Parse()

	resCh := make(chan *dnsbench.DNSReport, *loop)
	wg := sync.WaitGroup{}
	for i := 0; i < *loop; i++ {
		wg.Add(1)
		go func() {
			resCh <- dnsbench.NewDNSReport(*host)
			wg.Done()
		}()
		if i%100 == 0 {
			wg.Wait()
		}
	}
	var res []*dnsbench.DNSReport
	for i := 0; i < *loop; i++ {
		res = append(res, <-resCh)
	}

	fmt.Println("DNS Resolve")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "tests \t errors \t min \t max \t avg \t median \t top99 \t top95 \t top90 \t ips \t")
	fmt.Fprintf(
		w,
		"%v \t %v \t %v \t %v \t %v \t %v \t %v \t %v \t %v \t %v \t\n",
		len(res), numErr(res), minDur(res), maxDur(res), avgDur(res),
		medianDur(res), topN(res, 99), topN(res, 95), topN(res, 90), ips(res),
	)
	w.Flush()
}

func minDur(res []*dnsbench.DNSReport) time.Duration {
	if len(res) == 0 {
		return 0
	}

	m := res[0].Dur
	for _, r := range res[1:] {
		if r.Dur < m {
			m = r.Dur
		}
	}

	return m
}

func maxDur(res []*dnsbench.DNSReport) time.Duration {
	if len(res) == 0 {
		return 0
	}

	m := res[0].Dur
	for _, r := range res[1:] {
		if r.Dur > m {
			m = r.Dur
		}
	}

	return m
}

func avgDur(res []*dnsbench.DNSReport) time.Duration {
	if len(res) == 0 {
		return 0
	}

	var t int
	for _, r := range res {
		t += int(r.Dur)
	}
	return time.Duration(t / len(res))
}

func medianDur(res []*dnsbench.DNSReport) time.Duration {
	if len(res) == 0 {
		return 0
	}

	t := sorted(res)
	return t[len(t)/2]
}

func topN(res []*dnsbench.DNSReport, n int) time.Duration {
	if len(res) == 0 {
		return 0
	}

	var dur []int
	for _, r := range res {
		dur = append(dur, int(r.Dur))
	}
	sort.Ints(dur)

	idx := len(dur)*n/100 - 1
	return time.Duration(dur[idx])
}

func sorted(res []*dnsbench.DNSReport) []time.Duration {
	var dur []int
	for _, r := range res {
		dur = append(dur, int(r.Dur))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(dur)))

	r := make([]time.Duration, len(dur))
	for i, d := range dur {
		r[i] = time.Duration(d)
	}
	return r
}

func ips(res []*dnsbench.DNSReport) string {
	var str []string
	de := make(map[string]bool)
	for _, r := range res {
		for _, cr := range r.ConnReps {
			s := cr.IP.String()
			if _, ok := de[s]; !ok {
				str = append(str, s)
				de[s] = true
			}

		}
	}
	return strings.Join(str, " ")
}

func numErr(res []*dnsbench.DNSReport) int {
	n := 0
	for _, r := range res {
		if r.Err != nil {
			n += 1
		}
	}
	return n
}
