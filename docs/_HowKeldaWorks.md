# How Kelda Works

This section describes what happens when you run an application using Kelda,
and explains the main components in a deployment. As an example, we will
consider the deployment in this illustration:

![Kelda Diagram](Kelda_Diagram.png)

### Blueprint
The key idea behind Kelda is a blueprint -- in this case `my_app.js`. A
blueprint describes every aspect of
running a particular application in the cloud, and is written in JavaScript.
Kelda blueprints exist for many common applications.  Using Kelda, you can run
one of those applications by executing just two commands on your laptop:
`kelda daemon` and `kelda run`.

### Daemon
The first command, `kelda daemon`, starts a long-running process -- "the
daemon". The daemon is responsible for booting machines in the cloud,
configuring the machines, managing them, and rebooting them if anything goes
wrong. This means that if a machine dies while no daemon is running for its
deployment (e.g. because you closed the laptop that's running the daemon), the
machine will not be restarted. However, as soon as the daemon comes back online,
it will reboot the missing VM. The `kelda daemon` command starts
the daemon, but doesn't launch any machines until you `kelda run` a blueprint.
Likewise, stopping the daemon (e.g. when closing your laptop) doesn't stop the
running machines -- only `kelda stop` will cause the daemon to terminate VMs.

### Run
To launch an application, call `kelda run` with a JavaScript blueprint -- in
this example `my_app.js`. The `run` command passes the parsed blueprint to the
daemon, and the daemon sets up the infrastructure described in the blueprint.

### Containers
Kelda runs applications using [Docker](http://docker.com/). As shown
in the graphic, `my_app.js` described an application with three Docker
containers. You can think of a container
as being like a process: as a coarse rule-of-thumb, anything that you'd launch
as its own process should have it's own container with Kelda.  While containers
are lightweight (like processes), they each have their own environment
(including their own filesystem and their own software installed) and are
isolated from other containers running on the same machine (unlike processes).
If you've never used containers before, it may be helpful to review the
[Docker getting started guide](https://docs.docker.com/get-started).

### Machines
In terms of machines, `myApp.js` described a cluster with one master and two
workers.

**Workers**: The worker machines host the application containers. In this case,
Kelda ran two containers on one worker machine and one container on the other.

**Masters**: The master is responsible for managing the worker machines and the
application containers running on the workers. No application containers run on
the master. In the above explanation of [the daemon](#daemon), we described a
scenario where a worker machine disappears. In that scenario, the master machine
would reboot the application containers from the failed worker on one of the
healthy worker machines. It is possible to have multiple master machines to
safeguard against a master machine failing.
