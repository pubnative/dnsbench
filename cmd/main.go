package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
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
	var res []dnsbench.Report
	for i := 0; i < *loop; i++ {
		res = append(res, <-resCh)
	}

	fmt.Println("DNS Resolve")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "tests \t errors \t min \t max \t avg \t median \t top90 \t top95 \t top99 \t")
	fmt.Fprintf(
		w,
		"%v \t %v \t %v \t %v \t %v \t %v \t %v \t %v \t %v \t\n",
		len(res), numErr(res), minDur(res), maxDur(res), avgDur(res),
		medianDur(res), topN(res, 90), topN(res, 95), topN(res, 99),
	)
	w.Flush()

	fmt.Println("\nEstablish Connection")
	ipReps := make(map[string][]dnsbench.Report)
	for _, r := range res {
		for _, rep := range r.(*dnsbench.DNSReport).ConnReps {
			ip := rep.IP.String()
			reps, ok := ipReps[ip]
			if !ok {
				reps = make([]dnsbench.Report, 0)
			}
			ipReps[ip] = append(reps, rep)
		}
	}
	for ip, reps := range ipReps {
		fmt.Fprintln(w, "ip \t tests \t errors \t min \t max \t avg \t median \t top90 \t top95 \t top99 \t")
		fmt.Fprintf(
			w,
			"%v \t %v \t %v \t %v \t %v \t %v \t %v \t %v \t %v \t %v \t\n\n",
			ip, len(reps), numErr(reps), minDur(reps), maxDur(reps), avgDur(reps),
			medianDur(reps), topN(reps, 90), topN(reps, 95), topN(reps, 99),
		)
		w.Flush()
	}
}

func minDur(res []dnsbench.Report) time.Duration {
	if len(res) == 0 {
		return 0
	}

	m := res[0].Dur()
	for _, r := range res[1:] {
		if r.Dur() < m {
			m = r.Dur()
		}
	}

	return m
}

func maxDur(res []dnsbench.Report) time.Duration {
	if len(res) == 0 {
		return 0
	}

	m := res[0].Dur()
	for _, r := range res[1:] {
		if r.Dur() > m {
			m = r.Dur()
		}
	}

	return m
}

func avgDur(res []dnsbench.Report) time.Duration {
	if len(res) == 0 {
		return 0
	}

	var t int
	for _, r := range res {
		t += int(r.Dur())
	}
	return time.Duration(t / len(res))
}

func medianDur(res []dnsbench.Report) time.Duration {
	if len(res) == 0 {
		return 0
	}

	t := sorted(res)
	return t[len(t)/2]
}

func topN(res []dnsbench.Report, n int) time.Duration {
	if len(res) == 0 {
		return 0
	}

	var dur []int
	for _, r := range res {
		dur = append(dur, int(r.Dur()))
	}
	sort.Ints(dur)

	idx := len(dur)*n/100 - 1
	return time.Duration(dur[idx])
}

func sorted(res []dnsbench.Report) []time.Duration {
	var dur []int
	for _, r := range res {
		dur = append(dur, int(r.Dur()))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(dur)))

	r := make([]time.Duration, len(dur))
	for i, d := range dur {
		r[i] = time.Duration(d)
	}
	return r
}

func numErr(res []dnsbench.Report) int {
	n := 0
	for _, r := range res {
		if r.Err() != nil {
			n += 1
		}
	}
	return n
}
