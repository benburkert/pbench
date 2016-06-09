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

	"github.com/aristanetworks/goarista/atime"
)

type B struct {
	sync.Mutex
	*testing.B

	percs []float64

	pbs []*PB
}

func New(b *testing.B) *B {
	return &B{
		B:     b,
		percs: []float64{},
		pbs:   []*PB{},
	}
}

func (b *B) ReportPercentile(perc float64) {
	b.percs = append(b.percs, perc)
}

func (b *B) Run(name string, f func(b *B)) bool {
	innerB := &B{percs: b.percs}
	defer innerB.report()

	return b.B.Run(name, func(tb *testing.B) {
		innerB.B = tb

		f(innerB)
	})
}

func (b *B) report() {
	b.Lock()
	defer b.Unlock()

	durations := []float64{}
	for _, pb := range b.pbs {
		for _, d := range pb.s[:pb.idx] {
			durations = append(durations, float64(d))
		}
	}
	sort.Float64s(durations)

	v := reflect.ValueOf(b.B).Elem()
	name := v.FieldByName("name").String()
	maxLen := v.FieldByName("context").Elem().FieldByName("maxLen").Int()
	n := int(v.FieldByName("result").FieldByName("N").Int())

	for _, perc := range b.percs {
		idx := int(float64(len(durations)) * perc)
		pvalue := time.Duration(durations[idx])
		result := &testing.BenchmarkResult{
			N: n,
			T: pvalue * time.Duration(n),
		}

		var cpuList string
		if cpus := runtime.GOMAXPROCS(-1); cpus > 1 {
			cpuList = fmt.Sprintf("-%d", cpus)
		}

		benchName := fmt.Sprintf("%s/P%02.5g%s", name, perc*100, cpuList)
		fmt.Printf("%-*s\t%s\n", maxLen, benchName, result)
	}
}

func (b *B) RunParallel(body func(*PB)) {
	b.B.RunParallel(func(pb *testing.PB) {
		body(b.pb(pb))
	})
}

func (b *B) pb(inner *testing.PB) *PB {
	pb := &PB{
		PB: inner,
		s:  make([]uint64, b.N),
	}

	b.Lock()
	defer b.Unlock()
	b.pbs = append(b.pbs, pb)
	return pb
}

type PB struct {
	*testing.PB

	s    []uint64
	tick uint64
	idx  int
}

func (pb *PB) Next() bool {
	if pb.PB.Next() {
		pb.record()
		return true
	}
	return false
}

func (pb *PB) record() {
	if pb.tick == 0 {
		pb.tick = atime.NanoTime()
		return
	}

	now := atime.NanoTime()
	pb.s[pb.idx] = now - pb.tick
	pb.idx++
	pb.tick = now
}
