package listen

import (
	"encoding/json"

	"github.com/listendev/lstn/pkg/npm"
)

type Request struct {
	PackageLockJSON npm.PackageLockJSON `json:"package-lock.json"`
	Packages        npm.Packages        `json:"packages"`
	Context         string              `json:"context"` // TODO > define
}

// MarshalJSON is a custom marshaler that encodes the
// content of the package lock in the receiving Request.
func (req *Request) MarshalJSON() ([]byte, error) {
	type RequestAlias Request

	return json.Marshal(&struct {
		PackageLockJSON string `json:"package-lock.json"`
		*RequestAlias
	}{
		PackageLockJSON: req.PackageLockJSON.Encode(),
		RequestAlias:    (*RequestAlias)(req),
	})
}

type Verdict struct {
	Name    string
	Message string
	Data    map[string]string
}

type Message struct {
	Reason  uint8
	Message string
}

type Package struct {
	npm.Package
	Name     string
	Results  bool
	Verdicts []Verdict `json:"verdicts,omitempty"`
	Messages []Message `json:"messages,omitempty"`
}

type Response []Package

type Error struct {
	Message   string
	RequestID string `json:"request_id"`
	Reason    struct {
		Message string
		Reason  string
		Name    string
		npm.Package
	}
}
