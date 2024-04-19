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
package ci

import (
	"context"
	"fmt"
	"runtime"

	"github.com/listendev/lstn/internal/project"
	"github.com/listendev/lstn/pkg/cmd/groups"
	"github.com/listendev/lstn/pkg/cmd/options"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/spf13/cobra"
)

var (
	_, filename, _, _ = runtime.Caller(0)
)

func New(ctx context.Context) (*cobra.Command, error) {
	var ciCmd = &cobra.Command{
		Use:                   "ci",
		GroupID:               groups.Core.ID,
		DisableFlagsInUseLine: true,
		Short:                 "Listen in on what your CI does",
		Long: `Eavesdrop everything happening under the hoods into your CI.

Using this command, you can spy network and file activities occurring in your CI, whether it's your dependencies doing something shady or you.
This command requires a listen.dev pro account.`,
		Annotations: map[string]string{
			"source": project.GetSourceURL(filename),
		},
		RunE: func(c *cobra.Command, args []string) error {
			ctx = c.Context()
			// Obtain the local options from the context
			opts, err := pkgcontext.GetOptionsFromContext(ctx, pkgcontext.CiKey)
			if err != nil {
				return err
			}
			ciOpts, ok := opts.(*options.Ci)
			if !ok {
				return fmt.Errorf("couldn't obtain options for the current child command")
			}

			if ciOpts.DebugOptions {
				c.Println(ciOpts.AsJSON())

				return nil
			}

			return nil
		},
	}

	// Obtain the local options
	ciOpts, err := options.NewCi()
	if err != nil {
		return nil, err
	}

	// Persistent flags here will work for this command and all subcommands
	// ciCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Local flags will only run when this command is called directly
	ciOpts.Attach(ciCmd, []string{"--ignore-packages", "--ignore-deptypes", "--select", "lockfiles", "npm-endpoint", "pypi-endpoint", "reporter", "npm-registry", "gh-owner", "gh-pull-id", "gh-repo"})

	// Pass the options through the context
	ctx = context.WithValue(ctx, pkgcontext.CiKey, ciOpts)
	ciCmd.SetContext(ctx)

	return ciCmd, nil
}
