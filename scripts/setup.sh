#!/bin/bash

#US, Europe, Asia 
#3 data centers per region
#Partition users 
#Users, history replication at 3
#Content replication factor 9

export CLUSTER=andy-geo-partitioning
roachprod create $CLUSTER -n 12 --clouds=aws --aws-machine-type-ssd=c5d.4xlarge --geo --aws-zones=us-east-1a,us-east-1b,us-east-1c,eu-west-2a,eu-west-2b,eu-west-2c,ap-northeast-1a,ap-northeast-1c,ap-northeast-1d,us-east-1a,eu-west-2a,ap-northeast-1a 
roachprod stage $CLUSTER:1-9 cockroach
roachprod start $CLUSTER:1-9
roachprod run $CLUSTER:10-12 'sudo apt-get update'&& \
roachprod run $CLUSTER:10-12 'sudo apt-get install -y apt-transport-https ca-certificates curl gnupg-agent software-properties-common' && \
roachprod run $CLUSTER:10-12 'curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -' && \
roachprod run $CLUSTER:10-12 'sudo apt-key fingerprint 0EBFCD88' && \
roachprod run $CLUSTER:10-12 'sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"' && \
roachprod run $CLUSTER:10-12 'sudo apt-get -y install docker-ce docker-ce-cli containerd.io'
roachprod run $CLUSTER:10-12 'sudo apt-get install -y haproxy'
roachprod run $CLUSTER:10-12 './cockroach gen haproxy --insecure --host=<address of any node>'
roachprod run $CLUSTER:10
# Edit the haproxy.cfg file, removing the server addresses for all but the us-east1 nodes.
haproxy -f haproxy.cfg &
roachprod run $CLUSTER:11
# Edit the haproxy.cfg file, removing the server addresses for all but the us-east1 nodes.
haproxy -f haproxy.cfg &
roachprod run $CLUSTER:12
# Edit the haproxy.cfg file, removing the server addresses for all but the us-east1 nodes.
haproxy -f haproxy.cfg &



