// Package pbench reports percentiles for parallel benchmarks.
package pbench

import (
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/gavv/monotime"
)

// B wraps a testing.B and adds percentiles.
type B struct {
	sync.Mutex
	*testing.B

	percs []float64

	pbs []*PB

	subBs map[int]*B
	cpus  []int
}

// New initializes a B from a wrapped testing.B.
func New(b *testing.B) *B {
	return &B{
		B:     b,
		percs: []float64{},
		pbs:   []*PB{},
		subBs: map[int]*B{},
		cpus:  []int{},
	}
}

// ReportPercentile records and reports a percentile in sub-benchmark results.
func (b *B) ReportPercentile(perc float64) {
	b.percs = append(b.percs, perc)
}

// Run benchmarks f as a subbenchmark with the given name.
func (b *B) Run(name string, f func(b *B)) bool {
	defer b.report()

	return b.B.Run(name, func(tb *testing.B) {
		subB, cpus := &B{B: tb, percs: b.percs}, runtime.GOMAXPROCS(-1)
		b.subBs[cpus] = subB
		b.cpus = append(b.cpus, cpus)
		f(subB)
	})
}

func (b *B) report() {
	b.Lock()
	defer b.Unlock()

	cpus := b.cpus[1:]
	for _, perc := range b.percs {
		reported := map[int]struct{}{}
		for _, i := range cpus {
			if _, ok := reported[i]; !ok {
				b.subBs[i].reportSub(b, perc, i)
			}
			reported[i] = struct{}{}
		}
	}
}

func (b *B) reportSub(parent *B, perc float64, cpus int) {
	var durations durationSlice
	for _, pb := range b.pbs {
		durations = append(durations, pb.s[:pb.idx]...)
	}
	sort.Sort(durations)

	v := reflect.ValueOf(b.B).Elem()
	name := v.FieldByName("name").String()
	n := int(v.FieldByName("result").FieldByName("N").Int())

	ctx := v.FieldByName("context")
	if ctx.IsNil() {
		ctx = reflect.ValueOf(parent.B).Elem().FieldByName("context")
	}
	maxLen := ctx.Elem().FieldByName("maxLen").Int()

	idx := int(float64(len(durations)) * perc)
	pvalue := time.Duration(durations[idx])
	result := &testing.BenchmarkResult{
		N: n,
		T: pvalue * time.Duration(n),
	}

	var cpuList string
	if cpus > 1 {
		cpuList = fmt.Sprintf("-%d", cpus)
	}

	benchName := fmt.Sprintf("%s/P%02.5g%s", name, perc*100, cpuList)
	fmt.Printf("%-*s\t%s\n", maxLen, benchName, result)
}

// RunParallel runs a benchmark in parallel.
func (b *B) RunParallel(body func(*PB)) {
	b.B.RunParallel(func(pb *testing.PB) {
		body(b.pb(pb))
	})
}

func (b *B) pb(inner *testing.PB) *PB {
	pb := &PB{
		PB: inner,
		s:  make(durationSlice, b.N),
	}

	b.Lock()
	defer b.Unlock()
	b.pbs = append(b.pbs, pb)
	return pb
}

// A PB is used by RunParallel for running parallel benchmarks.
type PB struct {
	*testing.PB

	s    durationSlice
	tick time.Duration
	idx  int
}

// Next reports whether there are more iterations to execute.
func (pb *PB) Next() bool {
	if pb.PB.Next() {
		pb.record()
		return true
	}
	return false
}

func (pb *PB) record() {
	if pb.tick == 0 {
		pb.tick = monotime.Now()
		return
	}

	now := monotime.Now()
	pb.s[pb.idx] = now - pb.tick
	pb.idx++
	pb.tick = now
}

type durationSlice []time.Duration

func (p durationSlice) Len() int           { return len(p) }
func (p durationSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p durationSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
