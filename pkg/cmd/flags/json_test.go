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
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FlagsJSONSuite struct {
	suite.Suite
}

func TestFlagsJSONSuite(t *testing.T) {
	suite.Run(t, new(FlagsJSONSuite))
}

func (suite *FlagsJSONSuite) TestJSON() {
	i := &JSONFlags{}
	assert.False(suite.T(), i.IsJSON())
	i.JSON = true
	assert.True(suite.T(), i.IsJSON())
}

func (suite *FlagsJSONSuite) TestQuery() {
	i := &JSONFlags{JQ: "."}
	assert.Equal(suite.T(), ".", i.GetQuery())
}

func (suite *FlagsJSONSuite) TestOutput() {
	suite.T().Run("Failure", func(t *testing.T) {
		i := &JSONFlags{JSON: false, JQ: "."}
		input := bytes.NewReader([]byte("{\"key\":\"value\"}"))
		var output bytes.Buffer
		assert.EqualError(suite.T(), i.GetOutput(context.Background(), input, &output), "cannot output JSON")
	})
	suite.T().Run("Success", func(t *testing.T) {
		t.Run("QueryGetAll", func(t *testing.T) {
			i := &JSONFlags{JSON: true, JQ: "."}
			input := bytes.NewReader([]byte("{\"key\":\"value\"}"))
			var output bytes.Buffer
			assert.NoError(suite.T(), i.GetOutput(context.Background(), input, &output))
			assert.Equal(suite.T(), "{\"key\":\"value\"}\n", output.String())
		})
		t.Run("QueryGetValue", func(t *testing.T) {
			i := &JSONFlags{JSON: true, JQ: ".key"}
			input := bytes.NewReader([]byte("{\"key\":\"value\"}"))
			var output bytes.Buffer
			assert.NoError(suite.T(), i.GetOutput(context.Background(), input, &output))
			assert.Equal(suite.T(), "value\n", output.String())
		})
	})
}
