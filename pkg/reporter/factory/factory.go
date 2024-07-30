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
package factory

import (
	"context"
	"errors"
	"fmt"

	"github.com/cli/cli/pkg/iostreams"
	"github.com/listendev/lstn/pkg/ci"
	"github.com/listendev/lstn/pkg/cmd"
	"github.com/listendev/lstn/pkg/cmd/flags"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/lstn/pkg/reporter"
	ghcomment "github.com/listendev/lstn/pkg/reporter/gh/comment"
	"github.com/listendev/lstn/pkg/reporter/pro"
	"github.com/spf13/cobra"
)

var (
	ErrReporterNotFound               = errors.New("unsupported reporter")
	ErrReporterUnsupportedEnvironment = errors.New("the reporter is not running in a supported environment")
	ErrReporterNotOnPullRequest       = errors.New("the reporter is not running against a GitHub pull request")
	ErrReporterCantWrite              = errors.New("the GitHub token the reporter is running with is read-only")
	ErrReporterOnFork                 = errors.New("the GitHub action is running on a pull request of a fork")
)

// Make creates a new reporter.Reporter.
//
// Depending on the input cmd.ReportType and the current context,
// it ensures that the reporter.Reporter can actually run.
// When the reporter.Reporter cannot run in the calling setup,
// this function returns a false value for the canRun return value.
// In all the other cases (even when it errors for other reasons),
// it returns a true value for the canRun return value.
//
// Last but not least, this function takes care of configuring
// everything the reporter being instantiated needs.
func Make(ctx context.Context, reportType cmd.ReportType) (r reporter.Reporter, canRun bool, err error) {
	switch reportType {
	case cmd.ListenPro:
		info, infoErr := ci.NewInfo()
		if infoErr != nil {
			return nil, false, ErrReporterUnsupportedEnvironment
		}

		// This reporter doesn't run on fork's pull requests at the moment
		// It can run on GitHub events other than pull_request tho since it only calls our core APIs
		if info.HasReadOnlyGitHubToken() {
			return nil, false, ErrReporterOnFork
		}

		r, err := pro.New(ctx, reporter.WithContinuousIntegrationInfo(info))
		if err != nil {
			return nil, true, err
		}

		return r, true, nil

	case cmd.GitHubPullCommentReport:
		r, err := ghcomment.New(ctx)
		if err != nil {
			return nil, true, err
		}

		env, envErr := ci.NewInfo()
		if envErr != nil {
			return nil, false, ErrReporterUnsupportedEnvironment
		}

		// This reporter can only work on pull requests because it comments on them (with a sticky comment)
		if !env.IsGitHubPullRequest() {
			return nil, false, ErrReporterNotOnPullRequest
		}

		if env.HasReadOnlyGitHubToken() {
			// TODO: here we could fallback to another GitHub reporter (not existing yet) (ie., GitHubLoggingReport)
			// NOTE: see links below
			// https://docs.github.com/en/actions/reference/events-that-trigger-workflows#pull_request_target
			// https://help.github.com/en/actions/automating-your-workflow-with-github-actions/development-tools-for-github-actions#logging-commands.

			return nil, false, ErrReporterCantWrite
		}

		return r, true, nil
	default:
		return nil, true, ErrReporterNotFound
	}
}

func Exec(c *cobra.Command, reportingOpts flags.Reporting, resp listen.Response, source *string) error {
	ctx := c.Context()
	var cs *iostreams.ColorScheme
	io, ok := ctx.Value(pkgcontext.IOStreamsKey).(*iostreams.IOStreams)
	if ok {
		cs = io.ColorScheme()
	}

	for _, r := range reportingOpts.Types {
		rString := fmt.Sprintf("%q", r.String())
		if cs != nil {
			rString = cs.Gray(rString)
		}
		c.Printf("Reporting using the %s reporter...\n", rString)

		switch r {
		case cmd.ListenPro:
			fallthrough

		case cmd.GitHubPullCommentReport:
			rep, runnable, err := Make(c.Context(), r)
			if runnable && err != nil {
				return err
			}
			// Move on when the current reporter cannot run in the current context
			if !runnable {
				c.PrintErrf("Exiting: %s.\n", err)

				continue
			}

			err = rep.Run(resp, source)
			if err != nil {
				return fmt.Errorf("error while executing the %q reporter: %w", r.String(), err)
			}
			c.Printf("The report has been successfully sent using the %s reporter... %s\n", rString, cs.SuccessIcon())

		case cmd.GitHubPullCheckReport:
			c.PrintErrf("The %s reporter is coming soon...\n", rString)

		case cmd.GitHubPullReviewReport:
			c.PrintErrf("The %s reporter is coming soon...\n", rString)
		}
	}

	return nil
}
