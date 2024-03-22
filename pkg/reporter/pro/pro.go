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
package pro

import (
	"context"
	"fmt"

	"github.com/google/go-github/v53/github"
	"github.com/listendev/lstn/pkg/ci"
	"github.com/listendev/lstn/pkg/cmd/flags"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/lstn/pkg/reporter"
)

type rep struct {
	ctx  context.Context
	opts *flags.ConfigFlags
	info *ci.Info
}

func New(ctx context.Context, opts ...reporter.Option) (reporter.Reporter, error) {
	// Retrieve the config options from the context
	// Those are mandatory because they contain the GitHub reporting options
	cfgOpts, ok := ctx.Value(pkgcontext.ConfigKey).(*flags.ConfigFlags)
	if cfgOpts == nil || !ok {
		return nil, fmt.Errorf("couldn't retrieve the config options")
	}

	ret := &rep{
		ctx:  ctx,
		opts: cfgOpts,
	}

	for _, opt := range opts {
		ret = opt(ret).(*rep)
	}

	return ret, nil
}

func (r *rep) Run(res listen.Response) error {
	fmt.Println("TODO: for every listen.Package call the Dependency API", len(res))

	return nil
}

func (r *rep) WithConfigOptions(opts *flags.ConfigFlags) {
	r.opts = opts
}

func (r *rep) WithGitHubClient(_ *github.Client) {
	// Do nothing
}

func (r *rep) WithContinuousIntegrationInfo(info *ci.Info) {
	r.info = info
}
