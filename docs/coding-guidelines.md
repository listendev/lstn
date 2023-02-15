# Code Guidelines

Unless specified otherwise, we should follow the guidelines outlined in
[Effective Go](https://golang.org/doc/effective_go) and
[Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

We are using [golangci-lint](https://golangci-lint.run/) to ensure the code conforms to our code guidelines. Part of
the guidelines are outlined below.

### General

General pointers around writing code for the `lstn` CLI:

- Functions should generally (except when not possible otherwise) **return a specific type instead of an interface**.
  - The caller should have the ability to know the exact type of the returned object and not only the interface it fulfills.
- Try to cleanly **separate concerns** and do not let implementation details spill to the caller code.
- When naming types, always keep in mind that the type will be used with the package name. We should **write code that does not stutter** (e.g. use `cmd.New` instead of `cmd.NewCmd`).
- **Avoid global state**, rather pass things explicitly between structs and functions.

### Packages

We generally follow the [Style guideline for Go packages](https://rakyll.org/style-packages/). Here is a short summary
of those guidelines:

- Organize code into packages by their functional responsibility.
- Package names should be lowercase only (don't use snake_case or camelCase).
- Package names should be short, but should be unique and representative.
  - Avoid overly broad package names like `common` and `util`.
- Use singular package names (e.g. `transform` instead of `transforms`).
- Use `doc.go` to document a package.

Additionally, we encourage the usage of the `internal` package to hide complex internal implementation details of a
package and enforce a better separation of concerns between packages.

### Logging

TODO: ...

### Error Handling

- Any error needs to be handled by either logging the error and recovering from it, or wrapping and returning it to the caller.
  - We should never both log and return the error.
- It's preferred to have a single file called `errors.go` per package which contains all the error variables from that package.

### Testing

We have 3 test suites:

- Unit tests are normal Go tests that don't need any external services to run and mock internal dependencies.
  - They are located in files named `${FILE}_test.go`, where `${FILE}.go` contains the code that's being tested.
- TODO: Integration tests are also written in Go, but can expect external dependencies.
  - These tests should mostly be contained to code that directly communicates with those external dependencies.
  - Integration tests are located in files named `${FILE}_integration_test.go`, where `${FILE}.go` contains the code that's being tested.
  - Files that contain integration tests must contain the build tag `//go:build integration`.
- TODO: End-to-end tests are tests that spin up...

### Documentation

We should write in-line documentation that can be read by `godoc`. This means that exported types, functions and
variables need to have a preceding comment that starts with the name of the expression and end with a dot.
