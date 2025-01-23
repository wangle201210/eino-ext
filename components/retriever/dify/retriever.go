/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dify

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

// RetrieverConfig 定义了 Dify Retriever 的配置参数
type RetrieverConfig struct {
	// APIKey 是 Dify API 的认证密钥
	APIKey string `json:"api_key"`
	// Endpoint 是 Dify API 的服务地址
	Endpoint string `json:"endpoint"`
	// DatasetID 是知识库的唯一标识
	DatasetID string `json:"dataset_id"`
	// ScoreThreshold 是文档相关性评分的阈值
	ScoreThreshold *float64 `json:"score_threshold,omitempty"`
	// *retrievalModel `json:"retrieval_model,omitempty"`
	// TopK 定义了返回结果的最大数量
	TopK *int `json:"top_k,omitempty"`
	// Timeout 定义了 HTTP 连接超时时间
	Timeout time.Duration `json:"timeout,omitempty"`
	// SearchMethod 检索方法，如果不填则按照默认方式召回
	SearchMethod SearchMethod `json:"search_method"`
	// Weights 混合检索模式下语意检索的权重设置
	Weights float64 `json:"weights"`
	// ScoreThresholdEnabled 是否开启 score 阈值
	ScoreThresholdEnabled bool `json:"score_threshold_enabled"`
}

type Retriever struct {
	config         *RetrieverConfig
	client         *http.Client
	retrievalModel *retrievalModel
}

func NewRetriever(ctx context.Context, config *RetrieverConfig) (*Retriever, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if config.APIKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}
	if config.DatasetID == "" {
		return nil, fmt.Errorf("dataset_id is required")
	}

	if config.Endpoint == "" {
		config.Endpoint = defaultEndpoint
	}
	httpClient := &http.Client{}
	if config.Timeout != 0 {
		httpClient.Timeout = config.Timeout
	}
	return &Retriever{
		config:         config,
		client:         httpClient,
		retrievalModel: config.toRetrievalModel(),
	}, nil
}

// Retrieve 根据查询文本检索相关文档
func (r *Retriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) (docs []*schema.Document, err error) {
	// 设置回调和错误处理
	defer func() {
		if err != nil {
			ctx = callbacks.OnError(ctx, err)
		}
	}()

	// 合并检索选项
	options := retriever.GetCommonOptions(&retriever.Options{
		TopK:           r.config.TopK,
		ScoreThreshold: r.config.ScoreThreshold,
	}, opts...)

	// 开始检索回调
	ctx = callbacks.OnStart(ctx, &retriever.CallbackInput{
		Query:          query,
		TopK:           dereferenceOrZero(options.TopK),
		ScoreThreshold: options.ScoreThreshold,
	})

	// 发送检索请求
	result, err := r.doPost(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve documents: %w", err)
	}
	// 转换为统一的 Document 格式
	docs = make([]*schema.Document, 0, len(result.Records))

	// 转换为统一的 Document 格式
	for _, record := range result.Records {
		if record == nil || record.Segment == nil {
			continue
		}
		if options.ScoreThreshold != nil && record.Score < *options.ScoreThreshold {
			continue
		}
		doc := record.toDoc()
		docs = append(docs, doc)
	}

	// 结束检索回调
	ctx = callbacks.OnEnd(ctx, &retriever.CallbackOutput{Docs: docs})

	return docs, nil
}

func (r *Retriever) GetType() string {
	return typ
}

func (r *Retriever) IsCallbacksEnabled() bool {
	return true
}

func (x *RetrieverConfig) toRetrievalModel() *retrievalModel {
	if x == nil {
		return nil
	}
	return &retrievalModel{
		SearchMethod:          x.SearchMethod,
		Weights:               x.Weights,
		TopK:                  x.TopK,
		ScoreThresholdEnabled: x.ScoreThresholdEnabled,
		ScoreThreshold:        x.ScoreThreshold,
	}
}
