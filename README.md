# lstn

`lstn` is the [listen.dev](https://listen.dev) command line.

It lets you analyze the behavior of your dependencies.

TODO: image or asciinema

## Documentation

For [installation options see below](#installation), for usage instructions [see the manual](docs/cheatsheet.md).

## Installation

### CI

#### GitHub Actions

We recommend using TODO: our GitHub Action for running `lstn` in CI for GitHub projects.

#### Other CI

It's highly recommended installing a specific version of `lstn` available on the [releases page](https://github.com/listendev/lstn/releases/latest).

Here are a few ways to install it:

```bash
# The binary will be /usr/local/bin/lstn
curl -sSfL https://lstn.dev/get | sh -s -- -b /usr/local/bin

# Or install it into $PWD/bin/
curl -sSfL https://lstn.dev/get | sh -s

# In Alpine Linux (as it does not come with curl by default)
wget -O- -nv https://lstn.dev/get | sh -s
```

You can test the installation by running:

```bash
lstn version
```

### Locally

Wanna install `lstn` on your machine?

See options below.

#### Binaries

```bash
curl -sSfL https://lstn.dev/get | sh -s -- -b /usr/local/bin
lstn version
```

#### macOS

`lstn` is available via TODO: Homebrew, ..., and as a downloadable binary from the [releases page](https://github.com/listendev/lstn/releases/latest).

#### Linux & BSD

`lstn` is available via:

- TODO: our Debian and RPM repositories
- OS-agnostic package managers such as TODO: Homebrew, ...
- our [releases pages](https://github.com/listendev/lstn/releases/latest) as precompiled binaries.

#### From source

We recommend using binary installation. Using `go install` or `go get` might work but those aren't guaranteed to.

<details>
<summary>Why?</summary>
<ol>
<li>Some users use the <code>-u</code> flag for <code>go get</code> which upgrades our dependencies: we can not guarantee they work!</li>
<li>The <code>go.mod</code> replacement directive doesn't apply.</li>
<li>The <code>lstn</code> stability may depend on a user's Go version.</li>
<li>It allows installation from the main branch which can't be considered stable.</li>
<li>It is way slower than binary installation.</li>
</ol>
</details>

## Contributing

If anything feels off, or if you feel that some functionality is missing, please check out the [contributing page](.github/CONTRIBUTING.md).

There you will find instructions for sharing your feedback, building the tool locally, and submitting pull requests to the project.
