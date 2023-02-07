### General information

A `lstn` release has the following parts:

- a GitHub release, which further includes
  - packages for different operating systems and architectures
  - a file with checksums for the packages
  - a changelog
  - the source code

### How to release a new version

A release is triggered by pushing a new tag which starts with `v` (for example `v1.2.3`).

Everything else is then handled by GoReleaser and GitHub actions.

To push a new tag, please use the tool [make/tag](https://github.com/listendev/lstn/blob/main/make/tag/main.go),
which also checks if the version conforms to SemVer.

For example:

```bash
go build -o make/make make/main.go
make/make tag 0.22.1
```

### Nightly builds

We do not provide nightly builds at the moment.

### Implementation

The GitHub release is created with [GoReleaser](https://github.com/goreleaser/goreleaser/). You can take a look at the GoReleaser file [here](https://github.com/listendev/lstn/blob/main/.goreleaser.yml).
