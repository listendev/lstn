package context

import (
	"context"
	"errors"
	"fmt"
	"net"
)

func ContextError(ctx context.Context, input error) error {
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
	if err := ContextError(ctx, input); err != nil {
		return err
	}

	return input
}

func OutputErrorf(ctx context.Context, input error, format string, a ...any) error {
	if err := ContextError(ctx, input); err != nil {
		return err
	}

	return fmt.Errorf(format, a...)
}
