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
	"github.com/google/go-github/v50/github"
	"github.com/listendev/lstn/pkg/cmd/flags"
)

type Option func(Reporter) Reporter

func WithConfigOptions(opts *flags.ConfigFlags) Option {
	return func(r Reporter) Reporter {
		r.WithConfigOptions(opts)

		return r
	}
}

func WithGitHubClient(c *github.Client) Option {
	return func(r Reporter) Reporter {
		r.WithGitHubClient(c)

		return r
	}
}
