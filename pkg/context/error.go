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
package context

import (
	"context"
	"errors"
	"fmt"
	"net"
)

func Error(ctx context.Context, input error) error {
	if errors.Is(input, context.Canceled) {
		return context.Canceled
	}
	if e, ok := input.(net.Error); ok && e.Timeout() {
		return context.DeadlineExceeded
	}
	if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
		return ctx.Err()
	}

	return nil
}

func OutputError(ctx context.Context, input error) error {
	if err := Error(ctx, input); err != nil {
		return err
	}

	return input
}

func OutputErrorf(ctx context.Context, input error, format string, a ...any) error {
	if err := Error(ctx, input); err != nil {
		return err
	}

	return fmt.Errorf(format, a...)
}
