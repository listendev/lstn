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
package jq

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"

	"github.com/itchyny/gojq"
)

func Compile(expression string) (*gojq.Code, error) {
	query, err := gojq.Parse(expression)
	if err != nil {
		return nil, err
	}

	// Allow access to OS environment variables
	allowAccessToEnv := gojq.WithEnvironLoader(func() []string {
		return os.Environ()
	})

	code, err := gojq.Compile(query, allowAccessToEnv)
	if err != nil {
		return nil, err
	}

	return code, nil
}

func Eval(input io.Reader, output io.Writer, expression string) error {
	code, err := Compile(expression)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(input)
	if err != nil {
		return err
	}

	var resp interface{}
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return err
	}

	// TODO ? use RunWithContext
	iter := code.Run(resp)
	for {
		val, ok := iter.Next()
		if !ok {
			break
		}

		if err, isErr := val.(error); isErr {
			return err
		}

		if text, e := jsonScalarToString(val); e == nil {
			_, err := fmt.Fprintln(output, text)
			if err != nil {
				return err
			}
		} else {
			var jsonFragment []byte
			jsonFragment, err = json.Marshal(val)
			if err != nil {
				return err
			}
			_, err = output.Write(jsonFragment)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(output, "\n")
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func jsonScalarToString(input interface{}) (string, error) {
	switch tt := input.(type) {
	case string:
		return tt, nil
	case float64:
		if math.Trunc(tt) == tt {
			return strconv.FormatFloat(tt, 'f', 0, 64), nil
		} else {
			return strconv.FormatFloat(tt, 'f', 2, 64), nil
		}
	case nil:
		return "", nil
	case bool:
		return fmt.Sprintf("%v", tt), nil
	default:
		return "", fmt.Errorf("cannot convert type to string: %v", tt)
	}
}