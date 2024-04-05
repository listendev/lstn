# lstn environment variables

The environment variables override any corresponding configuration setting.

But flags override them.

`LSTN_GH_OWNER`: set the GitHub owner name (org|user)

`LSTN_GH_PULL_ID`: set the GitHub pull request ID

`LSTN_GH_REPO`: set the GitHub repository name

`LSTN_GH_TOKEN`: set the GitHub token

`LSTN_IGNORE_DEPTYPES`: the list of dependencies types to not process

`LSTN_IGNORE_PACKAGES`: the list of packages to not process

`LSTN_JWT_TOKEN`: set the listen.dev auth token

`LSTN_LOCKFILES`: set one or more lock file paths (relative to the working dir) to lookup for

`LSTN_LOGLEVEL`: set the logging level

`LSTN_NPM_ENDPOINT`: the listen.dev endpoint emitting the NPM verdicts

`LSTN_NPM_REGISTRY`: set a custom NPM registry

`LSTN_PYPI_ENDPOINT`: the listen.dev endpoint emitting the PyPi verdicts

`LSTN_REPORTER`: set one or more reporters to use

`LSTN_SELECT`: filter the output verdicts using a jsonpath script expression (server-side)

`LSTN_TIMEOUT`: set the timeout, in seconds

