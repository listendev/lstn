// SPDX-License-Identifier: Apache-2.0
//
// Copyright © 2023 The listen.dev team <engineering@garnet.ai>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package flags

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/XANi/goneric"
	"github.com/listendev/lstn/pkg/cmd/flagusages"
	t "github.com/listendev/lstn/pkg/transform"
	v "github.com/listendev/lstn/pkg/validate"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

// EnvSeparator is the separator between the env variable prefix and the global flag name.
var EnvSeparator = "_"

// EnvReplacer is the string replacer that defines the transformation for flag names into environment variable names.
var EnvReplacer = strings.NewReplacer("-", EnvSeparator, ".", EnvSeparator)

// EnvPrefix is the prefix of the env variables corresponding to the global flags.
var EnvPrefix = "lstn"

func Validate(o interface{}) []error {
	if err := v.Singleton.Struct(o); err != nil {
		all := []error{}
		for _, e := range err.(v.ValidationErrors) {
			all = append(all, fmt.Errorf(e.Translate(v.Translator)))
		}

		return all
	}

	return nil
}

func Transform(ctx context.Context, o interface{}) error {
	if err := t.Singleton.Struct(ctx, o); err != nil {
		return fmt.Errorf("couldn't transform configuration options properly")
	}

	return nil
}

func AsJSON(o interface{}) string {
	data, _ := json.MarshalIndent(o, "", "\t")

	var iface interface{}
	//nolint:errcheck // no need to check the error
	json.Unmarshal(data, &iface)

	data, _ = json.MarshalIndent(iface, "", "\t")

	return string(data)
}

func Define(c *cobra.Command, o interface{}, startingGroup string, exclusions []string) {
	ignore := goneric.SliceMapSetFunc(func(e string) string {
		return strings.TrimPrefix(e, "--")
	}, exclusions)
	val := getValue(o)

	// Check if the current value has a ConfigFlags field
	fld := getValue(o).FieldByName("ConfigFlags")
	if fld.IsValid() {
		// Call o.ConfigFlags.Define(c)
		if fldPtr := getValuePtr(o); fldPtr.IsValid() {
			fldPtr.MethodByName("Define").Call([]reflect.Value{
				getValuePtr(c),
				getValue(maps.Keys(ignore)),
			})
		}
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		f := val.Type().Field(i)
		short := f.Tag.Get("shorthand")
		tag := f.Tag.Get("flag")
		// Do not define flags which tag is in the ignore list
		if _, ok := ignore[tag]; ok {
			continue
		}

		descr := f.Tag.Get("desc")
		group := f.Tag.Get("flagset")
		if startingGroup != "" {
			group = startingGroup
		}

		// TODO: complete type switch as needed
		switch f.Type.Kind() {
		case reflect.Struct:
			// NOTE > field.Interface() doesn't work because it actually returns a copy of the object wrapping the interface
			Define(c, field.Addr().Interface(), group, exclusions)

			continue

		case reflect.Bool:
			val := field.Interface().(bool)
			ref := (*bool)(unsafe.Pointer(field.UnsafeAddr()))
			c.Flags().BoolVarP(ref, tag, short, val, descr)

		case reflect.String:
			val := field.Interface().(string)
			ref := (*string)(unsafe.Pointer(field.UnsafeAddr()))
			c.Flags().StringVarP(ref, tag, short, val, descr)

		case reflect.Int:
			val := field.Interface().(int)
			ref := (*int)(unsafe.Pointer(field.UnsafeAddr()))
			if f.Tag.Get("type") == "count" {
				c.Flags().CountVarP(ref, tag, short, descr)

				continue
			}
			c.Flags().IntVarP(ref, tag, short, val, descr)

		case reflect.Slice:
			if f.Type.Elem().Kind() == reflect.String {
				val := field.Interface().([]string)
				ref := (*[]string)(unsafe.Pointer(field.UnsafeAddr()))
				c.Flags().StringSliceVarP(ref, tag, short, val, descr)
			}

		default:
			continue
		}

		// Set the group annotation on the current flag
		if group != "" {
			_ = c.Flags().SetAnnotation(tag, flagusages.FlagGroupAnnotation, []string{group})
		}
	}
}

func getNames(val reflect.Value) map[string]string {
	ret := make(map[string]string)

	if val.Kind() != reflect.Struct {
		return ret
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		tag := field.Tag.Get("flag")

		if val.Field(i).Kind() == reflect.Struct {
			for k, v := range getNames(val.Field(i)) {
				ret[k] = fmt.Sprintf("%s.%s", field.Name, v)
			}
		}

		if tag != "" {
			ret[tag] = field.Name
		}
	}

	return ret
}

func GetNames(o interface{}) map[string]string {
	val := getValue(o)

	return getNames(val)
}

func getDefaults(val reflect.Value) map[string]string {
	ret := make(map[string]string)

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		tag := field.Tag.Get("default")

		if val.Field(i).Kind() == reflect.Struct {
			for k, v := range getDefaults(val.Field(i)) {
				ret[k] = v
			}
		}

		if tag != "" {
			ret[field.Tag.Get("flag")] = tag
		}
	}

	return ret
}

func GetDefaults(o interface{}) map[string]string {
	val := getValue(o)

	return getDefaults(val)
}

func getTypes(val reflect.Value) map[string]reflect.Type {
	ret := make(map[string]reflect.Type)

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		tag := field.Tag.Get("flag")

		if val.Field(i).Kind() == reflect.Struct {
			for k, v := range getTypes(val.Field(i)) {
				ret[k] = v
			}
		}

		if tag != "" {
			ret[field.Tag.Get("flag")] = field.Type
		}
	}

	return ret
}

func GetTypes(o interface{}) map[string]reflect.Type {
	val := getValue(o)

	return getTypes(val)
}

func getField(val reflect.Value, name string) reflect.Value {
	parts := strings.Split(name, ".")
	ret := val.FieldByName(parts[0])
	if ret.Kind() != reflect.Struct {
		return ret
	}

	return getField(ret, strings.Join(parts[1:], "."))
}

func GetField(o interface{}, name string) reflect.Value {
	val := getValue(o)

	return getField(val, name)
}

func getValue(o interface{}) reflect.Value {
	var ptr reflect.Value
	var val reflect.Value

	val = reflect.ValueOf(o)
	// When we get a pointer, we want to get the value pointed to.
	// Otherwise, we need to get a pointer to the value we got.
	if val.Type().Kind() == reflect.Ptr {
		ptr = val
		val = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(o))
		temp := ptr.Elem()
		temp.Set(val)
		val = temp
	}

	return val
}

func getValuePtr(o interface{}) reflect.Value {
	val := reflect.ValueOf(o)
	// When we get a pointer, we want to get the value pointed to.
	// Otherwise, we need to get a pointer to the value we got.
	if val.Type().Kind() == reflect.Ptr {
		return val
	}
	// ptr := reflect.New(reflect.TypeOf(o))
	// temp := ptr.Elem()
	// temp.Set(val)

	return reflect.New(reflect.TypeOf(o))
}
