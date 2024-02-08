#!/bin/bash

./network.sh down
./network.sh up
./network.sh createChannel
./network.sh deployCC -ccp ../chaincode/ -ccn bankchaincode1 -c channel1
#./network.sh deployCC -ccp ../chaincode/ -ccn bankchaincode2 -c channel2 -ccep "AND('Org1.member', 'Org2.member', 'Org3.member', 'Org4.member')"

# Init ledger
cd utils
./init_ledger.sh
