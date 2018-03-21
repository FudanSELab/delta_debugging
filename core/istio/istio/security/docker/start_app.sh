#!/bin/bash

echo "Starting NodeAgent..."
# Run node-agent
/usr/local/bin/node_agent \
  --env onprem \
  --onprem-cert-chain /usr/local/bin/node_agent.crt \
  --onprem-key /usr/local/bin/node_agent.key \
  --onprem-root-cert /usr/local/bin/istio_ca.crt&

echo "Starting Application..."
# Start app
apt-get update
apt-get -y install curl
curl -sL https://deb.nodesource.com/setup_8.x | bash -
apt-get install -y nodejs
npm install express
node /usr/local/bin/app.js
