# Quilt CLI
Quilt's CLI, `quilt`, is a handy command line tool for starting, stopping, and
managing deployments. Quilt CLI commands have the following format:

```console
$ quilt [OPTIONS] COMMAND
```

To see the help text for a specific command, run:

```console
$ quilt COMMAND --help
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
| `counters`   | Display internal counters tracked for debugging purposes. Most users will not need this command. |
| `daemon`     | Start the quilt daemon, which listens for quilt API requests.                                    |
| `debug-logs` | Fetch logs for a set of machines or containers.                                                  |
| `init`       | Create an infrastructure that can be accessed in blueprints using baseInfrastructure().          |
| `inspect`    | Visualize a blueprint.                                                                           |
| `logs`       | Fetch the logs of a container or machine minion.                                                 |
| `minion`     | Run the quilt minion.                                                                            |
| `show`       | Display the status of quilt-managed machines and containers.                                     |
| `run`        | Compile a blueprint, and deploy the system it describes.                                         |
| `ssh`        | SSH into or execute a command in a machine or container.                                         |
| `stop`       | Stop a deployment.                                                                               |
| `version`    | Show the Quilt version information.                                                              |

## Init
The `quilt init` command is a simple way to create reusable infrastructure. The
command prompts the user for information about their desired infrastructure
and then creates an infrastructure based on the answers.
The infrastructure can be used in blueprints by calling
`baseInfrastructure(NAME)`, where `NAME` is the infrastructure name given to
`quilt init`.

It is possible to create multiple infrastructures with `quilt init`, but we
recommend at least having a small infrastructure called `default` with your
standard configuration. Some example blueprints will assume such a `default`
infrastructure exists.

To edit the infrastructure after creation, either rerun `quilt init`
using the same name, or directly edit the infrastructure blueprint stored in
`~/.quilt/infra/<NAME>.js`.

Most of the `quilt init` questions are self-explanatory, but the following might
warrant a little explanation:

* **Infrastructure Name**: As explained above, the infrastructure name is used
when retrieving the infrastructure with
[`baseInfrastructure()`](#quilt-js-api-documentation).
* **Provider Keys**: In order to launch virtual machines from your account, Quilt needs access to
your provider credentials. The credentials are used when Quilt makes API calls
to the provider. Quilt will not store your credentials, but simply needs
access to a credentials file on your machine. If there is no existing
credentials file, `quilt init` helps create one with the correct format. See
[Cloud Provider Configuration](#cloud-provider-configuration)
for instructions on how to get your cloud provider credentials.
* **SSH Keys**: An SSH key is required for SSHing into VMs and containers, and
for executing a number of helpful Quilt CLI commands, such as `quilt logs`. It
is recommended to add an SSH key to all `Machine`s.
