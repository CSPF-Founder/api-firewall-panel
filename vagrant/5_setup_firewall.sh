#!/bin/bash
# Pull git and make docker image
mkdir -p /app/firewallbuild/

cd ~ 
wget https://github.com/wallarm/api-firewall/archive/refs/tags/v0.7.2.tar.gz
tar -C /app/firewallbuild/ -xzf v0.7.2.tar.gz
cd /app/firewallbuild/api-firewall-0.7.2/
sudo rm /app/firewallbuild/api-firewall-0.7.2/docker-entrypoint.sh
#Patch code
cp -a /vagrant/code/customfirewallentry/docker-entrypoint.sh /app/firewallbuild/api-firewall-0.7.2/docker-entrypoint.sh

docker build . --tag api-firewall