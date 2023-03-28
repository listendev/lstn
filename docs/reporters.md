# lstn reporters

## gh-pull-comment

It reports results as a sticky comment on the target GitHub pull request.

The target GitHub pull request comes from the values of the GitHub reporter flags (ie., `--gh-repo`, `--gh-owner`, `--gh-pull-id`).
Notice those values are automatically set when `lstn` detects it is running in a GitHub Action.

### Status

Working.

## gh-pull-review

It reports results to GitHub review & suggestion comments on the target GitHub pull request.

### Status

TBD.

## gh-pull-check

It reports results to the GitHub pull requests check tab.

### Limitations

When `lstn` detect it is running from a fork repository, due to [GitHub Actions restrictions](https://docs.github.com/en/actions/security-guides/automatic-token-authentication#permissions-for-the-github_token), this reporter will reports the verdicts to the GitHub Actions **log console**.

### Status

TBD.

