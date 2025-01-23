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
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	. "github.com/bytedance/mockey"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/smartystreets/goconvey/convey"
)

func TestNewRetriever(t *testing.T) {
	PatchConvey("test NewRetriever", t, func() {
		ctx := context.Background()

		PatchConvey("test config validation", func() {
			PatchConvey("test nil config", func() {
				ret, err := NewRetriever(ctx, nil)
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(err.Error(), convey.ShouldContainSubstring, "config is required")
				convey.So(ret, convey.ShouldBeNil)
			})

			PatchConvey("test empty api_key", func() {
				ret, err := NewRetriever(ctx, &RetrieverConfig{
					Endpoint:  "https://api.dify.ai/v1",
					DatasetID: "test",
				})
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(err.Error(), convey.ShouldContainSubstring, "api_key is required")
				convey.So(ret, convey.ShouldBeNil)
			})

			PatchConvey("test empty endpoint", func() {
				_, err := NewRetriever(ctx, &RetrieverConfig{
					APIKey:    "test",
					DatasetID: "test",
				})
				convey.So(err, convey.ShouldBeNil)
			})

			PatchConvey("test empty dataset_id", func() {
				ret, err := NewRetriever(ctx, &RetrieverConfig{
					APIKey:   "test",
					Endpoint: "https://api.dify.ai/v1",
				})
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(err.Error(), convey.ShouldContainSubstring, "dataset_id is required")
				convey.So(ret, convey.ShouldBeNil)
			})
		})

		PatchConvey("test success", func() {
			ret, err := NewRetriever(ctx, &RetrieverConfig{
				APIKey:    "test",
				Endpoint:  "https://api.dify.ai/v1",
				DatasetID: "test",
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(ret, convey.ShouldNotBeNil)
		})
	})
}

func TestRetrieve(t *testing.T) {
	PatchConvey("test Retrieve", t, func() {
		ctx := context.Background()
		r := &Retriever{
			config: &RetrieverConfig{
				APIKey:    "test",
				Endpoint:  "https://api.dify.ai/v1",
				DatasetID: "test",
				TopK:      ptrOf(10),
			},
			client: &http.Client{},
		}

		PatchConvey("test request error", func() {
			Mock(GetMethod(r.client, "Do")).Return(&http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"request failed"}}`)),
			}, nil).Build()

			docs, err := r.Retrieve(ctx, "test query")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "request failed")
			convey.So(docs, convey.ShouldBeNil)
		})

		PatchConvey("test response status error", func() {
			Mock(GetMethod(r.client, "Do")).Return(&http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"mock error"}}`)),
			}, nil).Build()

			docs, err := r.Retrieve(ctx, "test query")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "request failed")
			convey.So(docs, convey.ShouldBeNil)
		})

		PatchConvey("test success", func() {
			response := &successResponse{
				Query: &Query{Content: "test query"},
				Records: []*Record{
					{
						Score: 0.8,
						Segment: &Segment{
							Id:      "1",
							Content: "test content 1",
						},
					},
					{
						Score: 0.6,
						Segment: &Segment{
							Id:      "2",
							Content: "test content 2",
						},
					},
				},
			}

			respBytes, _ := json.Marshal(response)
			Mock(GetMethod(r.client, "Do")).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(string(respBytes))),
			}, nil).Build()

			PatchConvey("test without score threshold", func() {
				docs, err := r.Retrieve(ctx, "test query")
				convey.So(err, convey.ShouldBeNil)
				convey.So(len(docs), convey.ShouldEqual, 2)

				convey.So(docs[0].ID, convey.ShouldEqual, "1")
				convey.So(docs[0].Content, convey.ShouldEqual, "test content 1")
				convey.So(docs[0].MetaData["_score"], convey.ShouldEqual, 0.8)

			})

			PatchConvey("test with score threshold", func() {
				docs, err := r.Retrieve(ctx, "test query", retriever.WithScoreThreshold(0.7))
				convey.So(err, convey.ShouldBeNil)
				convey.So(len(docs), convey.ShouldEqual, 1)

				convey.So(docs[0].ID, convey.ShouldEqual, "1")
				convey.So(docs[0].Content, convey.ShouldEqual, "test content 1")
				convey.So(docs[0].MetaData["_score"], convey.ShouldEqual, 0.8)

			})
		})
	})
}

func TestNewRetrieverWithSearchMethod(t *testing.T) {
	PatchConvey("test NewRetriever with search method", t, func() {
		ctx := context.Background()

		PatchConvey("test full text search", func() {
			ret, err := NewRetriever(ctx, &RetrieverConfig{
				APIKey:       "test",
				Endpoint:     "https://api.dify.ai/v1",
				DatasetID:    "test",
				SearchMethod: SearchMethodFullText,
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(ret, convey.ShouldNotBeNil)
			convey.So(ret.retrievalModel, convey.ShouldNotBeNil)
			convey.So(ret.retrievalModel.SearchMethod, convey.ShouldEqual, SearchMethodFullText)
		})

		PatchConvey("test hybrid search with weights", func() {
			ret, err := NewRetriever(ctx, &RetrieverConfig{
				APIKey:       "test",
				Endpoint:     "https://api.dify.ai/v1",
				DatasetID:    "test",
				SearchMethod: SearchMethodHybrid,
				Weights:      0.7,
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(ret, convey.ShouldNotBeNil)
			convey.So(ret.retrievalModel, convey.ShouldNotBeNil)
			convey.So(ret.retrievalModel.SearchMethod, convey.ShouldEqual, SearchMethodHybrid)
			convey.So(ret.retrievalModel.Weights, convey.ShouldEqual, 0.7)
		})

		PatchConvey("test with score threshold", func() {
			threshold := 0.8
			ret, err := NewRetriever(ctx, &RetrieverConfig{
				APIKey:                "test",
				Endpoint:              "https://api.dify.ai/v1",
				DatasetID:             "test",
				SearchMethod:          SearchMethodFullText,
				ScoreThreshold:        &threshold,
				ScoreThresholdEnabled: true,
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(ret, convey.ShouldNotBeNil)
			convey.So(ret.retrievalModel, convey.ShouldNotBeNil)
			convey.So(ret.retrievalModel.ScoreThresholdEnabled, convey.ShouldBeTrue)
			convey.So(*ret.retrievalModel.ScoreThreshold, convey.ShouldEqual, threshold)
		})
	})
}

func TestGetType(t *testing.T) {
	PatchConvey("test GetType", t, func() {
		r := &Retriever{}
		convey.So(r.GetType(), convey.ShouldEqual, typ)
	})
}

func TestIsCallbacksEnabled(t *testing.T) {
	PatchConvey("test IsCallbacksEnabled", t, func() {
		r := &Retriever{}
		convey.So(r.IsCallbacksEnabled(), convey.ShouldBeTrue)
	})
}

func TestToRetrievalModel(t *testing.T) {
	PatchConvey("test toRetrievalModel", t, func() {
		PatchConvey("test nil config", func() {
			var config *RetrieverConfig
			model := config.toRetrievalModel()
			convey.So(model, convey.ShouldBeNil)
		})

		PatchConvey("test with search method", func() {
			threshold := 0.8
			config := &RetrieverConfig{
				SearchMethod:          SearchMethodFullText,
				Weights:               0.7,
				TopK:                  ptrOf(10),
				ScoreThreshold:        &threshold,
				ScoreThresholdEnabled: true,
			}
			model := config.toRetrievalModel()
			convey.So(model, convey.ShouldNotBeNil)
			convey.So(model.SearchMethod, convey.ShouldEqual, SearchMethodFullText)
			convey.So(model.Weights, convey.ShouldEqual, 0.7)
			convey.So(*model.TopK, convey.ShouldEqual, 10)
			convey.So(*model.ScoreThreshold, convey.ShouldEqual, threshold)
			convey.So(model.ScoreThresholdEnabled, convey.ShouldBeTrue)
		})
	})
}
