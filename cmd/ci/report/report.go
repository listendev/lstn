// SPDX-License-Identifier: Apache-2.0
//
// Copyright © 2024 The listen.dev team <engineering@garnet.ai>
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
package report

import (
	"context"
	"fmt"
	"runtime"

	"github.com/MakeNowJust/heredoc"
	"github.com/listendev/lstn/internal/project"
	"github.com/listendev/lstn/pkg/cmd"
	"github.com/listendev/lstn/pkg/cmd/flags"
	"github.com/listendev/lstn/pkg/cmd/options"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	reporterfactory "github.com/listendev/lstn/pkg/reporter/factory"
	"github.com/listendev/lstn/pkg/validate"
	"github.com/spf13/cobra"
)

var _, filename, _, _ = runtime.Caller(0)

func New(ctx context.Context) (*cobra.Command, error) {
	c := &cobra.Command{
		Use:                   "report",
		DisableFlagsInUseLine: true,
		Short:                 "Report the most critical findings into GitHub pull requests",
		Annotations: map[string]string{
			"source": project.GetSourceURL(filename),
		},
		RunE: func(c *cobra.Command, _ []string) error {
			ctx = c.Context()
			// Obtain the local options from the context
			optsFromContext, err := pkgcontext.GetOptionsFromContext(ctx, pkgcontext.CiReportKey)
			if err != nil {
				return err
			}
			opts, ok := optsFromContext.(*options.CiReport)
			if !ok {
				return fmt.Errorf("couldn't obtain options for the current child command")
			}

			// Token options are mandatory in this case
			errs := []error{}
			// GitHub token is mandatory for reporting (posting the comment)
			if err := validate.Singleton.Var(opts.Token.GitHub, "mandatory"); err != nil {
				tags, _ := flags.GetFieldTag(opts, "ConfigFlags.Token.GitHub")
				name, _ := tags.Lookup("name")
				errs = append(errs, flags.Translate(err, name)...)
			}
			// The listen.dev token is mandatory for fetching the data to report
			if err := validate.Singleton.Var(opts.Token.JWT, "mandatory"); err != nil {
				tags, _ := flags.GetFieldTag(opts, "ConfigFlags.Token.JWT")
				name, _ := tags.Lookup("name")
				errs = append(errs, flags.Translate(err, name)...)
			}
			if len(errs) > 0 {
				ret := "invalid configuration options/flags"
				for _, e := range errs {
					ret += "\n       "
					ret += e.Error()
				}

				return fmt.Errorf("%s", ret)
			}

			if opts.DebugOptions {
				c.Println(opts.AsJSON())

				return nil
			}
			source := "eavesdrop tool"
			reportingOpts := flags.Reporting{Types: []cmd.ReportType{cmd.GitHubPullCommentReport}}
			noFindingsMessage := heredoc.Doc(`
			The listen.dev eavesdropping engine didn't spot any critical runtime threat.

			Congrats, you're safe!`)

			// TODO: fetch critical findings from Core API

			return reporterfactory.Exec(c, reportingOpts, noFindingsMessage, &source)
		},
	}

	// Create the local options
	reportOpts, err := options.NewCiReport()
	if err != nil {
		return nil, err
	}
	// Local flags will only run when this command is called directly
	reportOpts.Attach(c, []string{"npm-registry", "select", "ignore-deptypes", "ignore-packages", "pypi-endpoint", "npm-endpoint", "lockfiles", "reporter"})

	// Pass the options through the context
	ctx = context.WithValue(ctx, pkgcontext.CiReportKey, reportOpts)
	c.SetContext(ctx)

	return c, nil
}
