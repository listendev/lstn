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

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/listendev/lstn/cmd/root"
	"github.com/listendev/lstn/internal/docs"
	"github.com/spf13/cobra"
)

func main() {
	if err := newRoot(context.Background()); err != nil {
		os.Exit(1)
	}
}

type rootFlags struct {
	Dest string
}

type contextKey string

var rootFlagsKey contextKey = "rf"

func getManpages(ctx context.Context) *cobra.Command {
	c := &cobra.Command{
		Use:                   "manpages",
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		Short:                 "Generate lstn manpages.",
		RunE: func(c *cobra.Command, args []string) error {
			rootFlagValues, ok := ctx.Value(rootFlagsKey).(*rootFlags)
			if !ok {
				return fmt.Errorf("couldn't obtain the flag values from the context")
			}
			lstn, _ := root.New(ctx)

			return docs.GenerateManTree(lstn.Command(), rootFlagValues.Dest)
		},
	}

	return c
}

func newRoot(ctx context.Context) error {
	rootF := &rootFlags{}

	rootC := &cobra.Command{
		Use:          "docs",
		SilenceUsage: true,
		Short:        "Generate lstn documentation.",
		Args:         cobra.ExactArgs(1),
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			// Eventually validate rootF struct

			return nil
		},
	}

	rootC.PersistentFlags().StringVarP(&rootF.Dest, "dest", "d", rootF.Dest, "path directory where to generate the documentation files")
	if err := rootC.MarkPersistentFlagRequired("dest"); err != nil {
		return err
	}

	ctx = context.WithValue(ctx, rootFlagsKey, rootF)

	manpagesC := getManpages(ctx)
	rootC.AddCommand(manpagesC)

	return rootC.ExecuteContext(ctx)
}
