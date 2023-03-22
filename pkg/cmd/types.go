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
