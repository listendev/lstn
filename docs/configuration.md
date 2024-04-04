# lstn configuration file

The `lstn` CLI looks for a configuration file .lstn.yaml in your `$HOME` directory when it starts.

In this file you can set the values for the global `lstn` configurations.
Anyways, notice that environment variables, and flags (if any) override the values in your configuration file.

Here's an example of a configuration file (with the default values):

```yaml
endpoint: 
  npm: "https://npm.listen.dev"
  pypi: "https://pypi.listen.dev"
filtering: 
  expression: "..."
  ignore: 
    deptypes: 
      - "..."
      - "..."
    packages: 
      - "..."
      - "..."
loglevel: "info"
registry: 
  npm: "https://registry.npmjs.org"
reporting: 
  github: 
    owner: "..."
    pull: 
      id: 0
    repo: "..."
  types: 
    - "..."
    - "..."
timeout: 60
token: 
  github: "..."
  jwt: "..."
```
