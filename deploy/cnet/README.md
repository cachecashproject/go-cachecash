# Containernet environment

[Containernet](https://github.com/containernet/containernet) is a [Mininet](https://github.com/mininet/mininet) fork
that supports using Docker containers as hosts in a simulated network.

There can be lots of moving pieces in a CacheCash environment:
- Many participants (publishers, caches, clients), each of which may be horizontally scalable.
- The relational databases and key-value stores that those participants use.
- Centralized log and metric collection services.
- The ledger (either simulated or itself composed of many participants).

The resources in this subdirectory make it easy to reproducibly run those pieces in different configurations.

## Creating images

Containernet requires
that [several utilities](https://github.com/containernet/containernet/wiki/Container-Requirements-and-Compatibility) be
present in each container image.

For images based on Alpine Linux, add the following to the Dockerfile.  If the Dockerfile contains a `USER` directive,
these lines must be inserted before that.

```
# --------------------
# These steps are only necessary for Containernet.  Containernet expects a version of ifconfig that supports CIDR
# notation ("ifconfig up 1.2.3.4/8") but installs ifconfig to /bin which is after /sbin (where busybox ifconfig, which
# is Alpine's default, lives) in root's PATH.  /usr/local/sbin is the first element in that PATH, so symlinking things
# there fixes the problem.
RUN apk add --no-cache bash net-tools iproute2 iputils
RUN mkdir -p /usr/local/sbin
RUN for XX in ifconfig route netstat domainname hostname ypdomainname nisdomainname; do ln -s /bin/$XX /usr/local/sbin/; done
# --------------------
```

## Environment

Mininet uses Open vSwitch, a software OpenFlow implementation.  Without it available, switch nodes in the Mininet
topology will drop traffic.  To install it on an Ubuntu system:

```
apt-get install openvswitch-switch
```

## Running

Containernet itself is run in a privileged container.  We mirror container images from `containernet/containernet` to
`cachecashproject/containernet` because The upstream project does not seem to tag releases and Docker Hub preserves only
the last several untagged images.

Each of the following commands should be run the following prefix in place of the ellipsis (...).
```
docker run --name containernet -it --rm --privileged --pid=host -v $PWD:/cachecash-cnet -v /var/run/docker.sock:/var/run/docker.sock cachecashproject/containernet
```

```
# Run this command from this directory to start the simulated network.  Send SIGINT to tear down the containers.
$ ... /cachecash-cnet/dataservice.py

# Remove anything that is left over after an unclean exit.
$ ... mn -c
```
