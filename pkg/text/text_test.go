package text

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TextSuite struct {
	suite.Suite
}

func TestTextSuiteTestSuite(t *testing.T) {
	suite.Run(t, new(TextSuite))
}

func (suite *TextSuite) TestDedent() {
	cases := []struct {
		s        string
		expected string
	}{}

	for i, tc := range cases {
		require.Equal(suite.T(), tc.expected, Dedent(tc.s), info{i, tc.s})
	}
}

func (suite *TextSuite) TestIndent() {}

func (suite *TextSuite) TestIndentBytes() {}

type info struct {
	i        int
	original interface{}
}

func (i info) String() string {
	return fmt.Sprintf("Index: %d\nOriginal:\n %v\n", i.i, i.original)
}
