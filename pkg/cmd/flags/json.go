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
	"io"

	"github.com/listendev/lstn/pkg/jq"
	"github.com/spf13/cobra"
)

type JsonFlags struct {
	Json bool   `name:"json"`
	JQ   string `name:"jq" validate:"excluded_without=Json,jq"`
}

func (o *JsonFlags) Attach(c *cobra.Command) {
	c.Flags().BoolVar(&o.Json, "json", o.Json, "output the verdicts (if any) in JSON form")
	c.Flags().StringVarP(&o.JQ, "jq", "q", o.JQ, "filter the output using a jq expression")
}

func (o *JsonFlags) JSON() bool {
	return o.Json
}

func (o *JsonFlags) Query() string {
	return o.JQ
}

func (o *JsonFlags) Output(ctx context.Context, input io.Reader, output io.Writer) error {
	if o.JSON() {
		return jq.Eval(ctx, input, output, o.Query())
	}

	return fmt.Errorf("cannot output JSON")
}
