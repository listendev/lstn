# Project Layout

At a high level, these areas make up the `github.com/listendev/lstn` project:

- [`cmd/`](../cmd) - implementation for individual `lstn` commands
- [`pkg/`](../pkg) - most other packages and libraries
- [`docs/`](../docs) - documentation for maintainers and contributors
- [`scripts/`](../scripts) - build and release scripts
- [`internal/`](../internal) - packages highly specific to our needs and thus internal
- [`go.mod`](../go.mod) - external Go dependencies for this project, automatically fetched at build time

## Command-line help text

_TODO_

## How the lstn CLI works

To illustrate how the `lstn` CLI works in its typical mode of operation, let's build the project, run a command,
and talk through which code gets run.

_TODO_

## How to add a new command

1. First, check on our issue tracker to verify that our team had approved the plans for a new command.
2. Create a package for the new command, e.g. for a new command `lstn snitch` create the following directory
   structure: `cmd/snitch/`
3. The new package should expose a `New()` function that accepts a `*context.Context` type and
   returns `(*cobra.Command, error)`.
4. Use the method from the previous step to generate the command and add it to the command tree
   - Typically this means adding it as subcommand of the root command (ie., `cmd/root.go`) in its `Boot()` function.

## How to write tests

This task might be tricky.

_TODO_
