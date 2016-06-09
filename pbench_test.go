package pbench

import (
	"math/rand"
	"testing"
	"time"
)

func BenchmarkTest(tb *testing.B) {
	b := New(tb)
	b.ReportPercentile(0.5)
	b.ReportPercentile(0.95)
	b.ReportPercentile(0.99)

	b.Run("Example", func(b *B) {
		rand.Seed(int64(time.Now().Nanosecond()))

		b.ResetTimer()

		b.RunParallel(func(pb *PB) {
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

func BenchmarkControl(b *testing.B) {
	b.Run("Example", func(b *testing.B) {
		rand.Seed(int64(time.Now().Nanosecond()))

		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
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
