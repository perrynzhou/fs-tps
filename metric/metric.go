package metric

import (
	"sync/atomic"
	"time"
)

type Metric struct {
	Name         string
	FileCount    uint64
	FileLength   uint64
	Milliseconds uint64
	Start        time.Time
	End          time.Time
}

func NewMetric(name string, start, end time.Time) *Metric {
	return &Metric{
		Name:  name,
		Start: start,
		End:   end,
	}
}
func (m *Metric) Compute(count, length, consumed uint64) {
	atomic.AddUint64(&m.FileCount, count)
	atomic.AddUint64(&m.Milliseconds, consumed)
	atomic.AddUint64(&m.FileLength, length)
}
