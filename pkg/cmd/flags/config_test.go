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
