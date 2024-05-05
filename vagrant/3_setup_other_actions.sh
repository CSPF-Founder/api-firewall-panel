#!/bin/bash

#Install other depedencies
sudo apt-get install make -y 


#Make app dirs
mkdir -p /app/bin/
mkdir -p /app/panel/src

#Other Depedencies
sudo apt install net-tools sudo wget nano telnet resolvconf ntpdate vim -y

#Set time
sudo ntpdate pool.ntp.org


# Other configs
sudo sed -i 's/#SystemMaxUse=/SystemMaxUse=100M/g' /etc/systemd/journald.conf
systemctl restart systemd-journald

mkdir -p /app/panel/
mkdir -p /app/data/docker/
mkdir -p /app/data/input/

sudo chown -R vagrant:vagrant /app/
sudo chmod -R 777 /app/