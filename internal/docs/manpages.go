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
package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func checkDirectory(dir string) error {
	if dir == "" {
		return fmt.Errorf("destination directory is empty")
	}
	info, erro := os.Stat(dir)
	if erro != nil {
		return fmt.Errorf("destination directory doesn't exist")
	}
	if !info.IsDir() {
		return fmt.Errorf("destination exists but it is not a directory")
	}

	return nil
}

func GenerateManTree(c *cobra.Command, dir string) error {
	if err := checkDirectory(dir); err != nil {
		return err
	}

	return genManTreeFromOpts(c, doc.GenManTreeOptions{
		Path:             dir,
		CommandSeparator: "-",
	})
}

func genManTreeFromOpts(c *cobra.Command, opts doc.GenManTreeOptions) error {
	header := opts.Header
	if header == nil {
		header = &doc.GenManHeader{}
	}
	for _, c := range c.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := genManTreeFromOpts(c, opts); err != nil {
			return err
		}
	}
	section := "1"
	if header.Section != "" {
		section = header.Section
	}

	separator := "_"
	if opts.CommandSeparator != "" {
		separator = opts.CommandSeparator
	}

	if header.Source == "" {
		header.Source = c.Annotations["source"]
	}

	basename := strings.ReplaceAll(c.CommandPath(), " ", separator)
	filename := filepath.Join(opts.Path, basename+"."+section)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	headerCopy := *header

	return doc.GenMan(c, &headerCopy, f)
}
