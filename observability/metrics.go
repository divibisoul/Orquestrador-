package observability

import "sync/atomic"
type Counter struct{value uint64}
func(c *Counter)Inc(){atomic.AddUint64(&c.value,1)}
func(c *Counter)Value()uint64{return atomic.LoadUint64(&c.value)}
type Metrics struct{Tasks,Failures Counter}
func(m *Metrics)RecordTask(failed bool){m.Tasks.Inc();if failed{m.Failures.Inc()}}
