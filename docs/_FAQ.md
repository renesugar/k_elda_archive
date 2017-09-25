# Frequently Asked Questions

This section includes answers to common questions about Quilt, and solutions
to various issues.  If you run into an issue and can't find the answer here,
don't hesitate to email us at [discuss@quilt.io](mailto:discuss@quilt.io).

### What does Quilt use SSH keys for?
Quilt `Machine`s optionally take one or more public SSH keys. It is strongly
recommended to always provide at least one SSH key, as this will allow you
to SSH into VMs and containers, and enables useful Quilt CLI commands like
`quilt ssh` and `quilt logs` from any computer that holds a private SSH key
matching the public key set on the `Machine`.

### I tried to `quilt run` a blueprint on Amazon and nothing seems to be working.

If you're running a blueprint on AWS and the containers are not getting properly
created, you may have an issue with your VPC (Virtual Private Cloud) settings
on Amazon.  When this issue occurs, if you run `quilt show`, the machines will
all have status `connected`, but the containers will never progress to the
`scheduled` state (either the status will be empty, or for Dockerfiles that are
built in the cluster, the status will say `built`).  This issue only occurs
if you've changed your default VPC on Amazon, so if you don't know what a VPC is
or you haven't used one before on Amazon, this is probably not the issue you're
experiencing.

If you previously changed your default VPC on Amazon, it may be configured in a
way that prevents Quilt containers from properly communicating with each other.
This happens if the subnet configured for your VPC overlaps with 10.0.0.0/8,
which is the subnet that Quilt uses. This problem can manifest in many ways;
typically it looks like nothing seems to be working correctly.  For example, a
recent user who experienced this problem saw the following logs on the etcd
container on a Quilt worker:

```console
2017-06-05 21:14:40.823760 W | etcdserver: could not get cluster response from http://10.201.160.199:2380: Get http://10.201.160.199:2380/members: dial tcp 10.201.160.199:2380: getsockopt: no route to host
2017-06-05 21:14:40.823787 W | etcdmain: proxy: could not retrieve cluster information from the given urls
```

You can check (and fix) your VPC settings in the
[VPC section of the online AWS console](http://console.aws.amazon.com/vpc).
