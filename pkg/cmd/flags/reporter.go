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
package flags

import (
	"context"

	"github.com/listendev/lstn/pkg/cmd"
)

var _ cmd.Options = (*ReporterFlags)(nil)

type ReporterFlags struct {
	Reporter string `name:"reporter" flag:"reporter" desc:"use a reporter" validate:"omitempty,reporter" default:""`

	// github-pr-review reporter flags
	GithubPRReviewReporterOwner      string `name:"github_pr_owner" flag:"github_pr_owner" desc:"PR owner name (organization/user)" validate:"required_if=Reporter github-pr-review"`
	GithubPRReviewReporterRepository string `name:"github_pr_repository" flag:"github_pr_repository" desc:"PR repository name" validate:"required_if=Reporter github-pr-review"`
	GithubPRReviewReporterPRID       int    `name:"github_pr_id" flag:"github_pr_id" desc:"PR repository name" validate:"required_if=Reporter github-pr-review"`
}

func (o *ReporterFlags) Validate() []error {
	return Validate(o)
}

func (o *ReporterFlags) Transform(ctx context.Context) error {
	return Transform(ctx, o)
}
