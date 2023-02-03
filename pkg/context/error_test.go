package context

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ErrorSuite struct {
	suite.Suite
}

func TestErrorSuite(t *testing.T) {
	suite.Run(t, new(ErrorSuite))
}

func (suite *ErrorSuite) TestContextError() {
	cases := []struct {
		name       string
		ctxFactory contextFactory
		input      error
		expected   error
	}{
		{
			name: "Input error is context.Canceled",
			ctxFactory: func() context.Context {
				return context.Background()
			},
			input:    context.Canceled,
			expected: context.Canceled,
		},
		{
			name: "Input error is a timeout error",
			ctxFactory: func() context.Context {
				return context.Background()
			},
			input:    &timeoutError{true},
			expected: context.DeadlineExceeded,
		},
		{
			name: "Context error is context.DeadlineExceeded",
			ctxFactory: func() context.Context {
				ctx, _ := context.WithDeadline(context.Background(), time.Now())
				return ctx
			},
			input:    errors.New("some error"),
			expected: context.DeadlineExceeded,
		},
		{
			name: "Context error is context.Canceled",
			ctxFactory: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				return ctx
			},
			input:    errors.New("some error"),
			expected: context.Canceled,
		},
		{
			name: "No error",
			ctxFactory: func() context.Context {
				return context.Background()
			},
			input:    errors.New("some error"),
			expected: nil,
		},
	}

	for _, tc := range cases {
		suite.T().Run(tc.name, func(t *testing.T) {
			assert.ErrorIs(t, tc.expected, ContextError(tc.ctxFactory(), tc.input))
		})
	}
}

// Utils

// Implements net.Error interface
type timeoutError struct {
	timeout bool
}

func (e *timeoutError) Error() string {
	return "timeout error"
}

func (e *timeoutError) Timeout() bool {
	return e.timeout
}

func (e *timeoutError) Temporary() bool {
	return e.timeout
}

type contextFactory func() context.Context
