// SPDX-License-Identifier: Apache-2.0
//
// Copyright © 2025 The listen.dev team <engineering@garnet.ai>
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
package comment

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ghetzel/testify/require"
	"github.com/google/go-github/v53/github"
	"github.com/jarcoal/httpmock"
	"github.com/listendev/lstn/pkg/cmd/flags"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/reporter"
	"github.com/stretchr/testify/assert"
)

func TestReporterCanRun(t *testing.T) {
	type test struct {
		desc string
		opts *flags.ConfigFlags
		want bool
		erro string
	}
	testCases := []test{
		{
			desc: "missing-reporter-github-options",
			opts: nil,
			want: false,
			erro: "aaaa",
		},
		{
			desc: "not-on-pull-request",
			opts: &flags.ConfigFlags{
				Reporting: flags.Reporting{
					GitHub: flags.GitHub{
						Owner: "",
						Repo:  "",
						Pull: flags.Pull{
							ID: 0,
						},
					},
				},
			},
			want: false,
			erro: "",
		},
		{
			desc: "on-pull-request",
			opts: &flags.ConfigFlags{
				Reporting: flags.Reporting{
					GitHub: flags.GitHub{
						Owner: "listendev",
						Repo:  "lstn",
						Pull: flags.Pull{
							ID: 205,
						},
					},
				},
			},
			want: true,
			erro: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := t.Context()
			ctx = context.WithValue(ctx, pkgcontext.ConfigKey, tc.opts)

			r, err := New(ctx, reporter.WithGitHubClient(github.NewClient(nil)))
			if tc.erro != "" {
				assert.Nil(t, r)
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				rep, isRep := r.(*rep)
				require.True(t, isRep)
				assert.Equal(t, tc.want, rep.CanRun())
			}
		})
	}
}

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

					httpmock.RegisterResponder("GET", "https://api.github.com/repos/fntlnz/lstnrepotest/issues/comments/1478606628",
						httpmock.NewBytesResponder(200, issueUpdatedData))

					httpmock.RegisterMatcherResponder("PATCH", "https://api.github.com/repos/fntlnz/lstnrepotest/issues/comments/1478606628",
						httpmock.BodyContainsString(`<!--@lstn-sticky-review-comment-->\n\nhello from listen.dev again`),
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
						httpmock.BodyContainsString(`<!--@lstn-sticky-review-comment-->\n\nhello from listen.dev first time ever`),
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

			r := &rep{
				ctx:      t.Context(),
				ghClient: ghClient,
			}
			err := r.stickyComment(tt.args.owner, tt.args.repo, tt.args.id, tt.args.comment)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
