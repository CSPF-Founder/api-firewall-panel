#!/bin/bash

mkdir -p /app/clibuild/
cp -a /vagrant/code/cli/. /app/clibuild/
cd /app/clibuild/
make
cp /app/clibuild/bin/apcli /app/bin/apcli
chmod +x /app/bin/apcli


echo 'DOCKER_IMAGE_TAG = api-firewall' >> /app/bin/.env
echo 'DOCKER_DATA_DIR = "/app/data/docker/"' >> /app/bin/.env
