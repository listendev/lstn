# lstn cheatsheet

## Global Flags

Every child command inherits the following flags:

```
--config string   config file (default is $HOME/.lstn.yaml)
```

## `lstn ci`

Listen in on what your CI does.

### `lstn ci enable`

Enable the CI eavesdropping.

#### Flags

```
--dir string   the directory where the jibril binary is
```

#### Config Flags

```
--core-api string   set a custom Core API base URL (default "https://core.listen.dev")
--loglevel string   set the logging level (default "info")
--timeout int       set the timeout, in seconds (default 60)
```

#### Debug Flags

```
--debug-options   output the options, then exit
```

#### Token Flags

```
--gh-token string    set the GitHub token
--jwt-token string   set the listen.dev auth token
```

### `lstn ci report`

Report the most critical findings into GitHub pull requests.

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

## `lstn in [path]`

Inspect the verdicts for your dependencies tree.

### Flags

```
    --json                output the verdicts (if any) in JSON form
-l, --lockfiles strings   set one or more lock file paths (relative to the working dir) to lookup for (default [package-lock.json,poetry.lock])
```

### Config Flags

```
--loglevel string        set the logging level (default "info")
--npm-endpoint string    the listen.dev endpoint emitting the NPM verdicts (default "https://npm.listen.dev")
--pypi-endpoint string   the listen.dev endpoint emitting the PyPi verdicts (default "https://pypi.listen.dev")
--timeout int            set the timeout, in seconds (default 60)
```

### Debug Flags

```
--debug-options   output the options, then exit
```

### Filtering Flags

```
-q, --jq string   filter the output verdicts using a jq expression (requires --json)
```

### Registry Flags

```
--npm-registry string   set a custom NPM registry (default "https://registry.npmjs.org")
```

### Reporting Flags

```
    --gh-owner string                                               set the GitHub owner name (org|user)
    --gh-pull-id int                                                set the GitHub pull request ID
    --gh-repo string                                                set the GitHub repository name
-r, --reporter (gh-pull-check,gh-pull-comment,gh-pull-review,pro)   set one or more reporters to use (default [])
```

### Token Flags

```
--gh-token string    set the GitHub token
--jwt-token string   set the listen.dev auth token
```

For example:

```bash
lstn in
lstn in .
lstn in /we/snitch
lstn in sub/dir
lstn in --lockfiles poetry.lock,package-lock.json
lstn in /pyproj --lockfiles poetry.lock
```

## `lstn manual`

A comprehensive reference of all the lstn commands.

## `lstn reporters`

A comprehensive guide to the `lstn` reporting mechanisms.

## `lstn scan [path]`

Inspect the verdicts for your direct dependencies.

### Flags

```
--json   output the verdicts (if any) in JSON form
```

### Config Flags

```
--loglevel string        set the logging level (default "info")
--npm-endpoint string    the listen.dev endpoint emitting the NPM verdicts (default "https://npm.listen.dev")
--pypi-endpoint string   the listen.dev endpoint emitting the PyPi verdicts (default "https://pypi.listen.dev")
--timeout int            set the timeout, in seconds (default 60)
```

### Debug Flags

```
--debug-options   output the options, then exit
```

### Filtering Flags

```
    --ignore-deptypes (dep,dev,optional,peer)   the list of dependencies types to not process (default [bundle])
    --ignore-packages strings                   the list of packages to not process
-q, --jq string                                 filter the output verdicts using a jq expression (requires --json)
-s, --select string                             filter the output verdicts using a jsonpath script expression (server-side)
```

### Registry Flags

```
--npm-registry string   set a custom NPM registry (default "https://registry.npmjs.org")
```

### Reporting Flags

```
    --gh-owner string                                               set the GitHub owner name (org|user)
    --gh-pull-id int                                                set the GitHub pull request ID
    --gh-repo string                                                set the GitHub repository name
-r, --reporter (gh-pull-check,gh-pull-comment,gh-pull-review,pro)   set one or more reporters to use (default [])
```

### Token Flags

```
--gh-token string   set the GitHub token
```

For example:

```bash
lstn scan
lstn scan .
lstn scan sub/dir
lstn scan /we/snitch
lstn scan /we/snitch --ignore-deptypes peer
lstn scan /we/snitch --ignore-deptypes dev,peer
lstn scan /we/snitch --ignore-deptypes dev --ignore-deptypes peer
lstn scan /we/snitch --ignore-packages react,glob --ignore-deptypes peer
lstn scan /we/snitch --ignore-packages react --ignore-packages glob,@vue/devtools
```

## `lstn to <name> [[version] [shasum] | [version constraint]]`

Get the verdicts of a package.

### Flags

```
--json   output the verdicts (if any) in JSON form
```

### Config Flags

```
--loglevel string        set the logging level (default "info")
--npm-endpoint string    the listen.dev endpoint emitting the NPM verdicts (default "https://npm.listen.dev")
--pypi-endpoint string   the listen.dev endpoint emitting the PyPi verdicts (default "https://pypi.listen.dev")
--timeout int            set the timeout, in seconds (default 60)
```

### Debug Flags

```
--debug-options   output the options, then exit
```

### Filtering Flags

```
-q, --jq string       filter the output verdicts using a jq expression (requires --json)
-s, --select string   filter the output verdicts using a jsonpath script expression (server-side)
```

### Registry Flags

```
--npm-registry string   set a custom NPM registry (default "https://registry.npmjs.org")
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
-v, -- count      increment the verbosity level
    --changelog   output the relase notes URL
```

### Debug Flags

```
--debug-options   output the options, then exit
```

