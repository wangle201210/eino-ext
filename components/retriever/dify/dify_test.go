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
	"testing"

	. "github.com/bytedance/mockey"
	"github.com/cloudwego/eino/schema"
	"github.com/smartystreets/goconvey/convey"
)

func TestRecord_ToDoc(t *testing.T) {
	PatchConvey("Test Record.toDoc", t, func() {
		PatchConvey("When record is valid", func() {
			record := &Record{
				Segment: &Segment{
					Id:         "1",
					Content:    "test content",
					DocumentId: "doc1",
				},
				Score: 0.8,
			}
			expected := &schema.Document{
				ID:      "1",
				Content: "test content",
				MetaData: map[string]interface{}{
					orgDocIDKey: "doc1",
				},
			}

			result := record.toDoc()
			convey.So(result.ID, convey.ShouldEqual, expected.ID)
			convey.So(result.Content, convey.ShouldEqual, expected.Content)
			convey.So(result.MetaData[orgDocIDKey], convey.ShouldEqual, expected.MetaData[orgDocIDKey])
		})

		PatchConvey("When record is nil", func() {
			var record *Record
			result := record.toDoc()
			convey.So(result, convey.ShouldBeNil)
		})

		PatchConvey("When segment is nil", func() {
			record := &Record{
				Segment: nil,
				Score:   0.8,
			}
			result := record.toDoc()
			convey.So(result, convey.ShouldBeNil)
		})
	})
}

func TestMetadataFunctions(t *testing.T) {
	PatchConvey("Test metadata functions", t, func() {
		PatchConvey("When document is not nil", func() {
			doc := &schema.Document{
				MetaData: map[string]interface{}{},
			}

			PatchConvey("Test OrgDocID functions", func() {
				setOrgDocID(doc, "doc1")
				convey.So(GetOrgDocID(doc), convey.ShouldEqual, "doc1")
			})

			PatchConvey("Test OrgDocName functions", func() {
				setOrgDocName(doc, "test doc")
				convey.So(GetOrgDocName(doc), convey.ShouldEqual, "test doc")
			})

			PatchConvey("Test Keywords functions", func() {
				keywords := []string{"test", "keywords"}
				setKeywords(doc, keywords)
				convey.So(GetKeywords(doc), convey.ShouldResemble, keywords)
			})
		})

		PatchConvey("When document is nil", func() {
			var nilDoc *schema.Document
			convey.So(GetOrgDocID(nilDoc), convey.ShouldEqual, "")
			convey.So(GetOrgDocName(nilDoc), convey.ShouldEqual, "")
			convey.So(GetKeywords(nilDoc), convey.ShouldBeNil)
		})
	})
}
