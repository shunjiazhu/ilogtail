package helper

import (
	"strconv"

	"github.com/alibaba/ilogtail/pkg/pipeline"
	"github.com/alibaba/ilogtail/pkg/protocol"
)

var _ pipeline.Counter = (*CounterImp)(nil)

type CounterImp struct {
	value float64
	prev  float64

	index *[]string
	ms    pipeline.MetricSet
}

func (c *CounterImp) Add(delta float64) {
	AtomicAddFloat64(&c.value, delta)
}

func (c *CounterImp) Get() pipeline.MetricValue {
	value := AtomicLoadFloat64(&c.value)
	AtomicStoreFloat64(&c.prev, value)
	return pipeline.MetricValue{Name: c.ms.Name(), Value: value}
}

func (c *CounterImp) Clear() {
	AtomicStoreFloat64(&c.value, 0)
	AtomicStoreFloat64(&c.prev, 0)
}

func (c *CounterImp) Name() string {
	return c.ms.Name()
}

func (c *CounterImp) Serialize(log *protocol.Log) {
	value := c.Get()
	v := strconv.FormatFloat(value.Value, 'f', 4, 64)
	log.Contents = append(log.Contents,
		&protocol.Log_Content{Key: value.Name, Value: v},
		&protocol.Log_Content{Key: SelfMetricNameKey, Value: c.Name()})

	for _, v := range c.ms.ConstLabels() {
		log.Contents = append(log.Contents, &protocol.Log_Content{Key: v.Name, Value: v.Value})
	}

	labelNames := c.ms.LabelNames()
	for i, v := range *c.index {
		log.Contents = append(log.Contents, &protocol.Log_Content{Key: labelNames[i], Value: v})
	}
}

// type MetricV2[T float64 | string] interface {
// 	Name() string
// 	Serialize(log *protocol.Log)
// 	Merge(t T)
// 	Get() T
// }

// var _ MetricV2[float64] = (*MetricV2Counter)(nil)

// type MetricV2Counter struct {
// }

// func (c *MetricV2Counter) Name() string {
// 	return ""
// }

// func (c *MetricV2Counter) Serialize(log *protocol.Log) {
// }

// func (c *MetricV2Counter) Merge(t float64) {
// }

// func (c *MetricV2Counter) Get() float64 {
// 	return 0
// }

// var _ MetricV2[string] = (*MetricV2String)(nil)

// type MetricV2String struct{}

// func (s *MetricV2String) Name() string {
// 	return ""
// }

// func (s *MetricV2String) Serialize(log *protocol.Log) {
// }

// func (s *MetricV2String) Merge(t string) {
// }

// func (s *MetricV2String) Get() string {
// 	return ""
// }

// type MetricVectorV2[T float64 | string] interface {
// 	WithLabels(labels map[string]string) MetricV2[T]
// }

// var _ MetricVectorV2[float64] = (*MetricVectorV2Impl[float64])(nil)
// var _ MetricVectorV2[string] = (*MetricVectorV2Impl[string])(nil)

// type MetricVectorV2Impl[T float64 | string] struct {
// 	metricType int
// }

// func NewMetricVectorV2[T float64 | string](metricType int) *MetricVectorV2Impl[T] {
// 	return &MetricVectorV2Impl[T]{
// 		metricType: metricType,
// 	}
// }

// func (v *MetricVectorV2Impl[T]) WithLabels(labels map[string]string) MetricV2[T] {
// 	// switch v.metricType {
// 	// case 0:
// 	// 	return &MetricV2Counter{}
// 	// case 1:
// 	// 	return &MetricV2String{}
// 	// }
// 	return nil
// }
