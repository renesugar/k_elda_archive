# How To

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
  appContainers.push(new Container('web', 'myWebsite'));
}
```

**Create a load balancer** to sit in front of `appContainers`:

```javascript
const loadBalancer = haproxy.simpleLoadBalancer(appContainers);
```

**Allow requests** from the public internet to the load balancer on port 80 (the
default `exposedPort`).

```javascript
loadBalancer.allowFrom(publicInternet, haproxy.exposedPort);
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
  domain names to forward incoming requests to the right application.
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
