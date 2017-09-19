# Cloud Provider Configuration

This section describes the basic configuration of the cloud providers supported
by Quilt, and gives some details about how to enable extra features (e.g.,
floating IP addresses) on each cloud provider.

## Amazon EC2

For Amazon EC2, you'll first need to create an account with [Amazon Web
Services](https://aws.amazon.com/ec2/) and then find your access credentials
from the [Security Credentials](https://console.aws.amazon.com/iam/home?#security_credential)
page in the AWS Management Console. Once you've done that, put your Amazon
credentials in a file called `~/.aws/credentials`:

```conf
[default]
aws_access_key_id = <YOUR_ID>
aws_secret_access_key = <YOUR_SECRET_KEY>
```

The file needs to appear exactly as above (including the `[default]` at the
top), except with `<YOUR_ID>` and `<YOUR_SECRET_KEY>` filled in appropriately.

## DigitalOcean

### Basic Setup

Quilt needs access to a DigitalOcean account token in order to make the API
calls needed to boot your deployment. To get and set up this token:

1. Create a new token [here](https://cloud.digitalocean.com/settings/api/tokens).
   The token must have both read and write permissions.

2. Save the token in `~/.digitalocean/key` on the machine that will be running
   the Quilt daemon.

### Floating IPs

Unless there are already droplets running, DigitalOcean doesn't allow users to
create floating IPs under the "Networking" tab on their website. Instead, [this
link](https://cloud.digitalocean.com/networking/floating_ips/datacenter) can be
used to reserve IPs that Quilt can then assign to droplets.

## Google Compute Engine

### Basic Setup

1. Create a Google Cloud Platform Project: All instances are booted under a
   Cloud Platform project. To setup a project for use with Quilt, go to the
   [console page](http://console.cloud.google.com), then click the project
   dropdown at the top of page, and hit the plus icon. Pick a name, and create
   your project.

2. Enable the Compute API: Select your newly created project from the project
   selector at the top of the [console page](http://console.cloud.google.com),
   and then select `APIs & services -> Library` from the navbar on the left. Search
   for and enable the `Google Compute Engine API`.

3. Save the Credentials File: Go to `Credentials` on the left navbar (under `APIs
   & services`), and create credentials for a `Service account key`. Create a new
   service account with the `Project -> Editor` role, and select the JSON output
   option. Copy the downloaded file to `~/.gce/quilt.json` on the machine from
   which you will be running the Quilt daemon.
