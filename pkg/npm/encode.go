package npm

import (
	"encoding/base64"
)

// Encode encodes the receiving packageLockJson instance in a base 64 string.
func (p *packageLockJSON) Encode() string {
	return base64.StdEncoding.EncodeToString(p.bytes)
}
