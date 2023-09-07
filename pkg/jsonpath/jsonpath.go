package jsonpath

import (
	"fmt"
)

func Make(scriptExpression string) string {
	if scriptExpression != "" {
		return fmt.Sprintf("$[?(%s)]", scriptExpression)
	}

	return scriptExpression
}
