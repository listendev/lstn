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
package flags

import (
	"context"
	"fmt"

	"github.com/listendev/lstn/pkg/transform"
	"github.com/listendev/lstn/pkg/validate"
)

type baseOptions struct {
}

func (b *baseOptions) Validate(o Options) []error {
	if err := validate.Singleton.Struct(o); err != nil {
		all := []error{}
		for _, e := range err.(validate.ValidationErrors) {
			all = append(all, fmt.Errorf(e.Translate(validate.Translator)))
		}

		return all
	}

	return nil
}

func (b *baseOptions) Transform(ctx context.Context, o Options) error {
	if err := transform.Singleton.Struct(ctx, o); err != nil {
		return fmt.Errorf("couldn't transform configuration options properly")
	}
	return nil
}
