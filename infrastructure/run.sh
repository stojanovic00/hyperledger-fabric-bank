#!/bin/bash
./network.sh down
./network.sh up
./network.sh createChannel
./network.sh deployCC -ccp ../chaincode/
