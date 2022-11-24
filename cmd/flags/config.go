package flags

import (
	"fmt"
	"log"
	"reflect"

	"github.com/creasty/defaults"
	"github.com/listendev/lstn/pkg/validate"
)

type ConfigOptions struct {
	LogLevel string `default:"info" name:"log level" flag:"loglevel"`                     // TODO > validator
	Timeout  int    `default:"60" name:"timeout" flag:"timeout" validate:"number,min=30"` // TODO ? make uint
	Endpoint string `default:"http://127.0.0.1:3000" flag:"endpoint" name:"endpoint" validate:"url"`
}

func NewConfigOptions() *ConfigOptions {
	o := &ConfigOptions{}

	if err := defaults.Set(o); err != nil {
		log.Fatal("error setting configuration defaults")
	}

	return o
}

func (o *ConfigOptions) GetField(name string) reflect.Value {
	return reflect.ValueOf(o).Elem().FieldByName(name)
}

func (o *ConfigOptions) Validate() []error {
	if err := validate.Singleton.Struct(o); err != nil {
		all := []error{}
		for _, e := range err.(validate.ValidationErrors) {
			all = append(all, fmt.Errorf(e.Translate(validate.Translator)))
		}

		return all
	}

	return nil
}

func GetConfigFlagsNames() map[string]string {
	ret := make(map[string]string)
	t := reflect.TypeOf(ConfigOptions{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("flag")
		if tag != "" {
			ret[tag] = field.Name
		}
	}

	return ret
}

func GetConfigFlagsDefaults() map[string]string {
	ret := make(map[string]string)
	e := reflect.TypeOf(ConfigOptions{})
	for i := 0; i < e.NumField(); i++ {
		field := e.Field(i)
		tag := field.Tag.Get("default")
		if tag != "" {
			ret[field.Tag.Get("flag")] = tag
		}
	}

	return ret
}
