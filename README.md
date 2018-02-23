[![Build Status](https://travis-ci.org/kelda/kelda.svg?branch=master)](https://travis-ci.org/kelda/kelda)
[![Go Report Card](https://goreportcard.com/badge/github.com/kelda/kelda)](https://goreportcard.com/report/github.com/kelda/kelda)
[![Code Coverage](https://codecov.io/gh/kelda/kelda/branch/master/graph/badge.svg)](https://codecov.io/gh/kelda/kelda)
# Kelda

A straightforward way to configure containerized applications.

# What is Kelda?

Kelda is a simple way to configure, run, and manage containerized applications in
the cloud.

No more YAML, templating, complicated CLIs, and new domain specific languages.
Kelda’s intuitive JavaScript API makes it easy to describe what you want
to run in the cloud — that is, the virtual machines, containers, network policies,
and more.

Using Kelda's CLI, simply pass your high-level blueprint to Kelda’s
Kubernetes-based deployment platform. We'll take care of the nitty gritty details
needed to get your system up and running in your cloud of choice.

# Install

Install Kelda with npm:

```console
$ npm install -g @kelda/install
```

Check out more in our [Quick Start](http://docs.kelda.io/#quick-start) tutorial.

# Deploy Quickly On...

![providers](./docs/source/images/providers.png)

# How Do I Use It?
1. **Describe your application using Kelda's JavaScript API** (or use
  a [pre-made blueprint](#reference-blueprints)).

    [//]: # (b1)
    ```javascript
    const kelda = require('kelda');

    // A Docker container running your application.
    const app = new kelda.Container({
      name: 'my-app',
      image: 'myDockerImage',
    });

    // Expose the application to the public internet on port 80.
    kelda.allowTraffic(kelda.publicInternet, app, 80);
    ```

2. **Describe where to run the application**.

    [//]: # (b1)
    ```javascript
    // A virtual machine with 8 GiB RAM and 2 CPUs on Amazon EC2.
    const virtualMachine = new kelda.Machine({
      provider: 'Amazon', // 'Google', 'DigitalOcean'.
      ram: 8,
      cpu: 2,
    });

    // An infrastructure with one Kelda Master machine and 6 Worker machines -- all in EC2.
    const infrastructure = new kelda.Infrastructure({
      masters: virtualMachine,
      workers: virtualMachine.replicate(6),
    });

    // Deploy the app to the infrastructure.
    app.deploy(infrastructure);
    ```

3. **Boot the application to your preferred cloud using Kelda's CLI.**

    ```console
    $ kelda run ./myBlueprint.js
    Your blueprint is being deployed. Check its status with `kelda show`.
    ```

4. **Manage your deployment from the command line.**

    Keep track of virtual machines and containers.

    ```console
    $ kelda show
    MACHINE         ROLE      PROVIDER    REGION       SIZE         PUBLIC IP         STATUS
    i-0d1d363a57    Master    Amazon      us-west-1    m3.medium    54.149.85.24      connected
    i-0b6ef239c7    Worker    Amazon      us-west-1    m3.medium    54.218.19.34      connected
    ...

    CONTAINER       MACHINE         COMMAND          HOSTNAME    STATUS     CREATED           PUBLIC IP
    bce2362f3a1f    i-0b6ef239c7                     my-app      running    20 seconds ago    54.218.19.34:80
    ...
    ```

    SSH into containers and machines (here, the `my-app` container).
    ```console
    $ kelda ssh my-app
    ```

    Stop the deployment.
    ```console
    $ kelda stop
    ```

    And more...

Learn more in our [Quick Start](http://docs.kelda.io/#quick-start) tutorial.

# API
Here are a some of the cool things you can do with Kelda's API.

### Share and Reuse Configuration
Don't waste time combing through reams of application specific documentation.
Share and reuse configuration like you do with other software libraries.

[//]: # (b1)
```javascript
const Redis = require('@kelda/redis');
const redisDB = new Redis(3, 'ADMIN_PASSWORD'); // A Redis database with 3 workers.
```

### Configure a Secure Network
Only allow the traffic your application needs.

[//]: # (b1)
<!-- redisDB.deploy(infrastructure); -->
```javascript
// Let the app send traffic to the Redis database on port 6379.
kelda.allowTraffic(app, redisDB.master, 6379);
```

### Run Containers
Quickly get running in the cloud.

[//]: # (b1)
```javascript
const otherApp = new kelda.Container({
  name: 'my-other-app',
  image: 'myOtherImage',
});
```

### Keep Sensitive Data Safe
We keep your keys, certificates, and passwords encrypted and in the right hands.

[//]: # (b1)
```javascript
otherApp.env['GITHUB_OATH_TOKEN'] = new kelda.Secret('githubOathToken');
```

### Deploy Anywhere
Kelda's API is cloud agnostic, so you, your users, and
colleagues can run the same blueprint in any of Kelda's supported clouds.

[//]: # (b1)
```javascript
const machine = new kelda.Machine({
  provider: 'Google', // 'Amazon', 'DigitalOcean'.
  size: 'n1-standard-2', // 'm3.medium', '1gb'.
});
```
See the full [API reference](http://docs.kelda.io/#kelda-js-api-documentation) in our docs.

# Reference Blueprints
Kelda blueprints have been written for many popular applications. Here are some that we recommend checking out:

* [Apache Spark](https://github.com/kelda/spark). Deploy a multi-node Spark cluster for data processing.
* [HAProxy](https://github.com/kelda/haproxy/blob/master/examples). Load balance traffic across different applications.
* [Kibana](https://github.com/kelda/kibana). Vizualize data from a hosted Elasticsearch instance.

# Get Involved
We would love to work with you, get your feedback, or answer any questions you have!
* If you have any questions or want to introduce yourself, come [join our Slack](http://slack.kelda.io/).
* If you have a suggestion or find a problem, please [submit a GitHub issue](https://github.com/kelda/kelda/issues).
* If you want to help make the cloud approachable, we would love to see your [pull requests](https://github.com/kelda/kelda/pulls). Check out our docs for [how to get started contributing](http://docs.kelda.io/#contributing-code).
* Want to learn more? Check out our [docs](http://docs.kelda.io/) and [website](http://kelda.io/).
