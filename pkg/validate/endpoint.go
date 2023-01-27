package validate

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var endpointRe = regexp.MustCompile(`^(http://(localhost|127\.0\.0\.1)(:\d{1,5})?|https://.*\.listen\.dev)`)

func isEndpoint(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		return endpointRe.MatchString(field.String())
	}

	panic(fmt.Sprintf("bad field type: %T", field.Interface()))
}
