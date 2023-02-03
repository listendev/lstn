package npm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type NpmEncodeSuite struct {
	suite.Suite
}

func TestNpmEncodeSuite(t *testing.T) {
	suite.Run(t, new(NpmEncodeSuite))
}

func (suite *NpmEncodeSuite) TestEncode() {
	i := &packageLockJSON{bytes: []byte("test")}
	assert.Equal(suite.T(), "dGVzdA==", i.Encode())
}
