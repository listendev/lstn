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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/pkg/verdictcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataFileToBytes(t *testing.T, dataFile string) []byte {
	b, err := os.ReadFile(dataFile)
	if err != nil {
		t.Fatal(err)
	}

	return b
}

func strPtr(s string) *string {
	return &s
}

func TestRenderContainer(t *testing.T) {
	tests := []struct {
		name           string
		packages       []listen.Package
		expectedOutput []byte
		snapshot       bool
		wantErr        bool
	}{
		{
			snapshot:       true,
			name:           "no packages",
			packages:       []listen.Package{},
			expectedOutput: testdataFileToBytes(t, "testdata/container_no_packages.md"),
			wantErr:        false,
		},
		{
			snapshot: true,
			name:     "with packages",
			packages: []listen.Package{
				{
					Name:    "react",
					Version: strPtr("18.0.0"),
					Verdicts: []listen.Verdict{
						{
							Pkg:     "react",
							Version: "18.0.0",
							Digest:  "555bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.FNI001,
							Message:  "outbound network connection",
							Severity: "medium",
							Metadata: map[string]interface{}{
								"npm_package_name":    "react",
								"npm_package_version": "18.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Pkg:     "react",
							Version: "18.0.0",
							Digest:  "555bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.FNI001,
							Message:  "outbound network connection",
							Severity: "medium",
							Metadata: map[string]interface{}{
								"npm_package_name":    "react",
								"npm_package_version": "17.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Pkg:     "react",
							Version: "18.0.0",
							Digest:  "555bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.STN001,
							Message:  "outbound network connection",
							Severity: "medium",
							Metadata: map[string]interface{}{
								"npm_package_name":    "react",
								"npm_package_version": "17.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Pkg:     "react",
							Version: "18.0.0",
							Digest:  "555bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.STN005,
							Message:  "outbound network connection",
							Severity: "medium",
							Metadata: map[string]interface{}{
								"npm_package_name":    "react",
								"npm_package_version": "17.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
					},
				},
				{
					Name:    "bufferutil",
					Version: strPtr("4.0.7"),
					Verdicts: []listen.Verdict{
						{
							Pkg:     "bufferutil",
							Version: "4.0.7",
							Digest:  "777bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.MDN01,
							Message:  "Empty description",
							Severity: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "bufferutil",
								"npm_package_version": "4.0.7",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Pkg:     "bufferutil",
							Version: "4.0.7",
							Digest:  "777bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.MDN01,
							Message:  "Empty description",
							Severity: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "bufferutil",
								"npm_package_version": "4.0.7",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Pkg:     "bufferutil",
							Version: "4.0.7",
							Digest:  "777bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.MDN01,
							Message:  "Empty description",
							Severity: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "bufferutil",
								"npm_package_version": "4.0.7",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Pkg:     "bufferutil",
							Version: "4.0.7",
							Digest:  "777bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.MDN01,
							Message:  "Empty description",
							Severity: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "farrukh",
								"npm_package_version": "1.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Pkg:     "bufferutil",
							Version: "4.0.7",
							Digest:  "777bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.MDN01,
							Message:  "Empty description",
							Severity: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "mulla",
								"npm_package_version": "1.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Pkg:     "bufferutil",
							Version: "4.0.7",
							Digest:  "777bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.MDN02,
							Message:  "Zero version",
							Severity: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "transitive",
								"npm_package_version": "1.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Pkg:     "bufferutil",
							Version: "4.0.7",
							Digest:  "777bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.MDN02,
							Message:  "Zero version",
							Severity: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "transitive",
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
			snapshot: true,
			name:     "with problems",
			packages: []listen.Package{
				{
					Name:    "bufferutil",
					Version: strPtr("4.0.7"),
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
					Name:    "vu3",
					Version: strPtr("0.0.1"),
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
			snapshot: true,
			name:     "with verdicts and problems",
			packages: []listen.Package{
				{
					Name:    "foo",
					Version: strPtr("1.0.0"),
					Verdicts: []listen.Verdict{
						{
							Pkg:     "foo",
							Version: "1.0.0",
							Digest:  "333bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.FNI001,
							Message:  "outbound network connection",
							Severity: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "foo",
								"npm_package_version": "1.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Pkg:     "foo",
							Version: "1.0.0",
							Digest:  "333bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.FNI001,
							Message:  "outbound network connection",
							Severity: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "foo",
								"npm_package_version": "1.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Pkg:     "foo",
							Version: "1.0.0",
							Digest:  "333bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.FNI001,
							Message:  "outbound network connection",
							Severity: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "foo",
								"npm_package_version": "1.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Pkg:     "foo",
							Version: "1.0.0",
							Digest:  "333bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.FNI002,
							Message:  "write to filesystem",
							Severity: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "foo",
								"npm_package_version": "1.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Pkg:     "foo",
							Version: "1.0.0",
							Digest:  "333bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.FNI002,
							Message:  "write to filesystem",
							Severity: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "bar",
								"npm_package_version": "1.0.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Pkg:     "foo",
							Version: "1.0.0",
							Digest:  "333bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.MDN01,
							Message:  "missing description",
							Severity: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "bar",
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
					Version: strPtr("1.0.0"),
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
					Version: strPtr("1.0.0"),
					Verdicts: []listen.Verdict{
						{
							Pkg:     "baz",
							Version: "1.0.0",
							Digest:  "333bd98592883255fa00de14f1151a917b5d77d5",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Code:     verdictcode.FNI001,
							Message:  "outbound network connection",
							Severity: "high",
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

			if tt.snapshot {
				assert.Nil(t, os.WriteFile(fmt.Sprintf("./testdata/snapshots/%s.md", tt.name), outBuf.Bytes(), 0o644))
			}

			require.Equal(t, tt.expectedOutput, outBuf.Bytes())
		})
	}
}
