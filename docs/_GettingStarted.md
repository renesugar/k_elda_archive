# Getting Started

## Installing Quilt
Quilt relies on Node.js.  Start by downloading Node.js from the [Node.js
download page](https://nodejs.org/en/download/).  We have only tested Quilt with
Node version v7.10.0 and above.

Next, use Node.js's package manager, `npm`, to install Quilt:

```console
$ npm install -g @quilt/install
```

Note that if installing as root, the `--unsafe-perm` flag is required:

```console
$ sudo npm install -g @quilt/install --unsafe-perm
```

To check that this worked, try launching the Quilt daemon.  This is a
long-running process, so it will not return (you'll need to use a new shell
window to edit and run blueprints).

```console
$ quilt daemon
```

## Configuring a Cloud Provider
Quilt needs access to your cloud provider credentials in order to launch
machines on your account. Before continuing, make sure you have an account
with one of Quilt's supported cloud providers ([Google Cloud](https://cloud.google.com/),
[Amazon EC2](https://aws.amazon.com/ec2/), or
[DigitalOcean](https://www.digitalocean.com/)), and that you have located your
credentials for this account.

For Amazon EC2, you can create an account with [Amazon Web Services](https://aws.amazon.com/ec2/)
and then find your access credentials from the
[Security Credentials](https://console.aws.amazon.com/iam/home?#security_credential)
page in the AWS Management Console.

If you prefer to use another provider than Amazon EC2, check out
[Cloud Provider Configuration](#cloud-provider-configuration). You'll need the
credentials in the next step.

## Creating an Infrastructure
To run any applications with Quilt, you need to give Quilt access to your cloud
provider account from the last step, and specify an infrastructure that Quilt
should launch your application on. The easiest way to do this, is to run

```console
$ quilt init
```

The `quilt init` command will ask a number of questions, and then set up the
provider and infrastructure based on the given answers. For the sake of this
tutorial, make sure to use the name **`default`** for the infrastructure.
Additionally, when choosing a machine instance size, keep in mind that some
providers have a free tier for certain instance types.

For more information about `quilt init`, see [the documentation](#init).

It is also possible to use the
[Quilt blueprint API](#quilt.js-api-documentation) to specify
[`Machine`s](#Machine) directly in blueprints, but that's a topic for another
time.

## Getting the Blueprint
This section will walk you through using Quilt to run Nginx, which is an
open-source HTTP server.  In the example, we'll use Nginx to serve a
simple webpage. Start by downloading the blueprint using git:

```console
$ git clone https://github.com/quilt/nginx.git
```

The blueprint in `main.js` imports the `app.js` Nginx blueprint, and then
deploys the Nginx app to the base infrastructure you created with `quilt init`.

Before running anything, you'll need to download the JavaScript dependencies of
the blueprint.  The Nginx blueprint depends on the `@quilt/quilt` module; more
complicated blueprints may have other dependencies as well. Use `npm`, the
Node.js package manager, to install all dependencies in the `nginx` folder:

```console
$ npm install .
```

## Running the Application
To run a blueprint, you first need to have a Quilt daemon running.  If you
haven't already started one, open a new terminal window and start it:

```console
$ quilt daemon
```

The daemon is a long running process that periodically prints some log messages.
Leave this running, and use a new terminal window to run the blueprint:

```console
$ quilt run ./main.js
```

This command tells the daemon to launch the machines and containers described in
`main.js`.  It will return immediately, because the `daemon` process does the
heavy lifting.

It takes a few minutes from you `quilt run` until the VMs and application are
fully up and running. To see how things are progressing, use Quilt's `show`
command:

```console
$ quilt show
MACHINE         ROLE      PROVIDER    REGION       SIZE     PUBLIC IP    STATUS
e5b1839d2bea    Master    Amazon      us-west-1    t2.micro              disconnected
e2401c348c78    Worker    Amazon      us-west-1    t2.micro              disconnected
```

Your output will look similar to the output above (note that you may get an
error that begins with "unable to query connections: rpc error" when you first
run `quilt show`; this error is benign and can occur while the machines are
booting).

The output above means that Quilt has launched two machines, one as a master and
one as a worker, in Amazon.  Both machines are `disconnected`, because they're
still being initialized. When a machine is fully booted and configured, it will
be marked as `connected` in the `STATUS` column.

## Accessing the Web App
After a few minutes, when the VMs and containers are up and running, the
output of `show` will look something like this:

```console
$ quilt show
MACHINE         ROLE      PROVIDER    REGION       SIZE        PUBLIC IP         STATUS
e5b1839d2bea    Master    Amazon      us-west-1    t2.micro    54.183.98.15      connected
e2401c348c78    Worker    Amazon      us-west-1    t2.micro    54.241.251.192    connected

CONTAINER       MACHINE         COMMAND        HOSTNAME           STATUS     CREATED               PUBLIC IP
bd681b3d3af7    e2401c348c78    nginx:1.13     web_tier           running    About a minute ago    54.241.251.192:80
```

The bottom row lists the container that's running `nginx`.  The `nginx`
deployment is relatively simple and has just one container, but a typical
application running in Quilt will have many containers running (one for each
part of the application; for example, your website application might require a
second container that runs a database).

The `PUBLIC IP` column shows the address you can use to access the website.
Simply copy-paste this IP address into your browser. A site with "Hello, world!"
text should appear.

This is all it takes to run an application on Quilt. The remainder of this
tutorial will cover some of the things you might want to do after your
application is up and running -- e.g. debugging or changing the website content,
and importantly, how to shut down the deployment.

<aside class="notice">You can skip the next few sections, but make sure to read
the section on <a href="http://docs.quilt.io/#stopping-the-application">
how to stop your application</a> to avoid getting charged for any VMs that are
left running.</aside>

## Making the Application Publicly Accessible
By default, Quilt-managed containers are disconnected from the public internet
and isolated from one another. This helps to keep your application secure by
preventing all access except for what you explicitly specify.
In order to make the Nginx container accessible
from the public internet,
[`nginx/app.js`](https://github.com/quilt/nginx/blob/master/app.js) explicitly
opens port 80 on the Nginx container to the outside world:

```javascript
webTier.allowFrom(publicInternet, 80);
```

Without this line, the website wouldn't be accessible from the browser.

## Debugging Applications with Quilt
Once the containers are running, you might need to log in to change something
or debug an issue.  The `quilt ssh` command makes this easy.  Use the container
ID from the `quilt show` output as the argument to `quilt ssh` to log in to that
container. For instance, to ssh into a container or VM whose ID starts with
bd68:

```console
$ quilt ssh bd68
```

To check the logs of the same container or VM, use `quilt logs`:

```console
$ quilt logs bd68
```

If you run `logs` on the nginx container for instance, you'll see that nginx
logs a GET request for each time you access the website. This is not thrilling
information, but the logs will come in handy if you ever encounter any errors.

Note that you don't need to type the whole ID; as long as Quilt gets a unique
prefix of the ID, it will log in to the correct machine.

## Changing the Website Content
You may later decide that you'd like to change the contents of the simple
website.  You could do this by logging into the container, but for the sake of
example, let's do it using Quilt.  On your laptop, open the `nginx/index.html`
that was downloaded from github and change the contents of the page (e.g., you
can change the "Hello, World!" message).  Now, with the daemon still running,
re-deploy the webpage with Quilt:

```console
$ quilt run ./main.js
```

Quilt automatically detects the changes to the deployment, and will update it
to implement your changes.  Note that we didn't need to tell Quilt to
stop the nginx container and start a new one; we just updated the view of what
the deployment should look like (in this case, by changing `index.html`), and
Quilt automatically detects this and updates the cluster accordingly.  Quilt
will prompt you to accept the changes that you're making to your deployment;
type `y`.

Run `quilt show` again and notice that Quilt has stopped the old
container and is starting a new one.  When the new container is `running`,
navigate to the new IP address and check that the modified page is up.

## Stopping the Application
When you're done experimenting with Quilt, __make sure to stop the machines
you've started!__  Otherwise, your cloud provider might charge you for the
VMs that are still running.  Quilt's `stop` command will stop all VMs and
containers:

```console
$ quilt stop
```

At this point, you can safely kill the Quilt daemon.
