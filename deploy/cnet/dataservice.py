#!/usr/bin/env python

"""
This is the most simple example to showcase Containernet.
"""


from mininet.net import Containernet
from mininet.node import Controller
from mininet.cli import CLI
from mininet.link import TCLink

from mininet.log import info, setLogLevel
setLogLevel('info')


class Network():
    def __init__(self):
        self.containers = {}
        self.switches = {}
        
        self.cn = Containernet(controller=Controller)
    
        info('*** Adding controller\n')
        self.cn.addController('c0')
    
        info('*** Adding docker containers\n')
        self.containers['d1'] = self.cn.addDocker('d1', ip='10.0.0.251', dimage="ubuntu:trusty")
        self.containers['d2'] = self.cn.addDocker('d2', ip='10.0.0.252', dimage="ubuntu:trusty")
        
        info('*** Adding switches\n')
        self.switches['s1'] = self.cn.addSwitch('s1')
        # self.switches['s2'] = self.cn.addSwitch('s2')
        
        info('*** Creating links\n')
        self.cn.addLink(self.containers['d1'], self.switches['s1'])
        self.cn.addLink(self.containers['d2'], self.switches['s1'])
        # self.cn.addLink(self.switches['s1'], self.switches['s2'], cls=TCLink, delay='100ms', bw=1)
        # self.cn.addLink(self.switches['s2'], self.containers['d2'])


def main():
    net = Network()
    
    info('*** Starting network\n')
    net.cn.start()
    info('*** Testing connectivity\n')
    net.cn.ping([net.containers['d1'], net.containers['d2']])
    info('*** Running CLI\n')
    CLI(net.cn)
    info('*** Stopping network')
    net.cn.stop()


if __name__ == '__main__':
    main()
    
