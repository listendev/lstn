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
package templates

import (
	"bytes"
	"testing"

	"github.com/listendev/lstn/pkg/listen"
	"github.com/stretchr/testify/require"
)

func TestRenderContainer(t *testing.T) {
	tests := []struct {
		name           string
		packages       []listen.Package
		expectedOutput []byte
		wantErr        bool
	}{
		{
			name:           "no packages",
			packages:       []listen.Package{},
			expectedOutput: testdataFileToBytes(t, "testdata/container_no_packages.md"),
			wantErr:        false,
		},
		{
			name: "with packages",
			packages: []listen.Package{
				{
					Name:    "foo",
					Version: "1.0.0",
					Verdicts: []listen.Verdict{
						{
							Message:  "outbound network connection",
							Priority: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "foo",
								"npm_package_version": "1.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
					},
				},
				{
					Name:    "bar",
					Version: "1.0.0",
					Verdicts: []listen.Verdict{
						{
							Message:  "outbound network connection",
							Priority: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "foo",
								"npm_package_version": "1.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
					},
				},
				{
					Name:    "foobar",
					Version: "1.0.0",
					Verdicts: []listen.Verdict{
						{
							Message:  "outbound network connection",
							Priority: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "foobar",
								"npm_package_version": "1.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Message:  "outbound network connection",
							Priority: "medium",
							Metadata: map[string]interface{}{
								"npm_package_name":    "foobar",
								"npm_package_version": "1.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Message:  "outbound network connection",
							Priority: "medium",
							Metadata: map[string]interface{}{
								"npm_package_name":    "foobar",
								"npm_package_version": "1.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Message:  "outbound network connection",
							Priority: "low",
							Metadata: map[string]interface{}{
								"npm_package_name":    "foobar",
								"npm_package_version": "1.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
					},
				},
			},
			expectedOutput: testdataFileToBytes(t, "testdata/container_with_packages.md"),
			wantErr:        false,
		},
		{
			name: "with problems",
			packages: []listen.Package{
				{
					Name:    "foo",
					Version: "1.0.0",
					Problems: []listen.Problem{
						{
							Type:   "https://listen.dev/probs/invalid-name",
							Title:  "Package name not valid",
							Detail: "Package name not valid",
						},
						{
							Type:   "https://listen.dev/probs/does-not-exist",
							Title:  "A problem that does not exist, just for testing",
							Detail: "A problem that does not exist, just for testing",
						},
					},
				},
				{
					Name:    "bar",
					Version: "1.2.0",
					Problems: []listen.Problem{
						{
							Type:   "https://listen.dev/probs/invalid-name",
							Title:  "Package name not valid",
							Detail: "Package name not valid",
						},
						{
							Type:   "https://listen.dev/probs/does-not-exist",
							Title:  "A problem that does not exist, just for testing",
							Detail: "A problem that does not exist, just for testing",
						},
						{
							Type:   "https://listen.dev/probs/something-something",
							Title:  "Something happened, must be aware",
							Detail: "Something happened, must be aware",
						},
					},
				},
			},
			expectedOutput: testdataFileToBytes(t, "testdata/container_with_problems.md"),
			wantErr:        false,
		},
		{
			name: "with verdicts and problems",
			packages: []listen.Package{
				{
					Name:    "foo",
					Version: "1.0.0",
					Verdicts: []listen.Verdict{
						{
							Message:  "outbound network connection",
							Priority: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "foo",
								"npm_package_version": "1.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
					},
				},
				{
					Name:    "foobar",
					Version: "1.0.0",
					Problems: []listen.Problem{
						{
							Type:   "https://listen.dev/probs/invalid-name",
							Title:  "Package name not valid",
							Detail: "Package name not valid",
						},
						{
							Type:   "https://listen.dev/probs/does-not-exist",
							Title:  "A problem that does not exist, just for testing",
							Detail: "A problem that does not exist, just for testing",
						},
					},
				},
				{
					Name:    "baz",
					Version: "1.0.0",
					Verdicts: []listen.Verdict{
						{
							Message:  "outbound network connection",
							Priority: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "baz",
								"npm_package_version": "1.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
					},
					Problems: []listen.Problem{
						{
							Type:   "https://listen.dev/probs/invalid-name",
							Title:  "Package name not valid",
							Detail: "Package name not valid",
						},
						{
							Type:   "https://listen.dev/probs/does-not-exist",
							Title:  "A problem that does not exist, just for testing",
							Detail: "A problem that does not exist, just for testing",
						},
					},
				},
			},
			expectedOutput: testdataFileToBytes(t, "testdata/container_with_verdicts_and_problems.md"),
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := &bytes.Buffer{}
			err := RenderContainer(outBuf, tt.packages)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderContainer() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			require.Equal(t, tt.expectedOutput, outBuf.Bytes())
		})
	}
}