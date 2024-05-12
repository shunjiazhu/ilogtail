// Copyright 2024 iLogtail Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package helper

import (
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/alibaba/ilogtail/pkg/models"
	"github.com/alibaba/ilogtail/pkg/pipeline"
)

type (
	Label    = pipeline.Label
	TagIndex = *[]string
)

type CounterMetricVector interface {
	WithLabels([]Label) pipeline.Counter
}

type GaugeMetricVector interface {
	WithLabels([]Label) pipeline.Gauge
}

type StringMetricVector interface {
	WithLabels([]Label) pipeline.MetricStr
}

type SummaryMetricVector interface {
	WithLabels([]Label) pipeline.Summary
}

var _ CounterMetricVector = (*CounterVectorImpl)(nil)

type CounterVectorImpl struct {
	MetricMap
}

func NewAverageMetricVector(metricName string, constLabels []Label, labelNames []string) CounterMetricVector {
	return &CounterVectorImpl{
		MetricMap: NewMetricVector(metricName, models.MetricTypeRateCounter, constLabels, labelNames),
	}
}

func (v *CounterVectorImpl) WithLabels(labels []Label) pipeline.Counter {
	return v.MetricMap.WithLabels(labels).(pipeline.Counter)
}

type MetricMap interface {
	pipeline.MetricCollector
	pipeline.MetricVector
}

var (
	_ pipeline.MetricCollector = (*MetricVectorImpl)(nil)
	_ pipeline.MetricVector    = (*MetricVectorImpl)(nil)
	_ pipeline.MetricSet       = (*MetricVectorImpl)(nil)
)

type MetricVectorImpl struct {
	name        string // metric name
	metricType  models.MetricType
	constLabels []Label  // constLabels is the labels that are not changed when the metric is created.
	labelNames  []string // labelNames is the names of the labels. The values of the labels can be changed.

	indexPool   GenericPool[string]
	bytesPool   GenericPool[byte]
	collector   sync.Map
	seriesCount int64
}

func NewMetricVector(metricName string, metricType models.MetricType, constLabels []Label, labelNames []string) *MetricVectorImpl {
	mv := &MetricVectorImpl{
		name:        metricName,
		metricType:  metricType,
		constLabels: constLabels,
		labelNames:  labelNames,
		indexPool:   NewGenericPool(func() []string { return make([]string, 0, 10) }),
		bytesPool:   NewGenericPool(func() []byte { return make([]byte, 0, 128) }),
	}
	return mv
}

func (v *MetricVectorImpl) Name() string {
	return v.name
}

func (v *MetricVectorImpl) ConstLabels() []Label {
	return v.constLabels
}

func (v *MetricVectorImpl) LabelNames() []string {
	return v.labelNames
}

func (v *MetricVectorImpl) WithLabels(labels []Label) pipeline.Metric {
	index, err := v.buildIndex(labels)
	if err != nil {
		panic(err)
	}

	buffer := v.bytesPool.Get()
	for _, tagValue := range *index {
		*buffer = append(*buffer, '|')
		*buffer = append(*buffer, tagValue...)
	}
	k := *(*string)(unsafe.Pointer(buffer))
	acV, loaded := v.collector.Load(k)
	if loaded {
		metric := acV.(pipeline.Metric)
		v.indexPool.Put(index)
		v.bytesPool.Put(buffer)
		return metric
	}

	var newMetric pipeline.Metric
	switch v.metricType {
	case models.MetricTypeCounter:
		newMetric = NewCounterMetric(v.name)
	case models.MetricTypeRateCounter:
		newMetric = NewAverageMetric(v.name)
	case models.MetricTypeGauge:
	case models.MetricTypeSummary:
	case models.MetricTypeUntyped:
		newMetric = NewStringMetric(v.name)
	}

	acV, loaded = v.collector.LoadOrStore(k, newMetric)
	if !loaded {
		atomic.AddInt64(&v.seriesCount, 1)
	} else {
		v.bytesPool.Put(buffer)
	}
	return acV.(pipeline.Metric)
}

func (v *MetricVectorImpl) Collect() []pipeline.Metric {
	res := make([]pipeline.Metric, 0, 10)
	v.collector.Range(func(key, value interface{}) bool {
		res = append(res, value.(pipeline.Metric))
		return true
	})
	return res
}

func (v *MetricVectorImpl) buildIndex(labels []Label) (*[]string, error) {
	index := v.indexPool.Get()
	for range v.labelNames {
		*index = append(*index, "-")
	}

	for d, tag := range labels {
		if v.labelNames[d] == tag.Name { // fast path
			(*index)[d] = tag.Value
		} else {
			err := v.slowConstructIndex(index, tag)
			if err != nil {
				v.indexPool.Put(index)
				return nil, err
			}
		}
	}

	return index, nil
}

func (v *MetricVectorImpl) slowConstructIndex(index *[]string, tag pipeline.Label) error {
	for i, tagName := range v.labelNames {
		if tagName == tag.Name {
			(*index)[i] = tag.Value
			return nil
		}
	}
	return fmt.Errorf("undefined tag name: %s", tag.Name)
}
