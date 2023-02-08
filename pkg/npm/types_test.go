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
package npm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageStruct(t *testing.T) {
	tests := []struct {
		desc string
		pack Package
		okay bool
	}{
		{
			"both-empty",
			Package{Version: "", Shasum: ""},
			false,
		},
		{
			"version-empty",
			Package{Version: "", Shasum: "b468736d1f4a5891f38585ba8e8fb29f91c3cb96"},
			false,
		},
		{
			"shasum-empty",
			Package{Version: "18.0.0", Shasum: ""},
			false,
		},
		{
			"version-invalid",
			Package{Version: "invalid", Shasum: "b468736d1f4a5891f38585ba8e8fb29f91c3cb96"},
			false,
		},
		{
			"shasum-invalid",
			Package{Version: "18.0.0", Shasum: "b468736d1f4a5891f38585ba8e8fb29f91c3cb9"},
			false,
		},
		{
			"valid",
			Package{Version: "18.0.0", Shasum: "b468736d1f4a5891f38585ba8e8fb29f91c3cb96"},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			assert.Equal(t, tc.pack.Ok(), tc.okay)
		})
	}
}

func TestPackagesType(t *testing.T) {
	tests := []struct {
		desc string
		pkgs Packages
		okay bool
	}{
		{
			"empty",
			Packages{},
			false,
		},
		{
			"nil",
			nil,
			false,
		},
		{
			"invalid-key",
			Packages{
				"": Package{Version: "18.0.0", Shasum: "b468736d1f4a5891f38585ba8e8fb29f91c3cb96"},
			},
			false,
		},
		{
			"invalid-package-version",
			Packages{
				"react": Package{Version: "", Shasum: "b468736d1f4a5891f38585ba8e8fb29f91c3cb96"},
			},
			false,
		},
		{
			"invalid-package-shasum",
			Packages{
				"": Package{Version: "18.0.0", Shasum: ""},
			},
			false,
		},
		{
			"valid",
			Packages{
				"react": Package{Version: "18.0.0", Shasum: "b468736d1f4a5891f38585ba8e8fb29f91c3cb96"},
			},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			assert.Equal(t, tc.pkgs.Ok(), tc.okay)
		})
	}
}
