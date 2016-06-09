package pbench

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/aristanetworks/goarista/atime"
	"github.com/bmizerany/perks/quantile"
)

type B struct {
	sync.Mutex
	*testing.B

	percs []float64
	pbs   []*PB
}

func ReportPercentiles(b *testing.B, percs ...float64) *B {
	return &B{
		B:     b,
		percs: percs,
		pbs:   []*PB{},
	}
}

func (b *B) Run(name string, f func(b *B)) bool {
	innerB := ReportPercentiles(nil, b.percs...)
	defer innerB.report()

	return b.B.Run(name, func(tb *testing.B) {
		innerB.B = tb

		f(innerB)
	})
}

func (b *B) report() {
	b.Lock()
	defer b.Unlock()

	stream := quantile.NewTargeted(b.percs...)
	for _, pb := range b.pbs {
		for _, d := range pb.s[:pb.idx] {
			stream.Insert(float64(d))
		}
	}

	v := reflect.ValueOf(b.B).Elem()
	name := v.FieldByName("name").String()
	maxLen := v.FieldByName("context").Elem().FieldByName("maxLen").Int()
	n := int(v.FieldByName("result").FieldByName("N").Int())

	for _, perc := range b.percs {
		result := &testing.BenchmarkResult{
			N: n,
			T: time.Duration(stream.Query(perc)) * time.Duration(n),
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
