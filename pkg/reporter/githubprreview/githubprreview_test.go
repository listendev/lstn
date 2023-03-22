// SPDX-License-Identifier: Apache-2.0
//
// Copyright © 2023 The listen.dev team <engineering@garnet.ai>
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
package githubprreview

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/jarcoal/httpmock"
)

func TestReviewReporter_stickyComment(t *testing.T) {

	type args struct {
		owner   string
		repo    string
		id      int
		comment io.Reader
		mockFn  func()
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "issue with sticky comment",
			args: args{
				owner:   "fntlnz",
				repo:    "lstnrepotest",
				id:      1,
				comment: strings.NewReader("hello from listen.dev again"),
				mockFn: func() {
					issueData, err := os.ReadFile("testdata/issue_with_sticky_comment.json")
					if err != nil {
						t.Fatal(err)
					}
					issueUpdatedData, err := os.ReadFile("testdata/issue_with_sticky_comment_updated.json")
					if err != nil {
						t.Fatal(err)
					}
					httpmock.RegisterResponder("GET", "https://api.github.com/repos/fntlnz/lstnrepotest/issues/1/comments",
						httpmock.NewBytesResponder(200, issueData))

					httpmock.RegisterMatcherResponder("PATCH", "https://api.github.com/repos/fntlnz/lstnrepotest/issues/comments/1478606628",
						httpmock.BodyContainsString(`<!--@lstn-sticky-review-comment-->hello from listen.dev again`),
						httpmock.NewBytesResponder(200, issueUpdatedData))
				},
			},
		},
		{
			name: "issue without comment",
			args: args{
				owner:   "fntlnz",
				repo:    "lstnrepotest",
				id:      1,
				comment: strings.NewReader("hello from listen.dev first time ever"),
				mockFn: func() {
					issueData, err := os.ReadFile("testdata/issue_without_sticky_comment.json")
					if err != nil {
						t.Fatal(err)
					}
					issueUpdatedData, err := os.ReadFile("testdata/issue_without_sticky_comment_updated.json")
					if err != nil {
						t.Fatal(err)
					}
					httpmock.RegisterResponder("GET", "https://api.github.com/repos/fntlnz/lstnrepotest/issues/1/comments",
						httpmock.NewBytesResponder(200, issueData))

					httpmock.RegisterMatcherResponder("POST", "https://api.github.com/repos/fntlnz/lstnrepotest/issues/1/comments",
						httpmock.BodyContainsString(`<!--@lstn-sticky-review-comment-->hello from listen.dev first time ever`),
						httpmock.NewBytesResponder(200, issueUpdatedData))
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			tt.args.mockFn()

			ghClient := github.NewClient(nil)

			r := &ReviewReporter{
				ctx:      context.TODO(),
				ghClient: ghClient,
			}
			err := r.stickyComment(tt.args.owner, tt.args.repo, tt.args.id, tt.args.comment)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
