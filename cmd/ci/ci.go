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
package ci

import (
	"context"
	"runtime"

	"github.com/listendev/lstn/cmd/ci/enable"
	"github.com/listendev/lstn/cmd/ci/report"
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
	var c = &cobra.Command{
		Use:                   "ci",
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		GroupID:               groups.Core.ID,
		Short:                 "Listen in on what your CI does",
		Long: `Eavesdrop everything happening under the hoods into your CI.

Using this set of commands, you can spy network and file activities occurring in your CI, whether it's your dependencies doing something shady or your code.

A listen.dev pro is necessary.`,
		Annotations: map[string]string{
			"source": project.GetSourceURL(filename),
		},
		Args: func(c *cobra.Command, args []string) error {
			if err := cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)(c, args); err != nil {
				_ = c.Help()

				return err
			}

			return nil
		},
		ValidArgs: []string{"enable", "report"},
		Run:       func(_ *cobra.Command, _ []string) {},
	}

	// Attach `enable` child command
	enableCmd, err := enable.New(ctx)
	if err != nil {
		return nil, err
	}
	c.AddCommand(enableCmd)

	// Attach `report` child command
	reportCmd, err := report.New(ctx)
	if err != nil {
		return nil, err
	}
	c.AddCommand(reportCmd)

	// Create the local options
	emptyOpts, err := options.NewEmpty()
	if err != nil {
		return nil, err
	}
	emptyOpts.Attach(c, []string{})

	// Pass the options through the context
	ctx = context.WithValue(ctx, pkgcontext.EmptyKey, emptyOpts)
	c.SetContext(ctx)

	return c, nil
}
