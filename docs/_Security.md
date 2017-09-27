# Security

## TLS
Quilt uses [grpc](http://www.grpc.io/) for communication with the daemon and
deployed clusters. Functionality exposed through `grpc` includes deploying new
blueprints and querying deployment information. All communication is
automatically encrypted and verified using TLS.

### Quickstart
Start the daemon. If credentials don't already exist, they will be
automatically generated.

```console
$ quilt daemon
```

Use the other Quilt commands as normal.

```console
$ quilt run ./example.js
$ quilt show
MACHINE         ROLE      PROVIDER    REGION       SIZE         PUBLIC IP         STATUS
8a0d2198229c    Master    Amazon      us-west-1    m3.medium    54.153.11.92      connected
b92d625c6847    Worker    Amazon      us-west-1    m3.medium    52.53.170.129     connected

CONTAINER       MACHINE         COMMAND                     HOSTNAME  STATUS     CREATED           PUBLIC IP
1daa461f0805    b92d625c6847    alpine tail -f /dev/null    alpine    running    24 seconds ago    52.53.170.129:8000
```

Only a user with access to the correct credentials can connect to the cluster.
As an example of this, if you delete your credentials, restart the daemon, and
run the same blueprint, you won't be able to connect to the machines:

```console
$ rm -rf ~/.quilt/tls
$ quilt daemon
$ quilt run ./example.js
$ quilt show
MACHINE         ROLE      PROVIDER    REGION       SIZE         PUBLIC IP        STATUS
8a0d2198229c    Master    Amazon      us-west-1    m3.medium    54.153.11.92     connecting
b92d625c6847    Worker    Amazon      us-west-1    m3.medium    52.53.170.129    connecting
```

### TLS credentials
`quilt daemon` autogenerates TLS credentials if necessary. They are stored in
`~/.quilt/tls`. The directory structure is as follows:

```console
$ tree ~/.quilt/tls
├── certificate_authority.crt
├── certificate_authority.key
├── quilt.crt
├── quilt.key
```

- certificate_authority.crt: The certificate authority certificate.
- certificate_authority.crt: The private key of the certificate authority.
Used by the daemon to generate minion certificates.
- quilt.crt: A certificate signed by the certificate authority.
Used for connecting to the cluster.
- quilt.key: The private key associated with the signed certificate.
Used for connecting to the cluster.

Other files in the directory are ignored by Quilt.

### admin-ssh-private-key
The daemon installs keys using SFTP, so the daemon requires SSH access to the
machines. By default, the daemon generates an in-memory key to use for distributing
keys. A key can be specified from the filesystem using the
`admin-ssh-private-key` flag.

For example,

```console
$ quilt daemon -admin-ssh-private-key ~/.quilt/id_rsa
```
