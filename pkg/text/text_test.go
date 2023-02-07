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

func (suite *TextSuite) TestIndentDedent() {
	cases := []struct {
		b string
	}{
		{""},
		{"line1"},
		{"line1\nline2\n"},
		{"line1\nline2"},
		{"\tline1\n\tline2"},
	}

	for _, tc := range cases {
		suite.T().Run(tc.b, func(t *testing.T) {
			assert.Equal(t, tc.b, Dedent(Indent(tc.b, "  ")))
		})
	}
}

func (suite *TextSuite) TestDedent() {
	cases := []struct {
		s        string
		expected string
	}{
		{"", ""},
		{"line1\nline2", "line1\nline2"},
		{" line1\n  line2\n\n line4", "line1\n line2\n\nline4"},
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
