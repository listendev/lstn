# lstn cheatsheet

## Global Flags

Every child command inherits the following flags:

```
--config string   config file (default is $HOME/.lstn.yaml)
```

## `lstn completion <bash|fish|powershell|zsh>`

Generate the autocompletion script for the specified shell.

### `lstn completion bash`

Generate the autocompletion script for bash.

#### Flags

```
--no-descriptions   disable completion descriptions
```

### `lstn completion fish [flags]`

Generate the autocompletion script for fish.

#### Flags

```
--no-descriptions   disable completion descriptions
```

### `lstn completion powershell [flags]`

Generate the autocompletion script for powershell.

#### Flags

```
--no-descriptions   disable completion descriptions
```

### `lstn completion zsh [flags]`

Generate the autocompletion script for zsh.

#### Flags

```
--no-descriptions   disable completion descriptions
```

## `lstn config`

Details about the ~/.lstn.yaml config file.

## `lstn environment`

Which environment variables you can use with lstn.

## `lstn exit`

Details about the lstn exit codes.

## `lstn help [command]`

Help about any command.

## `lstn in <path>`

Inspect the verdicts for your dependencies tree.

### Flags

```
-q, --jq string   filter the output using a jq expression
    --json        output the verdicts (if any) in JSON form
```

### Config Flags

```
--endpoint string   the listen.dev endpoint emitting the verdicts (default "https://npm.listen.dev")
--loglevel string   set the logging level (default "info")
--timeout int       set the timeout, in seconds (default 60)
```

### Debug Flags

```
--debug-options   output the options, then exit
```

### Registry Flags

```
--npm-registry string   set a custom NPM registry (default "https://registry.npmjs.org")
```

### Token Flags

```
--gh-token string   set the GitHub token
```

For example:

```bash
lstn in
lstn in .
lstn in /we/snitch
lstn in sub/dir
```

## `lstn manual`

A comprehensive reference of all the lstn commands.

## `lstn scan <path>`

Inspect the verdicts for your direct dependencies.

### Flags

```
-e, --exclude (dep,dev,optional,peer)   sets of dependencies to exclude (in addition to the default) (default [bundle])
-q, --jq string                         filter the output using a jq expression
    --json                              output the verdicts (if any) in JSON form
```

### Config Flags

```
--endpoint string   the listen.dev endpoint emitting the verdicts (default "https://npm.listen.dev")
--loglevel string   set the logging level (default "info")
--timeout int       set the timeout, in seconds (default 60)
```

### Debug Flags

```
--debug-options   output the options, then exit
```

### Registry Flags

```
--npm-registry string   set a custom NPM registry (default "https://registry.npmjs.org")
```

### Token Flags

```
--gh-token string   set the GitHub token
```

For example:

```bash
lstn scan
lstn scan .
lstn scan /we/snitch
lstn scan /we/snitch -e peer
lstn scan /we/snitch -e dev,peer
lstn scan /we/snitch -e dev -e peer
lstn scan sub/dir
```

## `lstn to <name> [[version] [shasum] | [version constraint]]`

Get the verdicts of a package.

### Flags

```
-q, --jq string   filter the output using a jq expression
    --json        output the verdicts (if any) in JSON form
```

### Config Flags

```
--endpoint string   the listen.dev endpoint emitting the verdicts (default "https://npm.listen.dev")
--loglevel string   set the logging level (default "info")
--timeout int       set the timeout, in seconds (default 60)
```

### Debug Flags

```
--debug-options   output the options, then exit
```

### Registry Flags

```
--npm-registry string   set a custom NPM registry (default "https://registry.npmjs.org")
```

### Token Flags

```
--gh-token string   set the GitHub token
```

For example:

```bash
# Get the verdicts for all the chalk versions that listen.dev owns
lstn to chalk
lstn to debug 4.3.4
lstn to react 18.0.0 b468736d1f4a5891f38585ba8e8fb29f91c3cb96

# Get the verdicts for all the existing chalk versions
lstn to chalk "*"
# Get the verdicts for nock versions >= 13.2.0 and < 13.3.0
lstn to nock "~13.2.x"
# Get the verdicts for tap versions >= 16.3.0 and < 16.4.0
lstn to tap "^16.3.0"
# Get the verdicts for prettier versions >= 2.7.0 <= 3.0.0
lstn to prettier ">=2.7.0 <=3.0.0"
```

## `lstn version`

Print out version information.

### Flags

```
--changelog   output the relase notes URL
```

### Debug Flags

```
-v, -- count          increment the verbosity level
    --debug-options   output the options, then exit
```

