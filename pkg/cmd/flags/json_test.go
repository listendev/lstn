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
	i := &JsonFlags{}
	assert.False(suite.T(), i.JSON())
	i.Json = true
	assert.True(suite.T(), i.JSON())
}

func (suite *FlagsJSONSuite) TestQuery() {
	i := &JsonFlags{JQ: "."}
	assert.Equal(suite.T(), ".", i.Query())
}

func (suite *FlagsJSONSuite) TestOutput() {
	suite.T().Run("Failure", func(t *testing.T) {
		i := &JsonFlags{Json: false, JQ: "."}
		input := bytes.NewReader([]byte("{\"key\":\"value\"}"))
		var output bytes.Buffer
		assert.EqualError(suite.T(), i.Output(context.Background(), input, &output), "cannot output JSON")
	})
	suite.T().Run("Success", func(t *testing.T) {
		t.Run("QueryGetAll", func(t *testing.T) {
			i := &JsonFlags{Json: true, JQ: "."}
			input := bytes.NewReader([]byte("{\"key\":\"value\"}"))
			var output bytes.Buffer
			assert.NoError(suite.T(), i.Output(context.Background(), input, &output))
			assert.Equal(suite.T(), "{\"key\":\"value\"}\n", output.String())
		})
		t.Run("QueryGetValue", func(t *testing.T) {
			i := &JsonFlags{Json: true, JQ: ".key"}
			input := bytes.NewReader([]byte("{\"key\":\"value\"}"))
			var output bytes.Buffer
			assert.NoError(suite.T(), i.Output(context.Background(), input, &output))
			assert.Equal(suite.T(), "value\n", output.String())
		})
	})
}
