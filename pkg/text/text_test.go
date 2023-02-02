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
		require.Equal(suite.T(), tc.expected, Dedent(tc.s), fmt.Sprintf("Index: %d\nOriginal:\n %v\n", i, tc.s))
	}
}

func (suite *TextSuite) TestIndent() {
}

func (suite *TextSuite) TestIndentBytes() {
	cases := []struct {
		b        []byte
		prefix   []byte
		expected []byte
	}{
		{[]byte(""), []byte("  "), []byte("")},
		{[]byte("line1"), []byte("  "), []byte("  line1")},
		{[]byte("line1\nline2\n"), []byte("\t"), []byte("\tline1\n\tline2\n")},
		{[]byte("line1\nline2"), []byte("  "), []byte("  line1\n  line2")},
		{[]byte("\tline1\n\tline2"), []byte("  "), []byte("  \tline1\n  \tline2")},
	}

	for i, tc := range cases {
		require.Equal(suite.T(), tc.expected, IndentBytes(tc.b, tc.prefix), fmt.Sprintf("Index: %d\n", i))
	}
}
