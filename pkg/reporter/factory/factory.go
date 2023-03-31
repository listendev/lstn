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

	"github.com/listendev/lstn/pkg/cmd"
	"github.com/listendev/lstn/pkg/reporter"
	ghcomment "github.com/listendev/lstn/pkg/reporter/gh/comment"
)

var (
	ErrReporterNotFound         = errors.New("unsupported reporter")
	ErrReporterNotOnPullRequest = errors.New("the reporter is not running against a GitHub pull request")
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
	case cmd.GitHubPullCommentReport:
		r, err := ghcomment.New(ctx)
		if err != nil {
			return nil, true, err
		}
		if !r.CanRun() {
			return nil, false, ErrReporterNotOnPullRequest
		}

		return r, true, nil
	default:
		return nil, true, ErrReporterNotFound
	}
}
