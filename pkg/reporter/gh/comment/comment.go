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
package comment

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/google/go-github/v53/github"
	"github.com/listendev/lstn/pkg/ci"
	"github.com/listendev/lstn/pkg/cmd/flags"
	"github.com/listendev/lstn/pkg/cmd/report"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/lstn/pkg/reporter"
)

const stickyReviewCommentAnnotation = "<!--@lstn-sticky-review-comment-->"

type rep struct {
	ctx      context.Context
	ghClient *github.Client
	opts     *flags.ConfigFlags
}

func New(ctx context.Context, opts ...reporter.Option) (reporter.Reporter, error) {
	// Retrieve the config options from the context
	// Those are mandatory because they contain the GitHub reporting options
	cfgOpts, ok := ctx.Value(pkgcontext.ConfigKey).(*flags.ConfigFlags)
	if cfgOpts == nil || !ok {
		return nil, fmt.Errorf("couldn't retrieve the config options")
	}

	ret := &rep{
		ctx:      ctx,
		opts:     cfgOpts,
		ghClient: github.NewTokenClient(ctx, cfgOpts.Token.GitHub),
	}

	for _, opt := range opts {
		ret = opt(ret).(*rep)
	}

	return ret, nil
}

func (r *rep) WithGitHubClient(client *github.Client) {
	r.ghClient = client
}

func (r *rep) WithConfigOptions(opts *flags.ConfigFlags) {
	r.opts = opts
}

func (r *rep) WithContinuousIntegrationInfo(_ *ci.Info) {
	// Do Nothing
}

func (r *rep) stickyComment(owner string, repo string, id int, comment io.Reader) error {
	buf := bytes.Buffer{}
	_, err := buf.WriteString(stickyReviewCommentAnnotation)
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

func (r *rep) Run(res interface{}, _ *string) error {
	owner := r.opts.Reporting.GitHub.Owner
	repo := r.opts.Reporting.GitHub.Repo
	id := r.opts.Reporting.GitHub.Pull.ID

	buf := bytes.Buffer{}

	switch v := res.(type) {
	case listen.Response:
		fullMarkdownReport := report.NewFullMarkdwonReport()
		fullMarkdownReport.WithOutput(&buf)

		if err := fullMarkdownReport.Render(v); err != nil {
			return err
		}

	case string:
		buf.WriteString(v)
	default:
		return fmt.Errorf("unsupported type: %T", res)
	}

	err := r.stickyComment(owner, repo, id, &buf)
	if err != nil {
		return err
	}

	return nil
}

// CanRun tells whether this reporter is being executed on a GitHub pull request
// (in which case it returns a true value) or not.
// FIXME: this is now unused (see pkg/reporter/factory/factory.go): it exists only for tests.
func (r *rep) CanRun() bool {
	ghOpts := r.opts.Reporting.GitHub

	return ghOpts.Owner != "" && ghOpts.Repo != "" && ghOpts.Pull.ID != 0
}
