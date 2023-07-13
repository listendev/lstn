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

func TestRenderSingleVerdictsPackage(t *testing.T) {
	tests := []struct {
		name           string
		p              listen.Package
		expectedOutput []byte
		wantErr        bool
	}{
		{
			name: "no verdicts",
			p: listen.Package{
				Name:     "foo",
				Version:  strPtr("1.0.0"),
				Verdicts: []listen.Verdict{},
			},
			expectedOutput: testdataFileToBytes(t, "testdata/single_verdicts_no_verdicts.md"),
		},
		{
			name: "one verdict",
			p: listen.Package{
				Name:    "foo",
				Version: strPtr("1.0.0"),
				Verdicts: []listen.Verdict{
					{
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
				},
			},
			expectedOutput: testdataFileToBytes(t, "testdata/single_verdicts_one_verdict.md"),
		},
		{
			name: "verdicts with ai_ctx",
			p: listen.Package{
				Name:    "foo",
				Version: strPtr("1.0.0"),
				Verdicts: []listen.Verdict{
					{
						Message:  "unexpected outbound connection destination",
						Severity: "high",
						Metadata: map[string]interface{}{
							"npm_package_name":    "foo",
							"npm_package_version": "1.0.0",
							"parent_name":         "node",
							"executable_path":     "/bin/sh",
							"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
							"server_ip":           "",
							"ai_ctx": map[string]interface{}{
								"actions": []string{
									"Do not continue to install this dependency as it could potentially harm your system.",
									"Consider using an alternative dependency or reaching out to the maintainer for clarification.",
									"Monitor your system for any suspicious activity.",
								},
								"concern": 1,
								"context": "The IP address 43.131.244.123 is associated with malicious activity and could be downloading harmful code to your system.",
							},
						},
					},
				},
			},
			expectedOutput: testdataFileToBytes(t, "testdata/single_verdicts_with_aictx.md"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := &bytes.Buffer{}
			err := RenderSingleVerdictsPackage(outBuf, tt.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderSingleVerdictsPackage() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			require.Equal(t, tt.expectedOutput, outBuf.Bytes())
		})
	}
}
