# pbench [![GoDoc](https://godoc.org/github.com/benburkert/pbench?status.svg)](https://godoc.org/github.com/benburkert/pbench)

Percentiles for benchmarks.

``` shell
$ benchstat <(for i in $(seq 1 5) ; do go test -run=NONE -bench=. -benchtime=10s -benchmem -cpu=1,2,4,8 github.com/benburkert/pbench ; done)
name                time/op
Test/Example          246µs ± 1%
Test/Example-2        114µs ± 2%
Test/Example-4       53.0µs ± 1%
Test/Example-8       25.5µs ± 1%
Test/Example/P50     92.6µs ±20%
Test/Example/P50-2    102µs ± 3%
Test/Example/P50-4   93.4µs ±17%
Test/Example/P50-8    104µs ± 0%
Test/Example/P95      568µs ±99%
Test/Example/P95-2    706µs ±65%
Test/Example/P95-4   1.01ms ± 0%
Test/Example/P95-8   502µs ±100%
Test/Example/P99    4.92ms ±109%
Test/Example/P99-2  4.84ms ±107%
Test/Example/P99-4  4.79ms ±109%
Test/Example/P99-8   6.46ms ±82%
Control/Example       250µs ± 6%
Control/Example-2     113µs ± 1%
Control/Example-4    53.1µs ± 2%
Control/Example-8    25.5µs ± 2%

name                alloc/op
Test/Example          72.0B ± 0%
Test/Example-2        80.0B ± 0%
Test/Example-4        96.0B ± 0%
Test/Example-8         128B ± 0%
Control/Example       64.0B ± 0%
Control/Example-2     64.0B ± 0%
Control/Example-4     64.0B ± 0%
Control/Example-8     64.0B ± 0%

name                allocs/op
Test/Example           1.00 ± 0%
Test/Example-2         1.00 ± 0%
Test/Example-4         1.00 ± 0%
Test/Example-8         1.00 ± 0%
Control/Example        1.00 ± 0%
Control/Example-2      1.00 ± 0%
Control/Example-4      1.00 ± 0%
Control/Example-8      1.00 ± 0%
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

