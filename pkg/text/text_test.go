package text

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TextSuite struct {
	suite.Suite
}

func TestTextSuite(t *testing.T) {
	suite.Run(t, new(TextSuite))
}

func (suite *TextSuite) TestDedent() {
	cases := []struct {
		s        string
		expected string
	}{
		{"", ""},
		{"line1\nline2", "line1\nline2"},
		// {"line1\n\tline2", "line1\nline2"}, // FIXME(fra): is implementation working as supposed?
	}

	for _, tc := range cases {
		suite.T().Run(tc.s, func(t *testing.T) {
			assert.Equal(t, tc.expected, Dedent(tc.s))
		})
	}
}

func (suite *TextSuite) TestIndent() {
	cases := []struct {
		b        string
		prefix   string
		expected string
	}{
		{"", "  ", ""},
		{"line1", "  ", "  line1"},
		{"line1\nline2\n", "\t", "\tline1\n\tline2\n"},
		{"line1\nline2", "  ", "  line1\n  line2"},
		{"\tline1\n\tline2", "  ", "  \tline1\n  \tline2"},
	}

	for _, tc := range cases {
		suite.T().Run(tc.b, func(t *testing.T) {
			assert.Equal(t, tc.expected, Indent(tc.b, tc.prefix))
		})
	}
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

	for _, tc := range cases {
		suite.T().Run(string(tc.b), func(t *testing.T) {
			assert.Equal(t, tc.expected, IndentBytes(tc.b, tc.prefix))
		})
	}
}
