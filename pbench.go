package pbench

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/bmizerany/perks/quantile"
)

type B struct {
	*sync.Mutex
	*testing.B

	percs []float64

	stream *quantile.Stream
}

func ReportPercentiles(b *testing.B, percs ...float64) *B {
	return &B{
		Mutex:  new(sync.Mutex),
		B:      b,
		percs:  percs,
		stream: quantile.NewTargeted(percs...),
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

	name := reflect.ValueOf(b.B).Elem().FieldByName("name").String()
	maxLen := reflect.ValueOf(b.B).Elem().FieldByName("context").Elem().FieldByName("maxLen").Int()
	n := int(reflect.ValueOf(b.B).Elem().FieldByName("result").FieldByName("N").Int())

	for _, perc := range b.percs {
		result := &testing.BenchmarkResult{
			N: n,
			T: time.Duration(b.stream.Query(perc)) * time.Duration(b.stream.Count()),
		}

		benchName := fmt.Sprintf("%s/P%02.5g-%d", name, perc*100, runtime.GOMAXPROCS(-1))
		fmt.Printf("%-*s\t%s\n", maxLen, benchName, result)
	}
}

func (b *B) RunParallel(body func(*PB)) {
	b.B.RunParallel(func(pb *testing.PB) {
		body(&PB{
			PB: pb,
			B:  b,
		})
	})
}

type PB struct {
	*testing.PB
	*B
}

func (pb *PB) Next() bool {
	defer func(s time.Time) { pb.record(time.Since(s)) }(time.Now())

	return pb.PB.Next()
}

func (pb *PB) record(d time.Duration) {
	pb.Lock()
	defer pb.Unlock()

	pb.stream.Insert(float64(d))
}
