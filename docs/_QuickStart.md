# Quick Start
This Quick Start will cover how to install Kelda and deploy [a simple Node.js and
MongoDB application](https://github.com/kelda/node-todo.git) to the cloud (e.g.
AWS or Google) using Kelda.

## Installing Kelda
1. **Install Node.js**. Download and install Node.js version 6 or higher
using the installer from the [Node.js download page](https://nodejs.org/en/download/).
2. **Install Kelda**. Use Node.js's package manager, `npm`, to install Kelda:

    ```console
    $ npm install -g @kelda/install
    ```

    Or, as root:

    ```console
    $ sudo npm install -g @kelda/install --unsafe-perm
    ```
3. **Check that Kelda was installed**. Verify that the `kelda` command outputs a
blurb about how to use the command.

## Configuring a Cloud Provider
This section explains how to get credentials for your cloud provider. Kelda will
need these to launch machines on your account. Before proceeding:

1. **Create an account**. Make sure you have an account with one of Kelda's
  supported cloud providers ([Google Cloud](https://cloud.google.com/),
  [Amazon EC2](https://aws.amazon.com/ec2/), or
  [DigitalOcean](https://www.digitalocean.com/))

2. **Get credentials**. Locate the credentials for your cloud provider account.

For Amazon EC2, create an account with [Amazon Web Services](https://aws.amazon.com/ec2/),
find your "Access Keys" on the
[Security Credentials](https://console.aws.amazon.com/iam/home?#security_credential)
page in the AWS Management Console, and "Create New Access Key" (or use an
existing key, if you already have one).

Alternatively, follow the instructions for [Google Cloud](http://docs.kelda.io/#google-compute-engine)
or [DigitalOcean](http://docs.kelda.io/#digitalocean), but **come back to this
tutorial before running `kelda init`**. In the next step, you will run
`kelda init` with some specific inputs.

## Creating an Infrastructure
1. **Specify an infrastructure**. Kelda needs to know what infrastructure (e.g. which
  cloud provider and VMs) to launch the application on. The easiest way to specify
  this is by running `kelda init`:

    ```console
    $ kelda init
    ```

**Note!** For the sake of this tutorial, make sure to use the name **`default`**
when answering the first question from `kelda init`.

**Note!** When asked for credentials, use the provider credentials from the
previous section.

<aside class="notice">If you are unsure about how to answer any of the
questions, the default values are appropriate for this tutorial.
</aside>

## Getting the Blueprint

1. **Download Kelda's Node.js and MongoDB blueprint** using git:

    ```console
    $ git clone https://github.com/kelda/nodejs.git
    ```

2. **Install the JavaScript dependencies** of the blueprint using `npm`, the
Node.js package manager:

    ```console
    $ cd nodejs
    $ npm install
    ```

## Running the Application

1. **Start the daemon**. To run a blueprint, the Kelda daemon must be running.
If no daemon is running, open a new terminal window and start one:

    ```console
    $ kelda daemon
    ```

    The daemon is a long running process that periodically prints some log messages.
    Leave this running.

2. **Start the application**. In another terminal window, navigate to the `nodejs`
    directory and run the blueprint:

    ```console
    $ kelda run ./nodeExample.js
    ```

3. **Check the status**. The VMs and application containers are now booting.
Use Kelda's `show` command, to see how things are progressing.
The output looks similar to this:

    ```console
    $ kelda show
    MACHINE    ROLE      PROVIDER    REGION       SIZE        PUBLIC IP    STATUS
               Master    Amazon      us-west-2    t2.micro                 booting
               Worker    Amazon      us-west-2    t2.micro                 booting
    ```

    `kelda show` might temporarily show an error message starting with
    "unable to query connections: rpc error"; this error is benign and can occur
    while the machines are booting.

4. **Wait** a few minutes until the machines' `STATUS` are `connected`, and
both containers' `STATUS` are `running`:

    ```console
    $ kelda show
    MACHINE         ROLE      PROVIDER    REGION       SIZE        PUBLIC IP         STATUS
    i-0e8b292380    Master    Amazon      us-west-2    t2.micro    54.149.231.255    connected
    i-0fee34512f    Worker    Amazon      us-west-2    t2.micro    54.214.109.175    connected

    CONTAINER       MACHINE         COMMAND                    HOSTNAME     STATUS     CREATED               PUBLIC IP
    703ed73b87ee    i-0fee34512f    keldaio/mongo              mongo        running    About a minute ago
    fa33354be048    i-0fee34512f    node-app:node-todo.git     node-app2    running    27 seconds ago        54.214.109.175:80
    ```

    Don't worry if the `STATUS` is empty for a a few minutes -- this is
    normal while things are starting up.

## Accessing the Web App
To access the website, simply copy-paste the `PUBLIC IP` address from the
`node-app` container row (in this case `54.214.109.175`) into your browser.
You should see the todo app.

<aside class="notice"><strong>Make sure to stop your deployment!</strong> You
can skip the next few sections, but make sure to read the section on
<a href="http://docs.kelda.io/#stopping-the-application">how to stop your
application</a> to avoid getting charged for any VMs that are left running.
</aside>

## Debugging Applications with Kelda
Note that the commands in this section (and all Kelda commands that take IDs)
don't need the full ID; a unique prefix of the ID is enough.

1. **Check the logs**. To get the logs of the Node.js app, find the
container's ID in the `kelda show` output, and pass it to the `kelda logs`
command:

    ```console
    $ kelda logs fa33
    ```
    Verify that as you interact with the todo app, the logs show the
    corresponding GET, POST, and DELETE requests. This isn't thrilling information,
    but the logs will come in handy if you ever encounter any errors.

2. **SSH in to the container**. To SSH in to the Node.js container, execute
`kelda ssh` with the container ID:

    ```console
    $ kelda ssh fa33
    ```

    Try to run `ls` to see the application source files.

3. **Log out**. When you're done poking around, log out with the `exit` command.

## Stopping the Application
1. **Stop the VMs**. When you're done playing around, __make sure to stop the machines
you've started!__ Otherwise, your cloud provider might charge you for the
VMs that are still running.  To stop all VMs and containers:

    ```console
    $ kelda stop
    ```
2. **Check that the machines are stopped**. Wait a few seconds until the VMs no
longer show up in `kelda show`.
3. **Stop the daemon**. Go to the terminal window that's running `kelda daemon`
and stop the process with `Ctrl+C`, or just close the window.

## Next steps
* [**Deploy your own Node.js/MongoDB app**](https://github.com/kelda/nodejs#deploying-your-own-nodejs-and-mongodb-application).
* [**Learn how Kelda works**](#how-kelda-works).
* [**Serve the application from multiple servers**](#how-to-run-a-replicated-load-balanced-application-behind-a-single-ip-address).
* [**Host a web app with a custom domain name**](#how-to-give-your-application-a-custom-domain-name).
