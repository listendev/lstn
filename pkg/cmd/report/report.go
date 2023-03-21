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
package report

import "github.com/listendev/lstn/pkg/listen"

type Renderer interface {
	Render(packages []listen.Package) error
}

type Report struct {
	reports []Report
}

func NewBuilder() *Report {
	return &Report{}
}

func (b *Report) RegisterReport(r Report) {
	b.reports = append(b.reports, r)
}

func (b *Report) Render(packages []listen.Package) error {
	for _, r := range b.reports {
		if err := r.Render(packages); err != nil {
			return err
		}
	}

	return nil
}
