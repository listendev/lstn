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
package version

import (
	"context"
	"fmt"
	"runtime"

	"github.com/listendev/lstn/internal/project"
	"github.com/listendev/lstn/pkg/cmd/options"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/version"
	"github.com/spf13/cobra"
)

var (
	_, filename, _, _ = runtime.Caller(0)
)

func New(ctx context.Context) (*cobra.Command, error) {
	var c = &cobra.Command{
		Use:                   "version",
		Short:                 "Print out version information",
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			"source": project.GetSourceURL(filename),
		},
		RunE: func(c *cobra.Command, args []string) error {
			ctx = c.Context()

			// Obtain the local options from the context
			opts, err := pkgcontext.GetOptionsFromContext(ctx, pkgcontext.VersionKey)
			if err != nil {
				return err
			}
			localOpts, ok := opts.(*options.Version)
			if !ok {
				return fmt.Errorf("couldn't obtain options for the current child command")
			}

			if localOpts.DebugOptions {
				c.Println(localOpts.AsJSON())

				return nil
			}

			// Obtain the version info from the context
			v := ctx.Value(pkgcontext.VersionTagKey).(string)
			switch localOpts.Verbosity {
			case 0:
				// default to version tag
			case 1:
				v = ctx.Value(pkgcontext.VersionShortKey).(string)
			case 2:
				fallthrough

			default:
				v = ctx.Value(pkgcontext.VersionLongKey).(string)
			}

			outputString := fmt.Sprintf("lstn %s", v)
			if localOpts.Changelog {
				changelogURL, _ := version.Changelog(v)
				if changelogURL != "" {
					outputString += fmt.Sprintf("\n%s", changelogURL)
				}
			}

			c.Println(outputString)

			return nil
		},
	}

	localOpts, err := options.NewVersion()
	if err != nil {
		return nil, err
	}

	// Local flags will only run when this command is called directly
	localOpts.Attach(c, []string{})

	// Pass the options through the context
	ctx = context.WithValue(ctx, pkgcontext.VersionKey, localOpts)
	c.SetContext(ctx)

	return c, nil
}
