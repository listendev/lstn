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
	"net/http"
	"runtime"

	"github.com/XANi/goneric"
	"github.com/google/go-github/v53/github"
	"github.com/listendev/lstn/pkg/ci"
	"github.com/listendev/lstn/pkg/cmd/flags"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/lstn/pkg/reporter"
	"github.com/listendev/pkg/apispec"
	"github.com/listendev/pkg/type/int64string"
)

const proBaseURL = "https://core.listen.dev"

type rep struct {
	ctx       context.Context
	opts      *flags.ConfigFlags
	info      *ci.Info
	proClient apispec.ClientWithResponsesInterface
}

func New(ctx context.Context, opts ...reporter.Option) (reporter.Reporter, error) {
	// Retrieve the config options from the context
	// Those are mandatory because they contain the GitHub reporting options
	cfgOpts, ok := ctx.Value(pkgcontext.ConfigKey).(*flags.ConfigFlags)
	if cfgOpts == nil || !ok {
		return nil, fmt.Errorf("couldn't retrieve the config options")
	}

	proClient, err := apispec.NewClientWithResponses(proBaseURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't setup the client for our pro APIs: %w", err)
	}

	ret := &rep{
		ctx:       ctx,
		opts:      cfgOpts,
		proClient: proClient,
	}

	for _, opt := range opts {
		ret = opt(ret).(*rep)
	}

	if ret.info == nil {
		return nil, fmt.Errorf("couldn't retrieve info from the CI")
	}

	return ret, nil
}

func (r *rep) Run(res listen.Response) error {
	verdicts := res.Verdicts()
	// if len(verdicts) == 0 {
	// Assuming we have problems...
	// FIXME: signal if a package has a problem (yet to analyse) ?
	// }

	// Spawn API calls in parallel
	type wrap struct {
		res *apispec.PostApiV1DependenciesEventResponse
		err error
	}
	cb := func(v listen.Verdict) wrap {
		res, err := r.proClient.PostApiV1DependenciesEventWithResponse(
			r.ctx,
			getDependencyEvent(v, *r.info),
			attachAuthBearer(r),
		)

		return wrap{res, err}
	}
	returns := goneric.ParallelMapSlice(cb, runtime.NumCPU(), verdicts)

	// Check the number of responses matches the number of requests
	numReturns := len(returns)
	if numVerdicts := len(verdicts); numReturns != numVerdicts {
		return pkgcontext.OutputError(r.ctx, fmt.Errorf("wrong number of responses: %d responses for %d verdicts", numReturns, numVerdicts))
	}

	// Error out if we had one single error
	for _, ret := range returns {
		if ret.err != nil {
			return pkgcontext.OutputError(r.ctx, ret.err)
		}
	}

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

// TODO: impl (only for tests/mocks?)
func (r *rep) WithProClientBaseURL() {
}

func attachAuthBearer(r *rep) apispec.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		if req == nil {
			return fmt.Errorf("request is nil")
		}
		// We assume the JWT token is never blank (flag/option gets validated)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.opts.Token.JWT))

		return nil
	}
}

func getDependencyEvent(v listen.Verdict, i ci.Info) apispec.DependencyEvent {
	return apispec.PostApiV1DependenciesEventJSONRequestBody{
		Verdict: v,
		GithubContext: apispec.GitHubDependencyEventContext{
			Action:            i.Action,
			ActionPath:        &i.ActionPath,
			ActionRepository:  &i.ActionRepository,
			Actor:             i.Actor,
			ActorId:           int64string.Int64String(i.ActorID),
			EventName:         i.EventName,
			Job:               i.Job,
			Ref:               i.Ref,
			RefName:           i.RefName,
			RefProtected:      i.RefProtected,
			RefType:           i.RefType,
			Repository:        i.RepoFullName,
			RepositoryId:      int64string.Int64String(i.RepoID),
			RepositoryOwner:   i.RepoOwner,
			RepositoryOwnerId: int64string.Int64String(i.RepoOwnerID),
			RunAttempt:        int64string.Int64String(i.RunAttempt),
			RunId:             int64string.Int64String(i.RunID),
			RunNumber:         int64string.Int64String(i.RunNumber),
			RunnerArch:        i.RunnerArch,
			RunnerDebug:       &i.RunnerDebug,
			RunnerOs:          i.RunnerOs,
			ServerUrl:         i.SeverURL,
			Sha:               i.SHA,
			TriggeringActor:   i.TriggeringActor,
			Workflow:          i.Workflow,
			WorkflowRef:       i.WorkflowRef,
			Workspace:         i.Workspace,
		},
	}

	// FIXME: other fields that may be useful to send to the product
	// Owner            string
	// Repo             string
	// Num              int    // Pull (merge) request number
	// Branch           string // Pull (merge) request branch
	// Fork             bool
}
