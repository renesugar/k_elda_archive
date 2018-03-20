# Kelda CLI
Kelda's CLI, `kelda`, is a handy command line tool for starting, stopping, and
managing deployments. Kelda CLI commands have the following format:

```console
$ kelda [OPTIONS] COMMAND
```

To see the help text for a specific command, run:

```console
$ kelda COMMAND --help
```

## Options
| Name, shorthand     | Default | Description                                               |
|---------------------|---------|-----------------------------------------------------------|
| `--log-level`, `-l` | `info`  | Logging level (debug, info, warn, error, fatal, or panic) |
| `--verbose`, `-v`   | `false` | Turn on debug logging                                     |
| `-log-file`         |         | Log output file (will be overwritten)                     |

## Commands
| Name         | Description                                                                                      |
|--------------|--------------------------------------------------------------------------------------------------|
| `base-infrastructure` | Create a new base infrastructure. The infrastructure can be used in blueprints by calling [`baseInfrastructure()`](#kelda-js-api-documentation). |
| `configure-provider` | Set up cloud provider credentials. This command helps ensure that the file format and location are as Kelda expects. |
| `counters`   | Display internal counters tracked for debugging purposes. Most users will not need this command. |
| `daemon`     | Start the kelda daemon, which listens for kelda API requests.                                    |
| `debug-logs` | Fetch logs for a set of machines or containers.                                                  |
| `init`       | Create an infrastructure that can be accessed in blueprints using baseInfrastructure().          |
| `inspect`    | Visualize a blueprint.                                                                           |
| `logs`       | Fetch the logs of a container or machine minion.                                                 |
| `minion`     | Run the kelda minion.                                                                            |
| `show`       | Display the status of kelda-managed machines and containers.                                     |
| `run`        | Compile a blueprint, and deploy the system it describes.                                         |
| `secret`     | Securely add a named secret to the cluster.                                                      |
| `ssh`        | SSH into or execute a command in a machine or container.                                         |
| `stop`       | Stop a deployment.                                                                               |
| `version`    | Show the Kelda version information.                                                              |
