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
package cmd

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/thediveo/enumflag/v2"
)

type Options interface {
	Validate() []error
	Transform(context.Context) error
}

type ReportType enumflag.Flag

const (
	AllReport ReportType = (iota + 1) * 11
	GitHubPullCommentReport
	GitHubPullReviewReport
	GitHubPullCheckReport
)

var AllReportTypes = []ReportType{
	GitHubPullCommentReport,
	GitHubPullReviewReport,
	GitHubPullCheckReport,
}

var ReporterTypeIDs = map[ReportType][]string{
	GitHubPullCommentReport: {GitHubPullCommentReport.String()},
	GitHubPullCheckReport:   {GitHubPullCheckReport.String()},
	GitHubPullReviewReport:  {GitHubPullReviewReport.String()},
}

func (t ReportType) String() string {
	switch t {
	case GitHubPullCommentReport:
		return "gh-pull-comment"
	case GitHubPullReviewReport:
		return "gh-pull-review"
	case GitHubPullCheckReport:
		return "gh-pull-check"
	default:
		return "all"
	}
}

func (t ReportType) Doc() string {
	lstn := "`lstn`"
	ghFlags := "`--gh-repo`, `--gh-owner`, `--gh-pull-id`"
	ghTokenFlag := "`--gh-token`"
	privateRepoScope := "`repo`"
	publicRepoScope := "`public_repo`"
	ghPullCheckReporter := "`gh-pull-check`"

	switch t {
	case GitHubPullCommentReport:
		ret := heredoc.Docf(`
It reports results as a sticky comment on the target GitHub pull request.

The target GitHub pull request comes from the values of the GitHub reporter flags (ie., %s).
Notice those values are automatically set when %s detects it is running in a GitHub Action.

### Status

Working.
`,
			ghFlags, lstn)

		return ret
	case GitHubPullReviewReport:
		ret := heredoc.Docf(`
It reports results to GitHub review comments on the target GitHub pull request.

### Token

It requires the GitHub token (%s) to contain a [personal access token (PAT)](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token).

For private repositories the GitHub token **MUST** have the %s scope, while for public repositories the %s scope is enough.

### Graceful degradation

When %s detects it is running from a fork repository, due to [GitHub Actions restrictions](https://docs.github.com/en/actions/security-guides/automatic-token-authentication#permissions-for-the-github_token) (ie., no write access to the GitHub Review API), this reporter will reports the verdicts to the GitHub Actions **log console**.

In such cases, %s fallbacks to [logging command of GitHub Actions](https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-error-message) to post results as [annotations](https://developer.github.com/v3/checks/runs/#annotations-object) similar to the %s reporter.

### Status

TBD.
`,
			ghTokenFlag, privateRepoScope, publicRepoScope, lstn, lstn, ghPullCheckReporter)

		return ret
	case GitHubPullCheckReport:
		ret := heredoc.Docf(`
It reports results to the GitHub pull requests check tab.

### Status

TBD.
`,
			lstn)

		return ret
	}

	return ""
}

func ParseReportType(s string) (ReportType, error) {
	for t, vals := range ReporterTypeIDs {
		for _, v := range vals {
			if s == v {
				return t, nil
			}
		}
	}

	return AllReport, fmt.Errorf(`a report type with ID "%s" doesn't exist`, s)
}
