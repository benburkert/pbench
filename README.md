# pbench [![GoDoc](https://godoc.org/github.com/benburkert/pbench?status.svg)](https://godoc.org/github.com/benburkert/pbench)

Percentiles for benchmarks.

``` shell
$ benchstat <({ for i in $(seq 1 5) ; do go test -run=NONE -bench=. -benchtime=10s -benchmem ./... ; done ; })
name                time/op
Test/Example-8      25.5µs ± 0%
Test/Example/P50-8   104µs ± 0%
Test/Example/P95-8   504µs ± 1%
Test/Example/P99-8  1.24ms ± 7%
Control/Example-8   25.5µs ± 1%

name                alloc/op
Test/Example-8        128B ± 0%
Control/Example-8    64.0B ± 0%

name                allocs/op
Test/Example-8        1.00 ± 0%
Control/Example-8     1.00 ± 0%
```

# Example

``` go
func BenchmarkPercentiles(tb *testing.B) {
	b := pbench.New(tb)
	b.ReportPercentile(0.5)
	b.ReportPercentile(0.95)
	b.ReportPercentile(0.99)

	b.Run("Example", func(b *pbench.B) {
		rand.Seed(int64(time.Now().Nanosecond()))

		b.ResetTimer()

		b.RunParallel(func(pb *pbench.PB) {
			for pb.Next() {
				v := rand.Float64()
				switch {
				case v <= 0.5:
					time.Sleep(10 * time.Microsecond)
				case v <= 0.95:
					time.Sleep(100 * time.Microsecond)
				case v <= 0.99:
					time.Sleep(1000 * time.Microsecond)
				default:
					time.Sleep(10000 * time.Microsecond)
				}
			}
		})
	})
}
```

