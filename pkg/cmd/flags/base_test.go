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
package flags

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FlagsBaseSuite struct {
	suite.Suite
}

func TestFlagsBaseSuite(t *testing.T) {
	suite.Run(t, new(FlagsBaseSuite))
}

func (suite *FlagsBaseSuite) TestValidate() {
	cases := []struct {
		desc        string
		o           Options
		expectedStr []string
	}{
		{
			"empty config options",
			&ConfigOptions{},
			[]string{"timeout must be 30 or greater", "endpoint must be a valid URL"},
		},
		{
			"invalid timeout",
			&ConfigOptions{Timeout: 29, Endpoint: "http://127.0.0.1:3000"},
			[]string{"timeout must be 30 or greater"},
		},
		{
			"invalid endpoint",
			&ConfigOptions{Timeout: 31, Endpoint: "http://invalid.endpoint"},
			[]string{"endpoint must be a valid listen.dev endpoint"},
		},
		{
			"valid config options",
			&ConfigOptions{Timeout: 31, Endpoint: "http://127.0.0.1:3000"},
			[]string{},
		},
	}

	for _, tc := range cases {
		bo := new(baseOptions)

		suite.T().Run(tc.desc, func(t *testing.T) {
			actual := bo.Validate(tc.o)
			assert.Equal(suite.T(), len(tc.expectedStr), len(actual))
			for _, a := range actual {
				assert.Contains(suite.T(), tc.expectedStr, a.Error())
			}
		})
	}
}

func (suite *FlagsBaseSuite) TestTransform() {
	cases := []struct {
		desc     string
		o        Options
		expected error
	}{
		{
			"empty config options",
			&ConfigOptions{},
			nil,
		},
	}

	ctx := context.Background()
	for _, tc := range cases {
		bo := new(baseOptions)
		suite.T().Run(tc.desc, func(t *testing.T) {
			assert.Equal(t, tc.expected, bo.Transform(ctx, tc.o))
		})
	}
}
