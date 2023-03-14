The lstn CLI follows the usual conventions regarding exit codes.

Meaning:

* when a command completes successfully, the exit code will be 0

* when a command fails for any reason, the exit code will be 1

* when a command is running but gets cancelled, the exit code will be 2

* when a command meets an authentication issue, the exit code will be 4

Notice that it's possible that a particular command may have more exit codes,
so it's a good practice to check the docs for the specific command
in case you're relying on the exit codes to control some behaviour.
