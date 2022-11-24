package transform

import (
	"github.com/go-playground/mold/v4"
	"github.com/go-playground/mold/v4/modifiers"
)

// Singleton it the transformers singleton instance.
var Singleton *mold.Transformer

func init() {
	Singleton = modifiers.New()
	Singleton.SetTagName("transform")
}
