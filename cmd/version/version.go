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
			ctx := c.Context()

			// Obtain the (short) version info from the context
			shortVersion := ctx.Value(pkgcontext.ShortVersionKey).(string)

			outputString := fmt.Sprintf("lstn %s", shortVersion)
			changelogURL, _ := version.Changelog(shortVersion)
			if changelogURL != "" {
				outputString += fmt.Sprintf("\n%s", changelogURL)
			}

			fmt.Println(outputString)

			return nil
		},
	}

	return c, nil
}
