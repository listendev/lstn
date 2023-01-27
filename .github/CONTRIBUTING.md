## Contributing

Hello!

Thank you for your interest in contributing to the listen.dev CLI!

We accept pull requests for bug fixes and features where we've discussed the approach in an issue and given the go-ahead for a community member to work on it. We'd also love to hear about ideas for new features as issues.

Please **do**:

- Check existing issues to verify that the [`bug`][bug issues] or [`feature request`][feature request issues] has not already been submitted.
- Open an issue if things aren't working as expected.
- Open an issue to propose a significant change.
- Open a pull request to fix a bug.
- Open a pull request to fix documentation about a command.
- Open a pull request for any issue labelled [`help wanted`][hw] or [`good first issue`][gfi].

Please **avoid**:

- Opening pull requests for issues marked [`needs-design`][needs design], [`needs-investigation`][needs investigation], or [`blocked`][blocked].
- Opening pull requests for any issue marked [`core`][core].
  - These issues require additional context from the core CLI team and any external pull requests will not be accepted.

## Building the CLI

Prerequisites:

- Go 1.19+

Build with:

_TODO_

Run the CLI as:

_TODO_

## Testing the CLI

Run tests with: `go test ./...`

## Documenting the CLI

See [project layout documentation](../docs/project-layout.md) for information on where to find specific source files.

We generate manual pages from source on every release. You do not need to submit pull requests for documentation specifically; manual pages for commands will automatically get updated after your pull requests gets accepted.

## Submitting a pull request

1. Create a new branch: `git checkout -b my-branch-name`
1. Make your change, add tests, and ensure tests pass
1. Submit a pull request: `gh pr create --web`

Contributions to this project are [released][legal] to the public under the [project's open source license][license].

Please note that this project adheres to a [Contributor Code of Conduct][code-of-conduct]. By participating in this project you agree to abide by its terms.

## Resources

- [How to Contribute to Open Source][]
- [Using Pull Requests][]
- [GitHub Help][]

[bug issues]: https://github.com/listendev/lstn/issues?q=is%3Aopen+is%3Aissue+label%3Abug
[feature request issues]: https://github.com/listendev/lstn/issues?q=is%3Aopen+is%3Aissue+label%3Aenhancement
[hw]: https://github.com/listendev/lstn/issues?q=is%3Aopen+is%3Aissue+label%3A"help+wanted"
[blocked]: https://github.com/listendev/lstn/issues?q=is%3Aopen+is%3Aissue+label%3Ablocked
[needs design]: https://github.com/listendev/lstn/issues?q=is%3Aopen+is%3Aissue+label%3A"needs+design"
[needs investigation]: https://github.com/listendev/lstn/issues?q=is%3Aopen+is%3Aissue+label%3A"needs+investigation"
[gfi]: https://github.com/listendev/lstn/issues?q=is%3Aopen+is%3Aissue+label%3A"good+first+issue"
[core]: https://github.com/listendev/lstn/issues?q=is%3Aopen+is%3Aissue+label%3Acore
[legal]: https://docs.github.com/en/free-pro-team@latest/github/site-policy/github-terms-of-service#6-contributions-under-repository-license
[license]: ../LICENSE
[code-of-conduct]: ./CODE_OF_CONDUCT.md
[how to contribute to open source]: https://opensource.guide/how-to-contribute/
[using pull requests]: https://docs.github.com/en/free-pro-team@latest/github/collaborating-with-issues-and-pull-requests/about-pull-requests
[github help]: https://docs.github.com/
