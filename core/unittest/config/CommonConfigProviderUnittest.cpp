// Copyright 2023 iLogtail Authors
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


#include "json/json.h"

#include "AppConfig.h"
#include "Application.h"
#include "collection_pipeline/CollectionPipelineManager.h"
#include "common/FileSystemUtil.h"
#include "common/version.h"
#include "config/ConfigDiff.h"
#include "config/InstanceConfigManager.h"
#include "config/common_provider/CommonConfigProvider.h"
#include "config/watcher/InstanceConfigWatcher.h"
#include "config/watcher/PipelineConfigWatcher.h"
#include "file_server/FileServer.h"
#include "gmock/gmock.h"
#include "monitor/Monitor.h"
#include "unittest/Unittest.h"
#ifdef __ENTERPRISE__
#include "config/provider/EnterpriseConfigProvider.h"
#endif

DECLARE_FLAG_BOOL(logtail_mode);

using namespace testing;
using namespace std;

namespace logtail {

class MockCommonConfigProvider : public CommonConfigProvider {
public:
    MOCK_METHOD5(SendHttpRequest,
                 bool(const std::string&, const std::string&, const std::string&, const std::string&, std::string&));
};

class CommonConfigProviderUnittest : public ::testing::Test {
public:
    std::string mRootDir;
    const std::string& PS = PATH_SEPARATOR;
    string ilogtailConfigPath;

    bool writeJsonToFile(const std::string& jsonString, const std::string& filePath) {
        Json::Reader reader;
        Json::Value root;

        bool parsingSuccessful = reader.parse(jsonString, root);
        if (!parsingSuccessful) {
            std::cout << "Failed to parse configuration\n" << reader.getFormattedErrorMessages();
            return false;
        }

        std::ofstream fout(filePath.c_str());
        if (!fout) {
            std::cout << "Failed to open file: " << filePath << std::endl;
            return false;
        }

        fout << root.toStyledString() << std::endl;
        fout.close();
        AppConfig::GetInstance()->LoadAppConfig(ilogtailConfigPath);
        return true;
    }

    static void SetUpTestCase() {}

    static void TearDownTestCase() {
        Application::GetInstance()->SetSigTermSignalFlag(true);
        FileServer::GetInstance()->Stop();
    }

    // 在每个测试用例开始前的设置
    void SetUp() override {
        if (BOOL_FLAG(logtail_mode)) {
            mRootDir = GetProcessExecutionDir();
            bfs::create_directories(mRootDir);
            ilogtailConfigPath = mRootDir + PS + STRING_FLAG(ilogtail_config);
            std::ofstream fout(ilogtailConfigPath.c_str());
            fout << "" << std::endl;
            fout.close();
            MockCommonConfigProvider provider;
            provider.Init("common_v2");
            provider.Stop();
            bfs::remove_all(provider.mContinuousPipelineConfigDir.string());
            bfs::remove_all(provider.mInstanceSourceDir.string());
        } else {
            CreateAgentDir();
            ilogtailConfigPath = GetAgentConfDir() + "/instance_config/local/loongcollector_config.json";
            AppConfig::GetInstance()->LoadAppConfig(ilogtailConfigPath);
            std::ofstream fout(ilogtailConfigPath.c_str());
            fout << "" << std::endl;
            fout.close();
            MockCommonConfigProvider provider;
            provider.Init("common_v2");
            provider.Stop();
            bfs::remove_all(provider.mContinuousPipelineConfigDir.string());
            bfs::remove_all(provider.mInstanceSourceDir.string());
        }
    }

    // 在每个测试用例结束后的清理
    void TearDown() override {
        MockCommonConfigProvider provider;
        provider.Init("common_v2");
        provider.Stop();
        bfs::remove_all(provider.mContinuousPipelineConfigDir.string());
        bfs::remove_all(provider.mInstanceSourceDir.string());
    }

    void TestInit();

    void TestGetConfigUpdateAndConfigWatcher();
};

void CommonConfigProviderUnittest::TestInit() {
    {
        string config = R"(
        {
            "config_server_list": [
                {
                    "cluster": "community",
                    "endpoint_list": [
                        "test.config.com:80"
                    ]
                }
            ],
            "ilogtail_tags": {
                "key1": "value1",
                "key2": "value2"
            }
        }
        )";
        APSARA_TEST_TRUE_FATAL(writeJsonToFile(config, ilogtailConfigPath));

        MockCommonConfigProvider provider;
        provider.Init("common_v2");
        APSARA_TEST_EQUAL(provider.mSequenceNum, 0);
        APSARA_TEST_EQUAL(provider.sName, "common config provider");
        APSARA_TEST_EQUAL(provider.mConfigServerAvailable, true);
        APSARA_TEST_EQUAL(provider.mConfigServerAddresses.size(), 1);
        APSARA_TEST_EQUAL(provider.mConfigServerAddresses[0].host, "test.config.com");
        APSARA_TEST_EQUAL(provider.mConfigServerAddresses[0].port, 80);
        APSARA_TEST_EQUAL(provider.mConfigServerTags.size(), 2);
        APSARA_TEST_EQUAL(provider.mConfigServerTags["key1"], "value1");
        APSARA_TEST_EQUAL(provider.mConfigServerTags["key2"], "value2");
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).host, "test.config.com");
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).port, 80);
        APSARA_TEST_EQUAL(provider.mConfigServerAddressId, 0);
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(true).host, "test.config.com");
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(true).port, 80);
        APSARA_TEST_EQUAL(provider.mConfigServerAddressId, 0);
        provider.Stop();
    }
    {
        string config = R"(
        {
            "config_server_list": [
                {
                    "cluster": "community",
                    "endpoint_list": [
                        "test.config.com:80",
                        "test1.config.com:81",
                        "test1.config.com81",
                        "test1.config.com:0",
                        "test1.config.com:65536"
                    ]
                }
            ]
        }
        )";
        APSARA_TEST_TRUE_FATAL(writeJsonToFile(config, ilogtailConfigPath));

        MockCommonConfigProvider provider;
        provider.Init("common_v2");
        APSARA_TEST_EQUAL(provider.mSequenceNum, 0);
        APSARA_TEST_EQUAL(provider.sName, "common config provider");
        APSARA_TEST_EQUAL(provider.mConfigServerAvailable, true);
        APSARA_TEST_EQUAL(provider.mConfigServerAddresses.size(), 2);
        APSARA_TEST_EQUAL(provider.mConfigServerAddresses[0].host, "test.config.com");
        APSARA_TEST_EQUAL(provider.mConfigServerAddresses[0].port, 80);
        APSARA_TEST_EQUAL(provider.mConfigServerAddresses[1].host, "test1.config.com");
        APSARA_TEST_EQUAL(provider.mConfigServerAddresses[1].port, 81);
        APSARA_TEST_EQUAL(provider.mConfigServerTags.size(), 0);
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).host,
                          provider.mConfigServerAddresses[provider.mConfigServerAddressId].host);
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).port,
                          provider.mConfigServerAddresses[provider.mConfigServerAddressId].port);
        APSARA_TEST_EQUAL(provider.mConfigServerAddressId, 0);

        CommonConfigProvider::ConfigServerAddress address = provider.GetOneConfigServerAddress(true);
        APSARA_TEST_EQUAL(address.host, provider.mConfigServerAddresses[provider.mConfigServerAddressId].host);
        APSARA_TEST_EQUAL(address.port, provider.mConfigServerAddresses[provider.mConfigServerAddressId].port);
        APSARA_TEST_EQUAL(provider.mConfigServerAddressId, 1);
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).host,
                          provider.mConfigServerAddresses[provider.mConfigServerAddressId].host);
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).port,
                          provider.mConfigServerAddresses[provider.mConfigServerAddressId].port);
        APSARA_TEST_EQUAL(provider.mConfigServerAddressId, 1);

        address = provider.GetOneConfigServerAddress(true);
        APSARA_TEST_EQUAL(address.host, provider.mConfigServerAddresses[provider.mConfigServerAddressId].host);
        APSARA_TEST_EQUAL(address.port, provider.mConfigServerAddresses[provider.mConfigServerAddressId].port);
        APSARA_TEST_EQUAL(provider.mConfigServerAddressId, 0);
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).host,
                          provider.mConfigServerAddresses[provider.mConfigServerAddressId].host);
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).port,
                          provider.mConfigServerAddresses[provider.mConfigServerAddressId].port);
        APSARA_TEST_EQUAL(provider.mConfigServerAddressId, 0);

        provider.Stop();
    }
    {
        string config = R"(
        {
            "config_server_list1": [
                {
                    "cluster": "community",
                    "endpoint_list": [
                        "test.config.com:80",
                        "test1.config.com:81"
                    ]
                }
            ],
            "config_server_list": [
                {
                    "cluster": "community",
                    "endpoint_list2": [
                        "test.config.com:80",
                        "test1.config.com:81"
                    ],
                    "endpoint_list": "test.config.com:80"
                }
            ]
        }
        )";
        APSARA_TEST_TRUE_FATAL(writeJsonToFile(config, ilogtailConfigPath));

        MockCommonConfigProvider provider;
        provider.Init("common_v2");
        APSARA_TEST_EQUAL(provider.mSequenceNum, 0);
        APSARA_TEST_EQUAL(provider.sName, "common config provider");
        APSARA_TEST_EQUAL(provider.mConfigServerAvailable, false);
        APSARA_TEST_EQUAL(provider.mConfigServerAddresses.size(), 0);
        APSARA_TEST_EQUAL(provider.mConfigServerTags.size(), 0);
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).host, "");
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).port, -1);
        APSARA_TEST_EQUAL(provider.mConfigServerAddressId, 0);

        CommonConfigProvider::ConfigServerAddress address = provider.GetOneConfigServerAddress(true);
        APSARA_TEST_EQUAL(address.host, "");
        APSARA_TEST_EQUAL(address.port, -1);
        APSARA_TEST_EQUAL(provider.mConfigServerAddressId, 0);
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).host, "");
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).port, -1);
        APSARA_TEST_EQUAL(provider.mConfigServerAddressId, 0);

        address = provider.GetOneConfigServerAddress(true);
        APSARA_TEST_EQUAL(address.host, "");
        APSARA_TEST_EQUAL(address.port, -1);
        APSARA_TEST_EQUAL(provider.mConfigServerAddressId, 0);
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).host, "");
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).port, -1);
        APSARA_TEST_EQUAL(provider.mConfigServerAddressId, 0);

        provider.Stop();
    }
}

void CommonConfigProviderUnittest::TestGetConfigUpdateAndConfigWatcher() {
    // add config
    {
        string config = R"(
        {
            "config_server_list": [
                {
                    "cluster": "community",
                    "endpoint_list": [
                        "test.config.com:80"
                    ]
                }
            ],
            "ilogtail_tags": {
                "key1": "value1",
                "key2": "value2"
            }
        }
        )";
        APSARA_TEST_TRUE_FATAL(writeJsonToFile(config, ilogtailConfigPath));

        MockCommonConfigProvider provider;
        auto setResponse
            = [&provider](const string& operation, const string& reqBody, const string& configType, std::string& resp) {
                  static int sequence_num = 0;
                  configserver::proto::v2::HeartbeatRequest heartbeatReq;
                  heartbeatReq.ParseFromString(reqBody);
                  APSARA_TEST_EQUAL(heartbeatReq.sequence_num(), sequence_num);
                  sequence_num++;
                  APSARA_TEST_TRUE(heartbeatReq.capabilities() & configserver::proto::v2::AcceptsInstanceConfig);
                  APSARA_TEST_TRUE(heartbeatReq.capabilities()
                                   & configserver::proto::v2::AcceptsContinuousPipelineConfig);
                  APSARA_TEST_EQUAL(heartbeatReq.instance_id(), provider.GetInstanceId());
                  APSARA_TEST_EQUAL(heartbeatReq.agent_type(), "LoongCollector");
                  APSARA_TEST_EQUAL(heartbeatReq.attributes().ip(), LoongCollectorMonitor::mIpAddr);
                  APSARA_TEST_EQUAL(heartbeatReq.attributes().hostname(), LoongCollectorMonitor::mHostname);
                  APSARA_TEST_EQUAL(heartbeatReq.attributes().version(), ILOGTAIL_VERSION);
                  auto it = heartbeatReq.tags().begin();
                  APSARA_TEST_EQUAL(it->name(), "key1");
                  APSARA_TEST_EQUAL(it->value(), "value1");
                  it++;
                  APSARA_TEST_EQUAL(it->name(), "key2");
                  APSARA_TEST_EQUAL(it->value(), "value2");

                  APSARA_TEST_EQUAL(heartbeatReq.running_status(), "running");
                  APSARA_TEST_EQUAL(heartbeatReq.startup_time(), provider.mStartTime);


                  configserver::proto::v2::HeartbeatResponse heartbeatRespPb;
                  heartbeatRespPb.set_capabilities(configserver::proto::v2::ResponseFlags::ReportFullState);
                  {
                      auto pipeline = heartbeatRespPb.mutable_continuous_pipeline_config_updates();
                      auto configDetail = pipeline->Add();
                      configDetail->set_name("config1");
                      configDetail->set_version(1);
                      configDetail->set_detail(R"(
                        {
                            "enable": true,
                            "flushers": [
                                {
                                    "OnlyStdout": true,
                                    "Type": "flusher_stdout"
                                }
                            ],
                            "inputs": [
                                {
                                    "FilePaths": [
                                        "\/workspaces\/test1.log"
                                    ],
                                    "Type": "input_file"
                                }
                            ]
                        }
                        )");
                      configDetail = pipeline->Add();
                      configDetail->set_name("config2");
                      configDetail->set_version(1);
                      configDetail->set_detail(R"(
                {
                    "enable": true,
                    "flushers": 
                        {
                            "OnlyStdout": true,
                            "Type": "flusher_stdout"
                        }
                    ],
                    "inputs": [
                        {
                            "FilePaths": [
                                "\/workspaces\/test1.log"
                            ],
                            "Type": "input_file"
                        }
                    ]
                }
                )");
                  }
                  {
                      auto instanceconfig = heartbeatRespPb.mutable_instance_config_updates();
                      auto configDetail = instanceconfig->Add();
                      configDetail->set_name("instanceconfig1");
                      configDetail->set_version(1);
                      configDetail->set_detail(R"(
                        {
                                "enable": true,
                                "max_bytes_per_sec": 100012031023
                            }
                        )");
                      configDetail = instanceconfig->Add();
                      configDetail->set_name("instanceconfig2");
                      configDetail->set_version(1);
                      configDetail->set_detail(R"(
                        {
                                "enable": true
                                "max_bytes_per_sec": 100012031023
                            }
                        )");
                  }
                  {
                      auto commandconfig = heartbeatRespPb.mutable_onetime_pipeline_config_updates();
                      auto configDetail = commandconfig->Add();
                      configDetail->set_name("commandconfig1");
                      configDetail->set_detail(R"(
                        {
                                "enable": true,
                                "max_bytes_per_sec": 100012031023
                            }
                        )");
                  }
                  heartbeatRespPb.SerializeToString(&resp);
              };

        EXPECT_CALL(provider, SendHttpRequest(::testing::_, ::testing::_, ::testing::_, ::testing::_, ::testing::_))
            .WillRepeatedly(::testing::DoAll(::testing::Invoke([setResponse](const string& operation,
                                                                             const string& reqBody,
                                                                             const string& configType,
                                                                             const std::string& requestId,
                                                                             std::string& resp) {
                                                 setResponse(operation, reqBody, configType, resp);
                                             }),
                                             ::testing::Return(true)));


        provider.Init("common_v2");
        APSARA_TEST_EQUAL(provider.sName, "common config provider");
        APSARA_TEST_EQUAL(provider.mConfigServerAvailable, true);
        APSARA_TEST_EQUAL(provider.mConfigServerAddresses.size(), 1);
        APSARA_TEST_EQUAL(provider.mConfigServerAddresses[0].host, "test.config.com");
        APSARA_TEST_EQUAL(provider.mConfigServerAddresses[0].port, 80);
        APSARA_TEST_EQUAL(provider.mConfigServerTags.size(), 2);
        APSARA_TEST_EQUAL(provider.mConfigServerTags["key1"], "value1");
        APSARA_TEST_EQUAL(provider.mConfigServerTags["key2"], "value2");
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).host, "test.config.com");
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).port, 80);
        APSARA_TEST_EQUAL(provider.mConfigServerAddressId, 0);
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(true).host, "test.config.com");
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(true).port, 80);
        APSARA_TEST_EQUAL(provider.mConfigServerAddressId, 0);

        auto heartbeatRequest = provider.PrepareHeartbeat();
        configserver::proto::v2::HeartbeatResponse heartbeatResponse;
        provider.GetConfigUpdate();

        APSARA_TEST_EQUAL(provider.mContinuousPipelineConfigInfoMap.size(), 2);
        APSARA_TEST_EQUAL(provider.mContinuousPipelineConfigInfoMap["config1"].status, ConfigFeedbackStatus::APPLYING);
        APSARA_TEST_EQUAL(provider.mContinuousPipelineConfigInfoMap["config2"].status, ConfigFeedbackStatus::FAILED);

        // 处理 pipelineconfig
        auto pipelineConfigDiff = PipelineConfigWatcher::GetInstance()->CheckConfigDiff();
        size_t builtinPipelineCnt = 0;
#ifdef __ENTERPRISE__
        builtinPipelineCnt += EnterpriseConfigProvider::GetInstance()->GetAllBuiltInPipelineConfigs().size();
#endif
        CollectionPipelineManager::GetInstance()->UpdatePipelines(pipelineConfigDiff.first);
        APSARA_TEST_TRUE(!pipelineConfigDiff.first.IsEmpty());
        APSARA_TEST_EQUAL(1U + builtinPipelineCnt, pipelineConfigDiff.first.mAdded.size());
        APSARA_TEST_EQUAL(pipelineConfigDiff.first.mAdded[builtinPipelineCnt].mName, "config1");
        APSARA_TEST_EQUAL(CollectionPipelineManager::GetInstance()->GetAllConfigNames().size(),
                          1U + builtinPipelineCnt);
        APSARA_TEST_TRUE(CollectionPipelineManager::GetInstance()->FindConfigByName("config1").get() != nullptr);
        // 再次处理 pipelineconfig
        pipelineConfigDiff = PipelineConfigWatcher::GetInstance()->CheckConfigDiff();
        CollectionPipelineManager::GetInstance()->UpdatePipelines(pipelineConfigDiff.first);
        APSARA_TEST_TRUE(pipelineConfigDiff.first.IsEmpty());
        APSARA_TEST_TRUE(pipelineConfigDiff.first.mAdded.empty());
        APSARA_TEST_EQUAL(CollectionPipelineManager::GetInstance()->GetAllConfigNames().size(),
                          1U + builtinPipelineCnt);
        APSARA_TEST_TRUE(CollectionPipelineManager::GetInstance()->FindConfigByName("config1").get() != nullptr);


        APSARA_TEST_EQUAL(provider.mInstanceConfigInfoMap.size(), 2);
        APSARA_TEST_EQUAL(provider.mInstanceConfigInfoMap["instanceconfig1"].status, ConfigFeedbackStatus::APPLYING);
        APSARA_TEST_EQUAL(provider.mInstanceConfigInfoMap["instanceconfig2"].status, ConfigFeedbackStatus::FAILED);

        // 处理 instanceconfig
        InstanceConfigDiff instanceConfigDiff = InstanceConfigWatcher::GetInstance()->CheckConfigDiff();
        InstanceConfigManager::GetInstance()->UpdateInstanceConfigs(instanceConfigDiff);
        APSARA_TEST_TRUE(!instanceConfigDiff.IsEmpty());
        APSARA_TEST_EQUAL(1U, instanceConfigDiff.mAdded.size());
        APSARA_TEST_EQUAL(instanceConfigDiff.mAdded[0].mConfigName, "instanceconfig1");
        if (BOOL_FLAG(logtail_mode)) {
            APSARA_TEST_EQUAL(InstanceConfigManager::GetInstance()->GetAllConfigNames().size(), 1);
        } else {
            APSARA_TEST_EQUAL(InstanceConfigManager::GetInstance()->GetAllConfigNames().size(), 2);
        }
        APSARA_TEST_EQUAL(InstanceConfigManager::GetInstance()->GetAllConfigNames()[0], "instanceconfig1");
        // 再次处理 instanceconfig
        instanceConfigDiff = InstanceConfigWatcher::GetInstance()->CheckConfigDiff();
        InstanceConfigManager::GetInstance()->UpdateInstanceConfigs(instanceConfigDiff);
        APSARA_TEST_TRUE(instanceConfigDiff.IsEmpty());
        APSARA_TEST_TRUE(instanceConfigDiff.mAdded.empty());
        if (BOOL_FLAG(logtail_mode)) {
            APSARA_TEST_EQUAL(InstanceConfigManager::GetInstance()->GetAllConfigNames().size(), 1);
        } else {
            APSARA_TEST_EQUAL(InstanceConfigManager::GetInstance()->GetAllConfigNames().size(), 2);
        }
        APSARA_TEST_EQUAL(InstanceConfigManager::GetInstance()->GetAllConfigNames()[0], "instanceconfig1");

        provider.Stop();
    }
    // test LoadConfigFile
    {
        MockCommonConfigProvider provider;
        provider.Init("common_v2");
        APSARA_TEST_EQUAL(provider.mContinuousPipelineConfigInfoMap.size(), 1);
        APSARA_TEST_EQUAL(provider.mContinuousPipelineConfigInfoMap["config1"].status, ConfigFeedbackStatus::APPLYING);
        APSARA_TEST_EQUAL(provider.mInstanceConfigInfoMap.size(), 1);
        APSARA_TEST_EQUAL(provider.mInstanceConfigInfoMap["instanceconfig1"].status, ConfigFeedbackStatus::APPLYING);
        provider.Stop();
    }
    // delete config
    {
        string config = R"(
        {
            "config_server_list": [
                {
                    "cluster": "community",
                    "endpoint_list": [
                        "test.config.com:80"
                    ]
                }
            ],
            "ilogtail_tags": {
                "key1": "value1",
                "key2": "value2"
            }
        }
        )";
        APSARA_TEST_TRUE_FATAL(writeJsonToFile(config, ilogtailConfigPath));

        MockCommonConfigProvider provider;
        auto setResponse
            = [&provider](const string& operation, const string& reqBody, const string& configType, std::string& resp) {
                  static int sequence_num = 0;
                  configserver::proto::v2::HeartbeatRequest heartbeatReq;
                  heartbeatReq.ParseFromString(reqBody);
                  APSARA_TEST_EQUAL(heartbeatReq.sequence_num(), sequence_num);
                  sequence_num++;
                  APSARA_TEST_TRUE(heartbeatReq.capabilities() & configserver::proto::v2::AcceptsInstanceConfig);
                  APSARA_TEST_TRUE(heartbeatReq.capabilities()
                                   & configserver::proto::v2::AcceptsContinuousPipelineConfig);
                  APSARA_TEST_EQUAL(heartbeatReq.instance_id(), provider.GetInstanceId());
                  APSARA_TEST_EQUAL(heartbeatReq.agent_type(), "LoongCollector");
                  APSARA_TEST_EQUAL(heartbeatReq.attributes().ip(), LoongCollectorMonitor::mIpAddr);
                  APSARA_TEST_EQUAL(heartbeatReq.attributes().hostname(), LoongCollectorMonitor::mHostname);
                  APSARA_TEST_EQUAL(heartbeatReq.attributes().version(), ILOGTAIL_VERSION);
                  auto it = heartbeatReq.tags().begin();
                  APSARA_TEST_EQUAL(it->name(), "key1");
                  APSARA_TEST_EQUAL(it->value(), "value1");
                  it++;
                  APSARA_TEST_EQUAL(it->name(), "key2");
                  APSARA_TEST_EQUAL(it->value(), "value2");

                  APSARA_TEST_EQUAL(heartbeatReq.running_status(), "running");
                  APSARA_TEST_EQUAL(heartbeatReq.startup_time(), provider.mStartTime);


                  configserver::proto::v2::HeartbeatResponse heartbeatRespPb;
                  heartbeatRespPb.set_capabilities(configserver::proto::v2::ResponseFlags::ReportFullState);
                  // pipeline
                  {
                      auto pipeline = heartbeatRespPb.mutable_continuous_pipeline_config_updates();
                      auto configDetail = pipeline->Add();
                      configDetail->set_name("config1");
                      configDetail->set_version(-1);
                      configDetail->set_detail(R"(
                        {
                            "enable": true,
                            "flushers": [
                                {
                                    "OnlyStdout": true,
                                    "Type": "flusher_stdout"
                                }
                            ],
                            "inputs": [
                                {
                                    "FilePaths": [
                                        "\/workspaces\/test1.log"
                                    ],
                                    "Type": "input_file"
                                }
                            ]
                        }
                        )");
                      configDetail = pipeline->Add();
                      configDetail->set_name("config2");
                      configDetail->set_version(-1);
                      configDetail->set_detail(R"(
                {
                    "enable": true,
                    "flushers": 
                        {
                            "OnlyStdout": true,
                            "Type": "flusher_stdout"
                        }
                    ],
                    "inputs": [
                        {
                            "FilePaths": [
                                "\/workspaces\/test1.log"
                            ],
                            "Type": "input_file"
                        }
                    ]
                }
                )");
                  }
                  // instanceconfig
                  {
                      auto instanceconfig = heartbeatRespPb.mutable_instance_config_updates();
                      auto configDetail = instanceconfig->Add();
                      configDetail->set_name("instanceconfig1");
                      configDetail->set_version(-1);
                      configDetail->set_detail(R"(
                        {
                                "enable": true,
                                "max_bytes_per_sec": 100012031023
                            }
                        )");
                      configDetail = instanceconfig->Add();
                      configDetail->set_name("instanceconfig2");
                      configDetail->set_version(-1);
                      configDetail->set_detail(R"(
                        {
                                "enable": true
                                "max_bytes_per_sec": 100012031023
                            }
                        )");
                  }
                  // commandconfig
                  {
                      auto commandconfig = heartbeatRespPb.mutable_onetime_pipeline_config_updates();
                      auto configDetail = commandconfig->Add();
                      configDetail->set_name("commandconfig1");
                      configDetail->set_detail(R"(
                        {
                                "enable": true,
                                "max_bytes_per_sec": 100012031023
                            }
                        )");
                  }
                  heartbeatRespPb.SerializeToString(&resp);
              };

        EXPECT_CALL(provider, SendHttpRequest(::testing::_, ::testing::_, ::testing::_, ::testing::_, ::testing::_))
            .WillRepeatedly(::testing::DoAll(::testing::Invoke([setResponse](const string& operation,
                                                                             const string& reqBody,
                                                                             const string& configType,
                                                                             const std::string& requestId,
                                                                             std::string& resp) {
                                                 setResponse(operation, reqBody, configType, resp);
                                             }),
                                             ::testing::Return(true)));


        provider.Init("common_v2");
        APSARA_TEST_EQUAL(provider.sName, "common config provider");
        APSARA_TEST_EQUAL(provider.mConfigServerAvailable, true);
        APSARA_TEST_EQUAL(provider.mConfigServerAddresses.size(), 1);
        APSARA_TEST_EQUAL(provider.mConfigServerAddresses[0].host, "test.config.com");
        APSARA_TEST_EQUAL(provider.mConfigServerAddresses[0].port, 80);
        APSARA_TEST_EQUAL(provider.mConfigServerTags.size(), 2);
        APSARA_TEST_EQUAL(provider.mConfigServerTags["key1"], "value1");
        APSARA_TEST_EQUAL(provider.mConfigServerTags["key2"], "value2");
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).host, "test.config.com");
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(false).port, 80);
        APSARA_TEST_EQUAL(provider.mConfigServerAddressId, 0);
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(true).host, "test.config.com");
        APSARA_TEST_EQUAL(provider.GetOneConfigServerAddress(true).port, 80);
        APSARA_TEST_EQUAL(provider.mConfigServerAddressId, 0);

        auto heartbeatRequest = provider.PrepareHeartbeat();
        configserver::proto::v2::HeartbeatResponse heartbeatResponse;
        provider.GetConfigUpdate();

        APSARA_TEST_TRUE(provider.mContinuousPipelineConfigInfoMap.empty());

        // 处理pipelineConfigDiff
        auto pipelineConfigDiff = PipelineConfigWatcher::GetInstance()->CheckConfigDiff();
        size_t builtinPipelineCnt = 0;
#ifdef __ENTERPRISE__
        builtinPipelineCnt += EnterpriseConfigProvider::GetInstance()->GetAllBuiltInPipelineConfigs().size();
#endif
        CollectionPipelineManager::GetInstance()->UpdatePipelines(pipelineConfigDiff.first);
        APSARA_TEST_TRUE(!pipelineConfigDiff.first.IsEmpty());
        APSARA_TEST_EQUAL(1U, pipelineConfigDiff.first.mRemoved.size());
        APSARA_TEST_EQUAL(pipelineConfigDiff.first.mRemoved[0], "config1");
        APSARA_TEST_EQUAL(0U + builtinPipelineCnt,
                          CollectionPipelineManager::GetInstance()->GetAllConfigNames().size());
        // 再次处理pipelineConfigDiff
        pipelineConfigDiff = PipelineConfigWatcher::GetInstance()->CheckConfigDiff();
        CollectionPipelineManager::GetInstance()->UpdatePipelines(pipelineConfigDiff.first);
        APSARA_TEST_TRUE(pipelineConfigDiff.first.IsEmpty());
        APSARA_TEST_TRUE(pipelineConfigDiff.first.mRemoved.empty());
        APSARA_TEST_EQUAL(0U + builtinPipelineCnt,
                          CollectionPipelineManager::GetInstance()->GetAllConfigNames().size());

        APSARA_TEST_TRUE(provider.mInstanceConfigInfoMap.empty());
        // 处理instanceConfigDiff
        InstanceConfigDiff instanceConfigDiff = InstanceConfigWatcher::GetInstance()->CheckConfigDiff();
        InstanceConfigManager::GetInstance()->UpdateInstanceConfigs(instanceConfigDiff);
        if (BOOL_FLAG(logtail_mode)) {
            APSARA_TEST_EQUAL(InstanceConfigManager::GetInstance()->GetAllConfigNames().size(), 0);
        } else {
            APSARA_TEST_EQUAL(InstanceConfigManager::GetInstance()->GetAllConfigNames().size(), 1);
            APSARA_TEST_EQUAL(InstanceConfigManager::GetInstance()->GetAllConfigNames()[0], "loongcollector_config");
        }
        APSARA_TEST_EQUAL(1U, instanceConfigDiff.mRemoved.size());
        APSARA_TEST_EQUAL(instanceConfigDiff.mRemoved[0], "instanceconfig1");

        // 再次处理instanceConfigDiff
        instanceConfigDiff = InstanceConfigWatcher::GetInstance()->CheckConfigDiff();
        InstanceConfigManager::GetInstance()->UpdateInstanceConfigs(instanceConfigDiff);
        if (BOOL_FLAG(logtail_mode)) {
            APSARA_TEST_EQUAL(InstanceConfigManager::GetInstance()->GetAllConfigNames().size(), 0);
        } else {
            APSARA_TEST_EQUAL(InstanceConfigManager::GetInstance()->GetAllConfigNames().size(), 1);
            APSARA_TEST_EQUAL(InstanceConfigManager::GetInstance()->GetAllConfigNames()[0], "loongcollector_config");
        }
        APSARA_TEST_TRUE(instanceConfigDiff.IsEmpty());
        APSARA_TEST_TRUE(instanceConfigDiff.mRemoved.empty());

        provider.Stop();
    }
}

UNIT_TEST_CASE(CommonConfigProviderUnittest, TestInit)
UNIT_TEST_CASE(CommonConfigProviderUnittest, TestGetConfigUpdateAndConfigWatcher)

}; // namespace logtail


int main(int argc, char** argv) {
    logtail::Logger::Instance().InitGlobalLoggers();
    ::testing::InitGoogleTest(&argc, argv);
    return RUN_ALL_TESTS();
}
