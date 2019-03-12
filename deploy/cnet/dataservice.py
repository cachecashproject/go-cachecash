#!/usr/bin/env python

"""
This is the most simple example to showcase Containernet.
"""


from mininet.net import Containernet
from mininet.node import Controller
from mininet.cli import CLI
from mininet.link import TCLink

from mininet.log import info, setLogLevel
setLogLevel('debug')


class Network():
    def __init__(self):
        self.containers = {}
        self.switches = {}
        
        self.cn = Containernet(controller=Controller)
    
        info('*** Adding controller\n')
        self.cn.addController('c0')
        
        info('*** Adding switches\n')
        self.switches['sw0'] = self.cn.addSwitch('sw0')

        info('*** Adding docker containers\n')
        self.containers['u0'] = self.cn.addDocker('u0', ip='10.0.0.10', dimage="ubuntu:trusty")
        self.containers['u1'] = self.cn.addDocker('u1', ip='10.0.0.11', dimage="ubuntu:trusty")
        self.containers['p0'] = self.cn.addDocker('p0', ip='10.0.0.100', dimage="cachecashproject/go-cachecash", dcmd='')
        
        info('*** Creating links\n')
        for c in self.containers.values():
            self.cn.addLink(c, self.switches['sw0'])


def main():
    net = Network()
    
    info('*** Starting network\n')
    net.cn.start()
    info('*** Testing connectivity\n')
    net.cn.ping([net.containers['u0'], net.containers['u1']])
    net.cn.ping([net.containers['u0'], net.containers['p0']])
    info('*** Running CLI\n')
    CLI(net.cn)
    info('*** Stopping network')
    net.cn.stop()


if __name__ == '__main__':
    main()
    
