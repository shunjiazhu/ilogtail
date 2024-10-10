// Copyright 2023 iLogtail Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      grpc://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opentelemetry

import (
	"context"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"

	"github.com/alibaba/ilogtail/pkg/pipeline"
	"github.com/alibaba/ilogtail/pkg/protocol/decoder/opentelemetry"
)

var _ ptraceotlp.GRPCServer = (*tracesGRPCServer)(nil)
var _ pmetricotlp.GRPCServer = (*metricsGRPCServer)(nil)
var _ plogotlp.GRPCServer = (*logsGRPCServer)(nil)

// The following functions implement Opten telemetry GRPCServer Interface.
type tracesGRPCServer struct {
	consumer.ConsumeTracesFunc
	ptraceotlp.UnimplementedGRPCServer
}

type metricsGRPCServer struct {
	consumer.ConsumeMetricsFunc
	pmetricotlp.UnimplementedGRPCServer
}

type logsGRPCServer struct {
	consumer.ConsumeLogsFunc
	plogotlp.UnimplementedGRPCServer
}

func newTracesReceiverV1(c pipeline.Collector) *tracesGRPCServer {
	return &tracesGRPCServer{
		ConsumeTracesFunc: func(ctx context.Context, td ptrace.Traces) error {
			logs, err := opentelemetry.ConvertOtlpTraceV1(td)
			if err != nil {
				return err
			}
			for _, log := range logs {
				c.AddRawLog(log)
			}
			return nil
		},
	}

}

func newTracesReceiver(pctx pipeline.PipelineContext) *tracesGRPCServer {
	return &tracesGRPCServer{
		ConsumeTracesFunc: func(ctx context.Context, td ptrace.Traces) error {
			groupEvents, err := opentelemetry.ConvertOtlpTracesToGroupEvents(td)
			if err != nil {
				return err
			}
			pctx.Collector().CollectList(groupEvents...)
			return nil
		},
	}
}

func newMetricsReceiverV1(c pipeline.Collector) *metricsGRPCServer {
	return &metricsGRPCServer{
		ConsumeMetricsFunc: func(ctx context.Context, md pmetric.Metrics) error {
			logs, err := opentelemetry.ConvertOtlpMetricV1(md)
			if err != nil {
				return err
			}
			for _, log := range logs {
				c.AddRawLog(log)
			}
			return nil
		},
	}
}

func newMetricsReceiver(pctx pipeline.PipelineContext) *metricsGRPCServer {
	return &metricsGRPCServer{
		ConsumeMetricsFunc: func(ctx context.Context, md pmetric.Metrics) error {
			groupEvents, err := opentelemetry.ConvertOtlpMetricsToGroupEvents(md)
			if err != nil {
				return err
			}
			pctx.Collector().CollectList(groupEvents...)
			return nil
		},
	}
}

func newLogsReceiverV1(c pipeline.Collector) *logsGRPCServer {
	return &logsGRPCServer{
		ConsumeLogsFunc: func(ctx context.Context, ld plog.Logs) error {
			logs, err := opentelemetry.ConvertOtlpLogV1(ld)
			if err != nil {
				return err
			}
			for _, log := range logs {
				c.AddRawLog(log)
			}
			return nil
		},
	}
}

func newLogsReceiver(pctx pipeline.PipelineContext) *logsGRPCServer {
	return &logsGRPCServer{
		ConsumeLogsFunc: func(ctx context.Context, ld plog.Logs) error {
			groupEvents, err := opentelemetry.ConvertOtlpLogsToGroupEvents(ld)
			if err != nil {
				return err
			}
			pctx.Collector().CollectList(groupEvents...)
			return nil
		},
	}
}

func (f tracesGRPCServer) Export(ctx context.Context, req ptraceotlp.ExportRequest) (ptraceotlp.ExportResponse, error) {
	td := req.Traces()
	numSpans := td.SpanCount()
	if numSpans == 0 {
		return ptraceotlp.NewExportResponse(), nil
	}

	err := f.ConsumeTracesFunc(ctx, td)
	return ptraceotlp.NewExportResponse(), err
}

func (f metricsGRPCServer) Export(ctx context.Context, req pmetricotlp.ExportRequest) (pmetricotlp.ExportResponse, error) {
	md := req.Metrics()
	numMetrics := md.MetricCount()
	if numMetrics == 0 {
		return pmetricotlp.NewExportResponse(), nil
	}
	err := f.ConsumeMetricsFunc(ctx, md)
	return pmetricotlp.NewExportResponse(), err
}

func (f logsGRPCServer) Export(ctx context.Context, req plogotlp.ExportRequest) (plogotlp.ExportResponse, error) {
	ld := req.Logs()
	numLogs := ld.LogRecordCount()
	if numLogs == 0 {
		return plogotlp.NewExportResponse(), nil
	}
	err := f.ConsumeLogsFunc(ctx, ld)
	return plogotlp.NewExportResponse(), err
}
