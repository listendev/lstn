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
	"os"
	"testing"

	"github.com/listendev/lstn/pkg/listen"
	"github.com/stretchr/testify/require"
)

func testdataFileToBytes(t *testing.T, dataFile string) []byte {
	b, err := os.ReadFile(dataFile)
	if err != nil {
		t.Fatal(err)
	}

	return b
}

func TestRenderSingleProblemsPackage(t *testing.T) {
	tests := []struct {
		name           string
		p              listen.Package
		expectedOutput []byte
		wantErr        bool
	}{
		{
			name: "no problems",
			p: listen.Package{
				Name:     "foo",
				Version:  strPtr("1.0.0"),
				Problems: []listen.Problem{},
			},
			expectedOutput: testdataFileToBytes(t, "testdata/single_problems_no_problems.md"),
		},
		{
			name: "with problems",
			p: listen.Package{
				Name:    "foo",
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
			expectedOutput: testdataFileToBytes(t, "testdata/single_problems_with_problems.md"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := &bytes.Buffer{}
			err := RenderSingleProblemsPackage(outBuf, tt.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderSingleProblemsPackage() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			require.Equal(t, tt.expectedOutput, outBuf.Bytes())
		})
	}
}
