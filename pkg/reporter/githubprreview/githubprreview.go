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
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/listendev/lstn/pkg/cmd/report"
	"github.com/listendev/lstn/pkg/reporter/request"
	"github.com/listendev/lstn/pkg/validate"
)

const ReporterIdentifier = "github-pr-review"

func init() {
	validate.RegisterAvailableReporter(ReporterIdentifier)
}

const stickyReviewCommentAnnotation = "<!--@lstn-sticky-review-comment-->"

type ReviewReporter struct {
	ctx      context.Context
	ghClient *github.Client
}

func New() *ReviewReporter {
	return &ReviewReporter{
		ghClient: github.NewClient(nil),
	}
}

func (r *ReviewReporter) WithGithubClient(client *github.Client) {
	r.ghClient = client
}

func (r *ReviewReporter) WithContext(ctx context.Context) {
	r.ctx = ctx
}

func (r *ReviewReporter) stickyComment(owner string, repo string, id int, comment io.Reader) error {
	buf := bytes.Buffer{}
	_, err := buf.Write([]byte(stickyReviewCommentAnnotation))
	if err != nil {
		return err
	}

	_, err = io.Copy(&buf, comment)
	if err != nil {
		return err
	}

	comments, _, err := r.ghClient.Issues.ListComments(r.ctx, owner, repo, id, nil)
	if err != nil {
		return err
	}
	issueComment := &github.IssueComment{
		Body: github.String(buf.String()),
	}
	commentFn := func() error {
		_, _, err = r.ghClient.Issues.CreateComment(r.ctx, owner, repo, id, issueComment)

		return err
	}
	for _, comment := range comments {
		if strings.HasPrefix(*comment.Body, stickyReviewCommentAnnotation) {
			commentFn = func() error {
				_, _, err = r.ghClient.Issues.EditComment(r.ctx, owner, repo, *comment.ID, issueComment)

				return err
			}

			break
		}
	}

	return commentFn()
}

func (r *ReviewReporter) Report(req *request.Report) error {
	buf := bytes.Buffer{}
	fullMarkdownReport := report.NewFullMarkdwonReport()
	fullMarkdownReport.WithOutput(&buf)
	if err := fullMarkdownReport.Render(req.Packages); err != nil {
		return err
	}

	owner := req.GithubPRReviewRequest.Owner
	repo := req.GithubPRReviewRequest.Repo
	id := req.GithubPRReviewRequest.ID

	err := r.stickyComment(owner, repo, id, &buf)
	if err != nil {
		return err
	}

	return nil
}
