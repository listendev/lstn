package flags

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		o           Options
		expectedStr []string
	}{
		{
			&ConfigOptions{},
			[]string{"timeout must be 30 or greater", "endpoint must be a valid URL"},
		},
		{
			&ConfigOptions{Timeout: 29, Endpoint: "http://127.0.0.1:3000"},
			[]string{"timeout must be 30 or greater"},
		},
		{
			&ConfigOptions{Timeout: 31, Endpoint: "http://invalid.endpoint"},
			[]string{"endpoint must be a valid listen.dev endpoint"},
		},
		{
			&ConfigOptions{Timeout: 31, Endpoint: "http://127.0.0.1:3000"},
			[]string{},
		},
	}

	for _, tc := range cases {
		bo := new(baseOptions)
		actual := bo.Validate(tc.o)
		assert.Equal(suite.T(), len(tc.expectedStr), len(actual))
		for _, a := range actual {
			assert.Contains(suite.T(), tc.expectedStr, a.Error())
		}
	}
}

func (suite *FlagsBaseSuite) TestTransform() {
	cases := []struct {
		o        Options
		expected error
	}{
		{
			&ConfigOptions{},
			nil,
		},
		// There are no other use cases to test. The only way for the underlying [(mold.Transformer).Struct](https://pkg.go.dev/github.com/go-playground/mold/v4#Transformer.Struct)
		// to fail is when it is provided with an input that cannot be transformed into a string typo. It can't happen as it is shielded by the struct ConfigOptions.
	}

	ctx := context.Background()
	for _, tc := range cases {
		bo := new(baseOptions)
		require.Equal(suite.T(), tc.expected, bo.Transform(ctx, tc.o))
	}
}
