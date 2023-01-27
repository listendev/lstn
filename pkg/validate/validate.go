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
package validate

import (
	"reflect"
	"strings"

	en "github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

type ValidationErrors = validator.ValidationErrors

// Singleton is the validator singleton instance.
//
// This way it caches the structs info.
var Singleton *validator.Validate

// Translator is the universal translator for validation errors.
var Translator ut.Translator

func init() {
	Singleton = validator.New()

	// Register a function to get the field name from "name" tags.
	Singleton.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("name"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	Singleton.RegisterValidation("endpoint", isEndpoint)

	eng := en.New()
	Translator, _ = (ut.New(eng, eng)).GetTranslator("en")
	en_translations.RegisterDefaultTranslations(Singleton, Translator)

	Singleton.RegisterTranslation(
		"endpoint",
		Translator,
		func(ut ut.Translator) error {
			return ut.Add("endpoint", "{0} must be a valid listen.dev endpoint", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("endpoint", fe.Field())
			return t
		},
	)
}
