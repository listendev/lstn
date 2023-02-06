/*
Copyright © 2022 The listen.dev team <engineering@garnet.ai>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type VersionSuite struct {
	suite.Suite
}

func TestVersionSuite(t *testing.T) {
	suite.Run(t, new(VersionSuite))
}

func (suite *VersionSuite) TestChangeLog() {
	cases := []struct {
		tag string
		url string
	}{
		{
			tag: "0.3.2",
			url: "https://github.com/listendev/lstn/releases/tag/v0.3.2",
		},
		{
			tag: "v0.3.2",
			url: "https://github.com/listendev/lstn/releases/tag/v0.3.2",
		},
		{
			tag: "v0.3.2-alpha.1",
			url: "https://github.com/listendev/lstn/releases/tag/v0.3.2-alpha.1",
		},
		{
			tag: "150d3f96.20230130",
			url: "",
		},
	}

	for _, tc := range cases {
		suite.T().Run(tc.tag, func(t *testing.T) {
			res, _ := Changelog(tc.tag)
			assert.Equal(t, res, tc.url)
		})
	}
}
