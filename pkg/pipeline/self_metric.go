// Copyright 2021 iLogtail Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pipeline

import (
	"github.com/alibaba/ilogtail/pkg/protocol"
)

type Label struct {
	Name  string
	Value string
}

type MetricValue struct {
	Name  string
	Value float64
}

type CounterMetric interface {
	Metric

	Add(v int64)

	// Clear same with set
	Clear(v int64)

	Get() int64
}

type StringMetric interface {
	Metric

	Set(v string)

	Get() string
}

type LatencyMetric interface {
	Metric

	Begin() // Deprecated
	Clear() // Deprecated
	End()   // Deprecated

	// nano second
	Get() int64
}

type MetricCollector interface {
	Collect() []Metric
}

type MetricSet interface {
	Name() string
	ConstLabels() []Label
	LabelNames() []string
}

type MetricVector interface {
	WithLabels([]Label) Metric
}

type Metric interface {
	Name() string
	Serialize(log *protocol.Log)
}

type Counter interface {
	Metric
	Add(float64)
	Get() MetricValue
	Clear()
}

type Gauge interface {
	Metric
	Set(float64)
	Get() MetricValue
	Clear()
}

type Summary interface {
	Metric

	Observe(float64)
	Get() MetricValue
}

type MetricStr interface {
	Metric

	Set(v string)

	Get() string
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
