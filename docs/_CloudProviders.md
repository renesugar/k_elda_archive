# Cloud Provider Configuration

The [Getting Started Guide](#getting-started) described how the basics of
setting up Amazon EC2.  This section describes the basic configuration of the
other cloud providers, and gives some details about how to enable extra features
(e.g., floating IP addresses) on each cloud provider.

## DigitalOcean

### Basic Setup

1. Create a new key [here](https://cloud.digitalocean.com/settings/api/tokens).
   Both read and write permissions are required.

2. Save the key in `~/.digitalocean/key` on the machine that will be running the
   Quilt daemon.

Now, to deploy a DigitalOcean droplet in the `sfo1` zone of size `512mb` as a
`Worker`:

```javascript
deployment.deploy(new Machine({
    provider: "DigitalOcean",
    region: "sfo1",
    size: "512mb",
    role: "Worker" }));
```

### Floating IPs

Creating a floating IP is slightly unintuitive. Unless there are already
droplets running, the floating IP tab under "Networking" doesn't allow users to
create floating IPs. However, [this
link](https://cloud.digitalocean.com/networking/floating_ips/datacenter) can be
used to reserve IPs for a specific datacenter. If that link breaks, floating IPs
can always be created by creating a droplet, _then_ assigning it a new floating
IP. The floating IP will still be reserved for use after the associated droplet
is removed.

Note that DigitalOcean charges a small hourly fee for floating IPs that have
been reserved, but are not associated with a droplet.

## Google Compute Engine

### Basic Setup

1. Create a Google Cloud Platform Project: All instances are booted under a
   Cloud Platform project. To setup a project for use with Quilt, go to the
   [console page](http://console.cloud.google.com), then click the project
   dropdown at the top of page, and hit the plus icon. Pick a name, and create
   your project.

2. Enable the Compute API: Select your newly created project from the project
   selector at the top of the [console page](http://console.cloud.google.com),
   and then select `API Manager -> Library` from the navbar on the left. Search
   for and enable the `Google Compute Engine API`.

3. Save the Credentials File: Go to `Credentials` on the left navbar (under `API
   Manager`), and create credentials for a `Service account key`. Create a new
   service account with the `Project -> Editor` role, and select the JSON output
   option. Copy the downloaded file to `~/.gce/quilt.json` on the machine from
   which you will be running the Quilt daemon.

Now, to deploy a GCE machine in the `us-east1-b` zone of size
`n1-standard-1` as a `Worker`:

```javascript
deployment.deploy(new Machine({
    provider: "Google",
    region: "us-east1-b",
    size: "n1-standard-1",
    role: "Worker" }));
```
