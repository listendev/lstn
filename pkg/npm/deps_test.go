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

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
)

func TestGetDepsByType(t *testing.T) {
	e1 := map[string]string{
		"react": "18.0.0",
	}

	p1 := &packageJSON{
		Dependencies: e1,
	}

	assert.Equal(t, e1, p1.getDepsByType(Dependencies))
	assert.Equal(t, map[string]string{}, p1.getDepsByType(DevDependencies))

	e2 := map[string]string{
		"react":      "18.0.0",
		"node-fetch": "2.6.9",
	}

	p2 := &packageJSON{
		DevDependencies: e2,
	}

	assert.Equal(t, map[string]string{}, p2.getDepsByType(Dependencies))
	assert.Equal(t, e2, p2.getDepsByType(DevDependencies))
}

func TestGetDepsMapFromDepList(t *testing.T) {
	ret := map[DependencyType]map[string]*semver.Version{}
	getDepsMapFromDepList([]*dep{}, DevDependencies, ret)
	assert.Empty(t, ret)

	depList := []*dep{
		&dep{
			name:    "core-js",
			version: semver.MustParse("3.29.1"),
		},
		&dep{
			name:    "vue",
			version: semver.MustParse("3.2.47"),
		},
		nil,
		&dep{
			name:    "react",
			version: semver.MustParse("18.2.0"),
		},
	}
	getDepsMapFromDepList(depList, Dependencies, ret)
	assert.Len(t, ret, 1)
	assert.Len(t, ret[Dependencies], 3)
	assert.Equal(t, map[DependencyType]map[string]*semver.Version{
		Dependencies: map[string]*semver.Version{
			"react":   semver.MustParse("18.2.0"),
			"vue":     semver.MustParse("3.2.47"),
			"core-js": semver.MustParse("3.29.1"),
		},
	}, ret)
}

func TestGetDepInstance(t *testing.T) {
	v1820, _ := semver.NewConstraint("18.2.0")
	v267, _ := semver.NewConstraint("^2.6.7")

	cases := []struct {
		des string
		pkg string
		ver string
		res *dep
	}{
		{
			des: "exact version",
			pkg: "react",
			ver: "18.2.0",
			res: &dep{
				name:        "react",
				constraints: v1820,
			},
		},
		{
			des: "constraints",
			pkg: "node-fetch",
			ver: "^2.6.7",
			res: &dep{
				name:        "node-fetch",
				constraints: v267,
			},
		},
		// Unsupported
		{
			des: "local path",
			pkg: "dyl",
			ver: "file:.../dyl",
			res: nil,
		},
		{
			des: "URL",
			pkg: "asd",
			ver: "http://asdf.com/asfg.tar.gz",
			res: nil,
		},
		{
			des: "git URLs",
			pkg: "cli",
			ver: "git+ssh://git@github.com:npm/cli#semver:^5.0",
			res: nil,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.des, func(t *testing.T) {
			assert.Equal(t, tc.res, getDepInstance(tc.pkg, tc.ver))
		})
	}
}
