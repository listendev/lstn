# Project Layout

At a high level, these areas make up the `github.com/listendev/lstn` project:

- [`cmd/`](../cmd) - implementation for individual `lstn` commands
- [`pkg/`](../pkg) - most other packages and libraries
- [`docs/`](../docs) - documentation for maintainers and contributors
- [`make/`](../make) - build and release scripts
- [`internal/`](../internal) - packages highly specific to our needs and thus internal
- [`go.mod`](../go.mod) - external Go dependencies for this project, automatically fetched at build time

## Command-line help text

Running `lstn help in` displays the help for a child command or for a topic.

The naming convention for commands is:

```
cmd/<command>/<command.go>
```

In case of nested commands:

```
cmd/<command>/<subcommand>/<subcommand.go>
```

Following the above example, the main implementation for the `lstn in` command (including its help) is in [cmd/in/in.go](../cmd/in/in.go).

Other help topics not specific to any command, for example `lsnt help environment`, are found in [pkg/cmd/help/topics.go](../pkg/cmd/help/topics.go).

TODO: During our release process, these help topics are automatically converted to manual pages.

## How the lstn CLI works

To illustrate how the `lstn` CLI works in its typical mode of operation, let's build the project, run a command,
and talk through which code gets run.

1. `go build -o make/make make/main.go` - Compiles the binary to build `lstn`
2. `make/make lstn` - Makes sure all external Go dependencies are fetched, then compiles the `lstn` binary
3. `./lstn in --json ./ciao` - Runs the newly build `lstn` binary and passes the following flags and arguments to the process:
   - `["in", "--json", "./ciao"]`
4. The `main` package sets up the CLI flags, its subcommands, its context, and dispatches the execution to the "root" command with the `rootCmd.ExecuteContext()` method
5. The [root command](../cmd/root.go) represents the top-level `lstn` command and knows how to dispatch execution to any other gh command nested under it
6. Because of the `in` argument, the execution reaches the `RunE()` function of the `cobra.Command` within [cmd/in/in.go](../cmd/in/in.go)
7. Because of the `--json` flag, the `inOpts.Json` value is set to `true`
8. The logic of the `RunE()` function of the `in` subcommand looks for a `package.json` inside the `./ciao` target directory
9. The `in` logic looks for the `npm` binary, creates a `package-lock.json` on the fly for the `package.json` in the `./ciao` target directory (if any)
10. Then, the logic of the `RunE()` function of the `in` subcommand queries the `npm` registry to collect the sha sums of all the dependencies of the `package-lock.json`
11. Finally, the `in` logic queries the [listen.dev](https://npm.listen.dev/api/analysis) API asking for the analysis verdicts of all the dependencies
12. The response (if any) gets print in JSON form
13. The program execution is now back at `func Boot()` in [root.go](../cmd/root.go)
14. In case there were any Go error as a result of processing the command, the function will abort the process with a non-zero exit status
15. Otherwise, the process ends with status 0 meaning success.

## How to add a new command

1. First, check on our issue tracker to verify that our team had approved the plans for a new command.
2. Create a package for the new command, e.g. for a new command `lstn snitch` create the following directory
   structure: `cmd/snitch/`
3. The new package should expose a `New()` function that accepts a `*context.Context` type and
   returns `(*cobra.Command, error)`.
4. Use the method from the previous step to generate the command and add it to the command tree
   - Typically this means adding it as subcommand of the root command (ie., `cmd/root/root.go`) in its `New()` function.

## How to write tests

This task might be tricky for some edge things we do in `lstn`.

For example, `lstn` looks for the `npm` (and - in the future - for other package managers) binary. Then it uses `npm` to generate a `package-lock.json` file in a temporary directory.

Naturally and generally, none of these things should ever happen for real when running tests, unless you are sure that any filesystem operations are strictly scoped to a location made for and maintained by the test itself.

Where possible we change the active FS during testing. You can take a look at how we do it in [pkg/git/config_test.go](../pkg/git/config_test.go).

Specifically to those bits:

```Go
func TestGitConfig(t *testing.T) {
fs := memfs.New()
activeFS = fs
defer func() { activeFS = defaultFS() }()
// ...
}
```

In other cases, ie. when we look for the `npm` binary using the Go `exec` package, we let the test binary itself pretend to be `npm` and we let the specific test call such a binary.

You can take a look at this approach by inspecting [pkg/npm/main_test.go](../pkg/npm/main_test.go) and [pkg/npm/packagelock_test.go](../pkg/npm/packagelock_test.go).

Finally, `lstn` also performs HTTP calls to our servers.

To avoid actually running things like making real API requests we stub the listen.dev server with the `httptest` Go package (see examples in [pkg/listen/listen_test.go](../pkg/listen/listen_test.go)).

These bits of functionality to help us test `lstn` are placed into the [internal/testing](../internal/testing/) directory.

As a general rule of thumb, to make your code testable write small and isolated pieces of functionality that are designed to be composed together.

Prefer table-driven tests for maintaining variations of different test inputs and expectations when exercising a single piece of functionality.
