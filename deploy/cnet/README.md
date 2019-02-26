# Containernet environment

[Containernet](https://github.com/containernet/containernet) is a [Mininet](https://github.com/mininet/mininet) fork
that supports using Docker containers as hosts in a simulated network.

There can be lots of moving pieces in a CacheCash environment:
- Many participants (publishers, caches, clients), each of which may be horizontally scalable.
- The relational databases and key-value stores that those participants use.
- Centralized log and metric collection services.
- The ledger (either simulated or itself composed of many participants).

The resources in this subdirectory make it easy to reproducibly run those pieces in different configurations.

## Environment

Mininet uses Open vSwitch, a software OpenFlow implementation.  Without it available, switch nodes in the Mininet
topology will drop traffic.  To install it on an Ubuntu system:

```
apt-get install openvswitch-switch
```

## Running

Containernet itself is run in a privileged container.

```
# Run this command from this directory.
$ docker run --name containernet -it --rm --privileged --pid=host -v $PWD:/cachecash-cnet -v /var/run/docker.sock:/var/run/docker.sock containernet/containernet:d764c67ec639 /cachecash-cnet/dataservice.py
```
