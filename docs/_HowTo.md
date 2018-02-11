# How To

## How to Give Your Application a Custom Domain Name

1. **Buy your domain name**  from a registrar like
  [Namecheap](https://www.namecheap.com/) or [GoDaddy](https://www.godaddy.com).
2. **Get a floating IP** (also called Elastic IP or Static External IP) through
  your cloud provider's management console or command line tool. When you
  reserve a floating IP, it is guaranteed to be yours until you explicitly
  release it.
3. **Point the domain to your IP** by modifying the A record on your registrar's
  website, so the domain points at the floating IP from last step.
4. **Run a blueprint** that hosts the website on that floating IP. The next
section describes how to write the blueprint.

### Blueprint: Hosting a Website on a Floating IP
Assigning a floating IP address to an application just involves two steps:

1. Deploy a worker machine with the floating IP:

    ```javascript
    // The floating IP you registered with the cloud provider -- say, Amazon.
    const floatingIP = '1.2.3.4';

    const baseMachine = new Machine({ provider: 'Amazon' });

    // Set the IP on the worker machine.
    const worker = baseMachine.clone();
    worker.floatingIp = floatingIp;

    // Create the infrastructure.
    const inf = new Infrastructure({
      masters: baseMachine,
      workers: worker,
    });
    ```
2. Tell Kelda to place the application on the machine with your floating IP:

    ```javascript
    const app = new Container({ name: 'myApp', image: 'myImage' });
    app.placeOn({ floatingIP });

    // Deploy the application.
    app.deploy(inf);
    ```
If your website is hosted on multiple servers, follow the guide for running a
[replicated, load balanced application](#how-to-run-a-replicated-load-balanced-application-behind-a-single-ip-address),
and simply place the `loadBalancer` on the floating IP.

## How to Update Your Application on Kelda
The most robust way to handle updates to your application is to [build and
push](https://docs.docker.com/get-started/part2/) your own tagged Docker images.
We recommend always [tagging images](https://docs.docker.com/engine/reference/commandline/build/#tag-an-image--t),
and not using the `:latest` tag.

Say that we want to update an application that uses the `me/myWebsite:0.1`
Docker image to use `me/myWebsite:0.2` instead. We can do that in two simple
steps:

1. In the blueprint, update all references to `me/myWebsite:0.1` to use the tag
`:0.2`.
2. Run `kelda run` with the updated blueprint to update the containers with the
new image.

Kelda will now restart all the relevant containers to use the new tag.

### Untagged Images and Images Specified in Blueprints
For users who do not use tagged Docker images, there are currently two ways of
updating an application. This section explains the two methods and when to use
each.

#### After Changing a Blueprint
If changes are made directly to the blueprint or a file or module that the
blueprint depends on, simply `kelda run` the blueprint again. Kelda will detect
the changes and reboot the affected containers.

Examples of when `kelda run` will update the application:

* You updated the contents of a file that is transferred to the application
container using the Container's `filepathToContent` attribute.
* You changed the Dockerfile string passed to the `Image` constructor in the
blueprint.


#### Updating an Image
Though we recommend using _tagged_ Docker images, some applications might use
untagged images either hosted in a registry like Docker Hub or created with
Kelda's `Image` constructor in a blueprint. To pull a newer version of a hosted
image or rebuild an `Image` object with Kelda, you need to restart the relevant
container:

1. In the blueprint, remove the code that `.deploy()`s the container that
should be updated.
2. `kelda run` this modified blueprint in order to stop the container.
3. Back in the blueprint, add the `.deploy()` call back in to the blueprint.
4. `kelda run` the blueprint from step 3 to reboot the container with the
updated image.

Examples of when you need to stop and start the container in order to update the
application:

* You use a hosted image (e.g. on Docker Hub), and pushed a new image with the
same `name:tag` as the currently running one.
* You want to rebuild an `Image` even though the Dockerfile in the blueprint
didn't change. For instance, this could be because the Dockerfile clones a
GitHub repository, so rebuilding the image would recreate the container with
the latest code from GitHub.

## How to Run a Replicated, Load Balanced Application Behind a Single IP Address
This guide describes how to write a blueprint that will run a replicated, load
balanced application behind a single IP address. We will use [HAProxy](https://www.haproxy.com/)
(High Availability Proxy), a popular open source load balancer, to evenly
distribute traffic between your application containers.

Before we start writing the blueprint, make sure the application you want to
replicate is listening on port 80. E.g., for an Node.js Express application
called `app`, call `app.listen(80)` in your application code.

### A Single Replicated Application
**Import [Kelda's HAProxy blueprint](https://github.com/kelda/haproxy)** in your
application blueprint:

```javascript
const haproxy = require('@kelda/haproxy');
```

**Replicate the application** container. E.g., to create 3 containers with the
`myWebsite` image:

```javascript
const appContainers = [];
for (let i = 0; i < 3; i += 1) {
  appContainers.push(new Container({ name: 'web', image: 'myWebsite' }));
}
```

**Create a load balancer** to sit in front of `appContainers`:

```javascript
const loadBalancer = haproxy.simpleLoadBalancer(appContainers);
```

**Allow requests** from the public internet to the load balancer on port 80 (the
default `exposedPort`).

```javascript
allowTraffic(publicInternet, loadBalancer, haproxy.exposedPort);
```

**Deploy** the application containers and load balancer to your infrastructure:

```javascript
const inf = baseInfrastructure();
appContainers.forEach(container => container.deploy(inf));
loadBalancer.deploy(inf);
```

*You can find a full example blueprint [here](https://github.com/kelda/haproxy/blob/master/examples/haproxyExampleSingleApp.js).*


#### Accessing the application
The application will be accessible on the `PUBLIC IP` of the `haproxy`
container.

### Multiple Replicated Applications
Say you want to run two different replicated websites with different domain
names. You could call `simpleLoadBalancer` for each of them, but that would
create two separate load balancers. This section explains how to put multiple
applications behind a single load balancer -- that is, behind a single
IP address.

The steps are basically identical to those for [running a single replicated application](#a-single-replicated-application).
There are just two important differences:

1. **Register a domain name** for each replicated application (e.g. `apples.com`
  and `oranges.com`) before deploying them. The load balancer will need the
  domain names to forward incoming requests to the right application. For more
  details, see the guide on [custom domain names](#how-to-give-your-application-a-custom-domain-name).
2. **Create the load balancer** using the `withURLrouting()` function _instead
  of_  `simpleLoadBalancer()`. As an example, the load balancer below will
  forward requests for `apples.com` to one of the containers in
  `appleContainers`, and requests for `oranges.com` will go to one of the
  containers in `orangeContainers`.

```javascript
const loadBalancer = haproxy.withURLrouting({
  'apples.com': appleContainers,
  'oranges.com': orangeContainers,
});
```

*You can find a full example blueprint [here](https://github.com/kelda/haproxy/blob/master/examples/haproxyExampleMultipleApps.js).*

#### Accessing the applications
The applications are now only available via their domain names. When the domains
are registered, the applications can be accessed with the domain name
(e.g. `apples.com`) in the browser. If a domain name isn't yet
registered, you can test that the redirects work by `cURL`ing the load balancer
and check you get the right response:

```console
$ curl -H "Host: apples.com" HAPROXY_PUBLIC_IP
```

## How to Run the Daemon
_We recommend reading about [the daemon](#daemon) before reading this section._

The default way to run the daemon is to run it on a local machine like your
laptop. However, when running a long term deployment or if multiple
people need access to the deployment, we recommend running the daemon in the
cloud.

### A Shared Daemon on a Separate VM
We recommend running a single, shared daemon on a small VM in the cloud, and
executing Kelda commands from there.

#### Setting up the Remote Daemon
1. **Create a VM** (Ubuntu or Debian) on your preferred cloud provider. You can
  choose a small instance type to keep costs low.
  - If the provider blocks ports by default, allow ingress TCP traffic on port 22.
2. **SSH** in to the VM and do the following from the VM:
  - [Install Node.js](https://nodejs.org/en/download/package-manager/#debian-and-ubuntu-based-linux-distributions).
  - [Install Kelda](#installing-kelda) with `npm`.
  - **Provider Credentials**. Set up [provider credentials](#cloud-provider-configuration).
  - **Start the Daemon**. The following command starts the daemon in the
  background, and redirects its logs to the `daemon.log` file. The `nohup`
  command ensures that the daemon keeps running even when you log out of the VM.

    ```console
    $ nohup kelda daemon > daemon.log 2>&1 &
    ```

3. **Run and Manage Applications**. All `kelda` CLI commands (e.g. `run`, `show`
  and `stop`) can now be run from this machine.

## How to Run Applications that Rely on Configuration Secrets
This section walks through an example of running an application that has
sensitive information in its configuration. Note that Kelda secrets are
currently only useful for _configuration_. Secrets generated at runtime, such
as customer information that needs to be stored in a secure database, are not
yet handled.

This section walks through deploying a GitHub bot that requires a GitHub OAuth
token in order to push to a private GitHub repository. Specifically, it
deploys the `keldaio/bot` Docker image, and configures its `GITHUB_OAUTH_TOKEN`
environment variable with a Kelda secret. Although this example uses an
environment variable, the workflow is exactly the same when installing a secret
onto the filesystem.

1. Create the Container in the blueprint. Note the secret name "githubToken".
    The name is arbitrary, but will be used in the next steps to interact with
    the secret.

    ```javascript
    const container = new kelda.Container({
      name: 'bot',
      image: 'keldaio/bot',
      env: { GITHUB_OAUTH_TOKEN: new kelda.Secret('githubToken') },
    });
    ```

2. Deploy the blueprint.

    ```console
    $ kelda run <blueprintName.js>
    ```

3. Kelda will not launch a container until all secrets needed by the container
    have been added to Kelda. Running `kelda show` after deploying the
    blueprint should result in the following:

    ```console
    CONTAINER       MACHINE         COMMAND           HOSTNAME   STATUS                                CREATED    PUBLIC IP
    d044f3880fdc    sir-m5erezkj    keldaio/bot       bot        Waiting for secrets: [githubToken]
    ```

    This means that Kelda is waiting to launch the container until the secret
    called `githubToken` is set. To set the secret value, use `kelda secret`:

    ```console
    $ kelda secret githubToken <tokenValue>
    ```

    If the command succeeds, there will be no output, and the exit code will be
    zero.

    Note that Kelda does not handle the lifecycle of the secret before `kelda
    secret` is run. For the GitHub token example, the GitHub token can be
    copied directly from the GitHub web UI to the `kelda secret` command.
    Another approach could be to store the token in a password manager, and
    paste it into `kelda secret` when needed.

4. `kelda show` should show that the bot container has started. It may take up to
    a minute for the container to start.

5. To change the secret value, run `kelda secret githubToken <newValue>`
   again, and the container will restart with the new value within a minute.

## How to Debug Network Connectivity Problems

One common problem when writing a Kelda blueprint is that the blueprint doesn't
open all of the ports necessary for the application. This typically manifests as
an application not starting properly or as the application logging messages
about being unable to connect to something. This can be difficult to debug
without a deep familiarity of which ports a particular application needs open.
One helpful way to debug this is to use the `lsof` command line tool, which can
be used to show which ports applications are trying to communicate on. The
instructions below describe how to install `lsof` and use the output to solve
a few different connectivity problems.

Suppose a blueprint includes a container called `buggyContainer` that is not
running properly because the network is not correctly setup:

```js
const buggyContainer = new kelda.Container({ name: 'buggyContainer', ... });
```

Start by enabling that container to access port 80 on the public internet, so
that the container can download the `lsof` tool, by adding the following code to
the blueprint:

```js
kelda.allowTraffic(buggyContainer, kelda.publicInternet, 80);
```

Re-run the blueprint so that Kelda will update the container's network access:

```console
$ kelda run <path to blueprint>
```

Next, login to the container and install the `lsof` tool.

```console
$ kelda ssh buggyContainer
# apt-get update
# apt-get install lsof
```

Use `lsof` to determine which ports the application in
`buggyContainer` is trying to access:

```console
# lsof -i -P -n
COMMAND PID USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
java      8 root   34u  IPv4  88572      0t0  TCP 10.14.9.118:56902->151.101.41.128:443 (SYN_SENT)
```

The useful output is the `NAME` column, which shows the source network address
(in this case, port 56902 on address 10.14.9.118) and the destination address
(in this case, port 443 on address 151.101.41.128). The `COMMAND` column may
also be useful; in this case, it shows that the network connection was initiated
by a Java process.  `SYN_SENT` means that the container tried to initiate a
connection, but the machine at 151.101.41.128 never replied (in this case
because the Kelda firewall blocked the connection). For this output, the way to
fix the problem is to enable the container to access the public internet at
port 443 by adding the following line to the blueprint:

```js
kelda.allowTraffic(buggyContainer, kelda.publicInternet, 443);
```

After re-running the blueprint, the container will be able to access port
443, which fixes this connectivity problem.

The output from `lsof` may instead show that the application
is trying to listen for connections on a particular port:

```console
# lsof -i -P -n
COMMAND PID USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
java     47 root  101u  IPv4 116320      0t0  TCP *:3000 (LISTEN)
```

In this case, there is a Java process that is running a web application on
port 3000, and the container will need to enable public access on port
3000 in order for users to access the application:

```js
kelda.allowTraffic(kelda.publicInternet, buggyContainer, 3000);
```

`lsof` may also show that internal containers are trying to communicate:

```console
# lsof -i -P -n
COMMAND PID USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
java     47 root  115u  IPv4 117041      0t0  TCP 10.181.11.2:60088->10.32.155.142:5432 (SYN_SENT)
```

In this case, the fact that the destination IP address begins with 10 signifies
that the destination is another container in the deployment, because IP
addresses beginning with 10 are private IP addresses (in the context of Kelda,
these are typically container IP addresses). This problem is somewhat
harder to debug because `kelda show` doesn't show each container's private IP
address, but it's often possible to determine which other container
`buggyContainer` is trying to connect to based on the application logs or
the port that `buggyContainer` is trying to connect to.  To see the application
logs:

```console
$ kelda logs buggyContainer
```

In this case, `buggyContainer` needs access to port 5432 on a container that
runs a postgres database (postgres runs on port 5432 by default), which can be
fixed by enabling access between those two containers:

```js
allowTraffic(buggyContainer, postgresContainer, 5432);
```

If you're curious, `lsof` lists all open files, which includes network
connections because Linux treats network connections as file handles.  The `-i`
argument tells `lsof` to only show IP files -- i.e., to only show network
connections.  The `-P` argument specifies to show port numbers rather than port
names, which is more useful here because Kelda relies on port numbers, not
names, to enable network connections.  Finally, the `-n` argument specifies to
show IP addresses rather than hostnames (the hostnames output by `lsof` are
unfortunately not the same hostnames assigned by Kelda, and are not helpful
as a result).
