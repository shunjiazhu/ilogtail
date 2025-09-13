package opentelemetry

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/alibaba/ilogtail/pkg/protocol/decoder/common"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
)

func genGaugeBytes(count int) []byte {
	gaugeData := genMetricsGaugeOTLP(count)
	req := pmetricotlp.NewExportRequestFromMetrics(gaugeData)
	bytes, _ := req.MarshalProto()
	return bytes
}

func genSumBytes(count int) []byte {
	sumData := genMetricsSumOTLP(count)
	req := pmetricotlp.NewExportRequestFromMetrics(sumData)
	bytes, _ := req.MarshalProto()
	return bytes
}

func genHistogramBytes(count int) []byte {
	histogramData := genMetricsHistogramOTLP(count)
	req := pmetricotlp.NewExportRequestFromMetrics(histogramData)
	bytes, _ := req.MarshalProto()
	return bytes
}

func genSummaryBytes(count int) []byte {
	summaryData := genMetricsSummaryOTLP(count)
	req := pmetricotlp.NewExportRequestFromMetrics(summaryData)
	bytes, _ := req.MarshalProto()
	return bytes
}

func genExponentialHistogramBytes(count int) []byte {
	exponentialHistogramData := genMetricsExponentialHistogramOTLP(count)
	req := pmetricotlp.NewExportRequestFromMetrics(exponentialHistogramData)
	bytes, _ := req.MarshalProto()
	return bytes
}

func genLogsOTLPEventsBytes(count int) []byte {
	logs := genLogsOTLPEvents(count)
	req := plogotlp.NewExportRequestFromLogs(logs)
	bytes, _ := req.MarshalProto()
	return bytes
}

func genTraceBytes(count int) []byte {
	traceData := genTracesOTLP(count)
	req := ptraceotlp.NewExportRequestFromTraces(traceData)
	bytes, _ := req.MarshalProto()
	return bytes
}

func BenchmarkDecodeMetrics_Gauge(b *testing.B) {
	eventCount := 500
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080", nil)
	req.Header.Set("Content-Type", pbContentType)

	decoder := &Decoder{Format: common.ProtocolOTLPMetricV1}

	testcases := []struct {
		name string
		data []byte
	}{
		{
			name: "gauge",
			data: genGaugeBytes(eventCount),
		},
	}

	for _, tc := range testcases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := decoder.DecodeV2(tc.data, req)
				assert.NoError(b, err)
			}
		})
	}
}

func BenchmarkDecodeMetrics_Sum(b *testing.B) {
	eventCount := 500
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080", nil)
	req.Header.Set("Content-Type", pbContentType)

	decoder := &Decoder{Format: common.ProtocolOTLPMetricV1}
	testcases := []struct {
		name string
		data []byte
	}{
		{
			name: "sum",
			data: genSumBytes(eventCount),
		},
	}
	for _, tc := range testcases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := decoder.DecodeV2(tc.data, req)
				assert.NoError(b, err)
			}
		})
	}
}

func BenchmarkDecodeMetrics_Histogram(b *testing.B) {
	eventCount := 500
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080", nil)
	req.Header.Set("Content-Type", pbContentType)

	decoder := &Decoder{Format: common.ProtocolOTLPMetricV1}
	testcases := []struct {
		name string
		data []byte
	}{
		{
			name: "histogram",
			data: genHistogramBytes(eventCount),
		},
	}
	for _, tc := range testcases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := decoder.DecodeV2(tc.data, req)
				assert.NoError(b, err)
			}
		})
	}
}

func BenchmarkDecodeMetrics_Summary(b *testing.B) {
	eventCount := 500
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080", nil)
	req.Header.Set("Content-Type", pbContentType)

	decoder := &Decoder{Format: common.ProtocolOTLPMetricV1}
	testcases := []struct {
		name string
		data []byte
	}{
		{
			name: "summary",
			data: genSummaryBytes(eventCount),
		},
	}
	for _, tc := range testcases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := decoder.DecodeV2(tc.data, req)
				assert.NoError(b, err)
			}
		})
	}
}

func BenchmarkDecodeMetrics_ExponentialHistogram(b *testing.B) {
	eventCount := 500
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080", nil)
	req.Header.Set("Content-Type", pbContentType)

	decoder := &Decoder{Format: common.ProtocolOTLPMetricV1}
	testcases := []struct {
		name string
		data []byte
	}{
		{
			name: "exponential_histogram",
			data: genExponentialHistogramBytes(eventCount),
		},
	}
	for _, tc := range testcases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := decoder.DecodeV2(tc.data, req)
				assert.NoError(b, err)
			}
		})
	}
}

func BenchmarkDecodeLogs(b *testing.B) {
	eventCount := 1000
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080", nil)
	req.Header.Set("Content-Type", pbContentType)

	decoder := &Decoder{Format: common.ProtocolOTLPLogV1}
	data := genLogsOTLPEventsBytes(eventCount)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decoder.DecodeV2(data, req)
		assert.NoError(b, err)
	}
}

func BenchmarkDecodeTraces(b *testing.B) {
	eventCount := 1000
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080", nil)
	req.Header.Set("Content-Type", pbContentType)

	decoder := &Decoder{Format: common.ProtocolOTLPTraceV1}
	data := genTraceBytes(eventCount)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decoder.DecodeV2(data, req)
		assert.NoError(b, err)
	}
}

func genLogsOTLPEvents(count int) plog.Logs {
	ld := plog.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr("host.name", "testHost")
	rl.Resource().SetDroppedAttributesCount(1)
	rl.SetSchemaUrl("testSchemaURL")
	il := rl.ScopeLogs().AppendEmpty()
	il.Scope().SetName("name")
	il.Scope().SetVersion("")
	il.Scope().SetDroppedAttributesCount(1)
	il.SetSchemaUrl("ScopeLogsSchemaURL")

	for i := 0; i < count; i++ {
		lg := il.LogRecords().AppendEmpty()
		lg.SetSeverityNumber(plog.SeverityNumberError)
		lg.SetSeverityText("Error")
		lg.SetDroppedAttributesCount(1)
		lg.SetFlags(plog.DefaultLogRecordFlags)
		traceID := pcommon.TraceID([16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10})
		spanID := pcommon.SpanID([8]byte{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18})
		lg.SetTraceID(traceID)
		lg.SetSpanID(spanID)
		lg.Body().SetStr("hello world")
		t := pcommon.NewTimestampFromTime(time.Now())
		lg.SetTimestamp(t)
		lg.SetObservedTimestamp(t)
		lg.Attributes().PutStr("sdkVersion", "1.0.1")
	}
	return ld
}

func genMetricsSumOTLP(count int) pmetric.Metrics {
	metric := pmetric.NewMetrics()
	rs := metric.ResourceMetrics().AppendEmpty()
	rs.SetSchemaUrl("schemaURL")
	// Add resource.
	rs.Resource().Attributes().PutStr("host.name", "testHost")
	rs.Resource().Attributes().PutStr("service.name", "testService")
	rs.Resource().SetDroppedAttributesCount(1)
	// Add InstrumentationLibraryMetrics.
	m := rs.ScopeMetrics().AppendEmpty()
	m.Scope().SetName("instrumentation name")
	m.Scope().SetVersion("instrumentation version")
	m.Scope().Attributes().PutStr("instrumentation.attribute", "test")
	m.SetSchemaUrl("schemaURL")
	// Add Metric
	for i := 0; i < count; i++ {
		sumMetric := m.Metrics().AppendEmpty()
		sumMetric.SetName("test sum" + strconv.Itoa(i))
		sumMetric.SetDescription("test sum")
		sumMetric.SetUnit("unit")
		sum := sumMetric.SetEmptySum()
		sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
		sum.SetIsMonotonic(true)
		datapoint := sum.DataPoints().AppendEmpty()
		datapoint.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		datapoint.SetIntValue(100)
		datapoint.Attributes().PutStr("string", "value")
		datapoint.Attributes().PutBool("bool", true)
		datapoint.Attributes().PutInt("int", 1)
		datapoint.Attributes().PutDouble("double", 1.1)
		datapoint.Attributes().PutEmptyBytes("bytes").FromRaw([]byte("foo"))

		exemplar := datapoint.Exemplars().AppendEmpty()
		exemplar.SetDoubleValue(99.3)
		exemplar.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		traceID := pcommon.TraceID([16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10})
		spanID := pcommon.SpanID([8]byte{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18})
		exemplar.SetSpanID(spanID)
		exemplar.SetTraceID(traceID)
		exemplar.FilteredAttributes().PutStr("service.name", "testService")
		datapoint.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	}
	return metric
}

func genMetricsGaugeOTLP(count int) pmetric.Metrics {
	metric := pmetric.NewMetrics()
	rs := metric.ResourceMetrics().AppendEmpty()
	rs.SetSchemaUrl("schemaURL")
	// Add resource.
	rs.Resource().Attributes().PutStr("host.name", "testHost")
	rs.Resource().Attributes().PutStr("service.name", "testService")
	rs.Resource().SetDroppedAttributesCount(1)
	// Add InstrumentationLibraryMetrics.
	m := rs.ScopeMetrics().AppendEmpty()
	m.Scope().SetName("instrumentation name")
	m.Scope().SetVersion("instrumentation version")
	m.SetSchemaUrl("schemaURL")
	// Add Metric
	for i := 0; i < count; i++ {
		gaugeMetric := m.Metrics().AppendEmpty()
		gaugeMetric.SetName("test gauge" + strconv.Itoa(i))
		gaugeMetric.SetDescription("test gauge")
		gaugeMetric.SetUnit("unit")
		gauge := gaugeMetric.SetEmptyGauge()
		datapoint := gauge.DataPoints().AppendEmpty()
		datapoint.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		datapoint.SetDoubleValue(10.2)
		datapoint.Attributes().PutStr("string", "value")
		datapoint.Attributes().PutBool("bool", true)
		datapoint.Attributes().PutInt("int", 1)
		datapoint.Attributes().PutDouble("double", 1.1)
		datapoint.Attributes().PutEmptyBytes("bytes").FromRaw([]byte("foo"))
		exemplar := datapoint.Exemplars().AppendEmpty()
		exemplar.SetDoubleValue(99.3)
		exemplar.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		traceID := pcommon.TraceID([16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10})
		spanID := pcommon.SpanID([8]byte{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18})
		exemplar.SetSpanID(spanID)
		exemplar.SetTraceID(traceID)
		exemplar.FilteredAttributes().PutStr("service.name", "testService")
		datapoint.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	}
	return metric
}

func genMetricsHistogramOTLP(count int) pmetric.Metrics {
	metric := pmetric.NewMetrics()
	rs := metric.ResourceMetrics().AppendEmpty()
	rs.SetSchemaUrl("schemaURL")
	// Add resource.
	rs.Resource().Attributes().PutStr("host.name", "testHost")
	rs.Resource().Attributes().PutStr("service.name", "testService")
	rs.Resource().SetDroppedAttributesCount(1)
	// Add InstrumentationLibraryMetrics.
	m := rs.ScopeMetrics().AppendEmpty()
	m.Scope().SetName("instrumentation name")
	m.Scope().SetVersion("instrumentation version")
	m.SetSchemaUrl("schemaURL")
	// Add Metric
	for i := 0; i < count; i++ {
		histogramMetric := m.Metrics().AppendEmpty()
		histogramMetric.SetName("test Histogram" + strconv.Itoa(i))
		histogramMetric.SetDescription("test Histogram")
		histogramMetric.SetUnit("unit")
		histogram := histogramMetric.SetEmptyHistogram()
		histogram.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
		datapoint := histogram.DataPoints().AppendEmpty()
		datapoint.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		datapoint.Attributes().PutStr("string", "value")
		datapoint.Attributes().PutBool("bool", true)
		datapoint.Attributes().PutInt("int", 1)
		datapoint.Attributes().PutDouble("double", 1.1)
		datapoint.Attributes().PutEmptyBytes("bytes").FromRaw([]byte("foo"))
		datapoint.SetCount(4)
		datapoint.SetSum(345)
		datapoint.BucketCounts().FromRaw([]uint64{1, 1, 2})
		datapoint.ExplicitBounds().FromRaw([]float64{10, 100})
		exemplar := datapoint.Exemplars().AppendEmpty()
		exemplar.SetDoubleValue(99.3)
		exemplar.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		datapoint.SetMin(float64(time.Now().Unix()))
		traceID := pcommon.TraceID([16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10})
		spanID := pcommon.SpanID([8]byte{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18})
		exemplar.SetSpanID(spanID)
		exemplar.SetTraceID(traceID)
		exemplar.FilteredAttributes().PutStr("service.name", "testService")
		datapoint.SetMax(float64(time.Now().Unix()))
		datapoint.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	}
	return metric
}

func genMetricsExponentialHistogramOTLP(count int) pmetric.Metrics {
	metric := pmetric.NewMetrics()
	rs := metric.ResourceMetrics().AppendEmpty()
	rs.SetSchemaUrl("schemaURL")
	// Add resource.
	rs.Resource().Attributes().PutStr("host.name", "testHost")
	rs.Resource().Attributes().PutStr("service.name", "testService")
	rs.Resource().SetDroppedAttributesCount(1)
	// Add InstrumentationLibraryMetrics.
	m := rs.ScopeMetrics().AppendEmpty()
	m.Scope().SetName("instrumentation name")
	m.Scope().SetVersion("instrumentation version")
	m.SetSchemaUrl("schemaURL")
	// Add Metric
	for i := 0; i < count; i++ {
		histogramMetric := m.Metrics().AppendEmpty()
		histogramMetric.SetName("test ExponentialHistogram" + strconv.Itoa(i))
		histogramMetric.SetDescription("test ExponentialHistogram")
		histogramMetric.SetUnit("unit")
		histogram := histogramMetric.SetEmptyExponentialHistogram()
		histogram.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
		datapoint := histogram.DataPoints().AppendEmpty()
		datapoint.SetScale(1)
		datapoint.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		datapoint.Attributes().PutStr("string", "value")
		datapoint.Attributes().PutBool("bool", true)
		datapoint.Attributes().PutInt("int", 1)
		datapoint.Attributes().PutDouble("double", 1.1)
		datapoint.Attributes().PutEmptyBytes("bytes").FromRaw([]byte("foo"))
		datapoint.SetCount(4)
		datapoint.SetSum(345)
		datapoint.Positive().BucketCounts().FromRaw([]uint64{1, 1, 2})
		datapoint.Positive().SetOffset(2)
		datapoint.Negative().BucketCounts().FromRaw([]uint64{1, 1, 2})
		datapoint.Negative().SetOffset(3)

		exemplar := datapoint.Exemplars().AppendEmpty()
		exemplar.SetDoubleValue(99.3)
		exemplar.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		datapoint.SetMin(float64(time.Now().Unix()))
		traceID := pcommon.TraceID([16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10})
		spanID := pcommon.SpanID([8]byte{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18})
		exemplar.SetSpanID(spanID)
		exemplar.SetTraceID(traceID)
		exemplar.FilteredAttributes().PutStr("service.name", "testService")
		datapoint.SetMax(float64(time.Now().Unix()))
		datapoint.Negative().BucketCounts().FromRaw([]uint64{1, 1, 2})
		datapoint.Negative().SetOffset(2)
		datapoint.SetZeroCount(5)
		datapoint.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	}
	return metric
}

func genMetricsSummaryOTLP(count int) pmetric.Metrics {
	metric := pmetric.NewMetrics()
	rs := metric.ResourceMetrics().AppendEmpty()
	rs.SetSchemaUrl("schemaURL")
	// Add resource.
	rs.Resource().Attributes().PutStr("host.name", "testHost")
	rs.Resource().Attributes().PutStr("service.name", "testService")
	rs.Resource().SetDroppedAttributesCount(1)
	// Add InstrumentationLibraryMetrics.
	m := rs.ScopeMetrics().AppendEmpty()
	m.Scope().SetName("instrumentation name")
	m.Scope().SetVersion("instrumentation version")
	m.SetSchemaUrl("schemaURL")
	// Add Metric
	for i := 0; i < count; i++ {
		summaryMetric := m.Metrics().AppendEmpty()
		summaryMetric.SetName("test summary" + strconv.Itoa(i))
		summaryMetric.SetDescription("test summary")
		summaryMetric.SetUnit("unit")
		summary := summaryMetric.SetEmptySummary()
		datapoint := summary.DataPoints().AppendEmpty()

		datapoint.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		datapoint.SetCount(100)
		datapoint.SetSum(1000)
		quantile := datapoint.QuantileValues().AppendEmpty()
		quantile.SetQuantile(0.5)
		quantile.SetValue(1.2)
		datapoint.Attributes().PutStr("string", "value")
		datapoint.Attributes().PutBool("bool", true)
		datapoint.Attributes().PutInt("int", 1)
		datapoint.Attributes().PutDouble("double", 1.1)
		datapoint.Attributes().PutEmptyBytes("bytes").FromRaw([]byte("foo"))
		datapoint.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	}
	return metric
}

func genTracesOTLP(count int) ptrace.Traces {
	traceID := pcommon.TraceID([16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10})
	spanID := pcommon.SpanID([8]byte{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18})
	td := ptrace.NewTraces()
	// Add ResourceSpans.
	rs := td.ResourceSpans().AppendEmpty()
	rs.SetSchemaUrl("schemaURL")
	// Add resource.
	rs.Resource().Attributes().PutStr("host.name", "testHost")
	rs.Resource().Attributes().PutStr("service.name", "testService")
	rs.Resource().SetDroppedAttributesCount(1)
	// Add ScopeSpans.
	il := rs.ScopeSpans().AppendEmpty()
	il.Scope().SetName("scope name")
	il.Scope().SetVersion("scope version")
	il.SetSchemaUrl("schemaURL")
	// Add spans.
	sp := il.Spans().AppendEmpty()
	sp.SetName("testSpan")
	sp.SetKind(ptrace.SpanKindClient)
	sp.SetDroppedAttributesCount(1)
	sp.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	sp.SetTraceID(traceID)
	sp.SetSpanID(spanID)
	sp.SetDroppedEventsCount(1)
	sp.SetDroppedLinksCount(1)
	sp.SetEndTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	sp.SetParentSpanID(spanID)
	sp.TraceState().FromRaw("state")
	sp.Status().SetCode(ptrace.StatusCodeOk)
	sp.Status().SetMessage("message")
	// Add attributes.
	sp.Attributes().PutStr("string", "value")
	sp.Attributes().PutBool("bool", true)
	sp.Attributes().PutInt("int", 1)
	sp.Attributes().PutDouble("double", 1.1)
	sp.Attributes().PutEmptyBytes("bytes").FromRaw([]byte("foo"))
	arr := sp.Attributes().PutEmptySlice("array")
	arr.AppendEmpty().SetInt(1)
	arr.AppendEmpty().SetStr("str")
	kvList := sp.Attributes().PutEmptyMap("kvList")
	kvList.PutInt("int", 1)
	kvList.PutStr("string", "string")
	// Add events.

	for i := 0; i < count; i++ {
		event := sp.Events().AppendEmpty()
		event.SetName("eventName" + strconv.Itoa(i))
		event.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		event.SetDroppedAttributesCount(1)
		event.Attributes().PutStr("string", "value")
		event.Attributes().PutBool("bool", true)
		event.Attributes().PutInt("int", 1)
		event.Attributes().PutDouble("double", 1.1)
		event.Attributes().PutEmptyBytes("bytes").FromRaw([]byte("foo"))

		// Add links.
		link := sp.Links().AppendEmpty()
		link.TraceState().FromRaw("state")
		link.SetTraceID(traceID)
		link.SetSpanID(spanID)
		link.SetDroppedAttributesCount(1)
		link.Attributes().PutStr("string", "value")
		link.Attributes().PutBool("bool", true)
		link.Attributes().PutInt("int", 1)
		link.Attributes().PutDouble("double", 1.1)
		link.Attributes().PutEmptyBytes("bytes").FromRaw([]byte("foo"))
		// Add another span.
		sp2 := il.Spans().AppendEmpty()
		sp2.SetName("testSpan2")
	}

	return td
}
