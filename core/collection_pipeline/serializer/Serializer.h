/*
 * Copyright 2024 iLogtail Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#pragma once

#include <chrono>
#include <string>

#include "collection_pipeline/batch/BatchedEvents.h"
#include "collection_pipeline/plugin/interface/Flusher.h"
#include "models/PipelineEventPtr.h"
#include "monitor/metric_constants/MetricConstants.h"

namespace logtail {

inline size_t GetInputSize(const PipelineEventPtr& p) {
    return p->DataSize();
}

inline size_t GetInputSize(const BatchedEvents& p) {
    return p.mSizeBytes;
}

inline size_t GetInputSize(const BatchedEventsList& p) {
    size_t size = 0;
    for (const auto& e : p) {
        size += GetInputSize(e);
    }
    return size;
}

// T: PipelineEventPtr, BatchedEvents, BatchedEventsList
template <typename T>
class Serializer {
public:
    Serializer(Flusher* f) : mFlusher(f) {
        WriteMetrics::GetInstance()->CreateMetricsRecordRef(
            mMetricsRecordRef,
            MetricCategory::METRIC_CATEGORY_COMPONENT,
            {{METRIC_LABEL_KEY_PROJECT, f->GetContext().GetProjectName()},
             {METRIC_LABEL_KEY_PIPELINE_NAME, f->GetContext().GetConfigName()},
             {METRIC_LABEL_KEY_COMPONENT_NAME, METRIC_LABEL_VALUE_COMPONENT_NAME_SERIALIZER},
             {METRIC_LABEL_KEY_FLUSHER_PLUGIN_ID, f->GetPluginID()}});
        mInItemsTotal = mMetricsRecordRef.CreateCounter(METRIC_COMPONENT_IN_ITEMS_TOTAL);
        mInItemSizeBytes = mMetricsRecordRef.CreateCounter(METRIC_COMPONENT_IN_SIZE_BYTES);
        mOutItemsTotal = mMetricsRecordRef.CreateCounter(METRIC_COMPONENT_OUT_ITEMS_TOTAL);
        mOutItemSizeBytes = mMetricsRecordRef.CreateCounter(METRIC_COMPONENT_OUT_SIZE_BYTES);
        mTotalProcessMs = mMetricsRecordRef.CreateTimeCounter(METRIC_COMPONENT_TOTAL_PROCESS_TIME_MS);
        mDiscardedItemsTotal = mMetricsRecordRef.CreateCounter(METRIC_COMPONENT_DISCARDED_ITEMS_TOTAL);
        mDiscardedItemSizeBytes = mMetricsRecordRef.CreateCounter(METRIC_COMPONENT_DISCARDED_SIZE_BYTES);
        WriteMetrics::GetInstance()->CommitMetricsRecordRef(mMetricsRecordRef);
    }
    virtual ~Serializer() = default;

    bool DoSerialize(T&& p, std::string& output, std::string& errorMsg) {
        auto inputSize = GetInputSize(p);
        ADD_COUNTER(mInItemsTotal, 1);
        ADD_COUNTER(mInItemSizeBytes, inputSize);

        auto before = std::chrono::system_clock::now();
        auto res = Serialize(std::move(p), output, errorMsg);
        ADD_COUNTER(mTotalProcessMs, std::chrono::system_clock::now() - before);

        if (res) {
            ADD_COUNTER(mOutItemsTotal, 1);
            ADD_COUNTER(mOutItemSizeBytes, output.size());
        } else {
            ADD_COUNTER(mDiscardedItemsTotal, 1);
            ADD_COUNTER(mDiscardedItemSizeBytes, inputSize);
        }
        return res;
    }

protected:
    // if serialized output contains output related info, it can be obtained via this member
    const Flusher* mFlusher = nullptr;

    mutable MetricsRecordRef mMetricsRecordRef;
    CounterPtr mInItemsTotal;
    CounterPtr mInItemSizeBytes;
    CounterPtr mOutItemsTotal;
    CounterPtr mOutItemSizeBytes;
    CounterPtr mDiscardedItemsTotal;
    CounterPtr mDiscardedItemSizeBytes;
    TimeCounterPtr mTotalProcessMs;

private:
    virtual bool Serialize(T&& p, std::string& res, std::string& errorMsg) = 0;

#ifdef APSARA_UNIT_TEST_MAIN
    friend class SerializerUnittest;
#endif
};

using EventSerializer = Serializer<PipelineEventPtr>;
using EventGroupSerializer = Serializer<BatchedEvents>;
using EventGroupListSerializer = Serializer<BatchedEventsList>;

} // namespace logtail
