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
package reporter

import (
	"context"
	"errors"

	"github.com/google/go-github/v50/github"
	"github.com/listendev/lstn/pkg/cmd"
	ghcomment "github.com/listendev/lstn/pkg/reporter/gh/comment"
	"github.com/listendev/lstn/pkg/reporter/request"
)

var (
	ErrReporterNotFound = errors.New("unsupported reporter")
)

type Reporter interface {
	Report(req *request.Report) error
	WithContext(ctx context.Context)
	WithGithubClient(client *github.Client)
}

func BuildReporter(reporterIdentifier cmd.ReportType) (Reporter, error) {
	switch reporterIdentifier {
	case cmd.GitHubPullCommentReport:
		return ghcomment.New(), nil
	default:
		return nil, ErrReporterNotFound
	}
}
