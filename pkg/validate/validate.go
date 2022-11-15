package validate

import (
	"github.com/go-playground/validator/v10"
)

// Singleton is the validator singleton instance.
//
// This way it caches the structs info.
var Singleton *validator.Validate

func init() {
	Singleton = validator.New()
}
