#!/bin/bash
./network.sh up createChannel -c inpoinchannel -ca -s couchdb

cd addOrg3

./addOrg3.sh up -c inpoinchannel -ca -s couchdb

cd ../