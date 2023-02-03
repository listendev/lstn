package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FlagsConfigSuite struct {
	suite.Suite
}

func TestFlagsConfigSuite(t *testing.T) {
	suite.Run(t, new(FlagsConfigSuite))
}

func (suite *FlagsConfigSuite) TestNewConfigOptions() {
	i, err := NewConfigOptions()
	assert.NoError(suite.T(), err)
	assert.IsType(suite.T(), &ConfigOptions{}, i)
}

func (suite *FlagsConfigSuite) TestGetConfigFlagsNames() {
	m := GetConfigFlagsNames()
	assert.Equal(suite.T(), 3, len(m))

	expected := make(map[string]string)
	expected["loglevel"] = "LogLevel"
	expected["endpoint"] = "Endpoint"
	expected["timeout"] = "Timeout"

	for k, v := range m {
		e, ok := expected[k]
		assert.True(suite.T(), ok)
		assert.Equal(suite.T(), e, v)
	}
}

func (suite *FlagsConfigSuite) TestGetConfigFlagsDefaults() {
	m := GetConfigFlagsDefaults()
	assert.Equal(suite.T(), 3, len(m))

	expected := make(map[string]string)
	expected["endpoint"] = "http://127.0.0.1:3000"
	expected["loglevel"] = "info"
	expected["timeout"] = "60"

	for k, v := range m {
		e, ok := expected[k]
		assert.True(suite.T(), ok)
		assert.Equal(suite.T(), e, v)
	}
}
