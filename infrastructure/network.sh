#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

# This script brings up a Hyperledger Fabric network for testing smart contracts
# and applications. The test network consists of two organizations with one
# peer each, and a single node Raft ordering service. Users can also use this
# script to create a channel deploy a chaincode on the channel
#
# prepending $PWD/../bin to PATH to ensure we are picking up the correct binaries
# this may be commented out to resolve installed version of tools if desired
export PATH=${PWD}/bin:$PATH
export FABRIC_CFG_PATH=${PWD}/configtx
export VERBOSE=false
export CONFIG_PATH=${PWD}/config
export SCRIPTS_PATH=${PWD}/scripts
export ORGANIZATIONS_PATH=${PWD}/organizations
export ORGANIZATION_NUMBER=4
export PEER_PER_ORGANIZATION_NUMBER=1
export DOCKER_PATH=${PWD}/docker
export CHANNEL_ARTIFACTS_PATH=${PWD}/channel-artifacts
export SYSTEM_GENESIS_BLOCK_PATH=${PWD}/system-genesis-block
export DOCKER_COMPOSE_FILE_PATH=${DOCKER_PATH}/peers_and_orderer-compose.yaml
export CRYPTOGEN_PATH=${ORGANIZATIONS_PATH}/cryptogen
export DOCKER_COMPOSE_CA_FILE_PATH=${DOCKER_PATH}/ca-compose.yaml

. ${SCRIPTS_PATH}/utils.sh

# Obtain CONTAINER_IDS and remove them
# TODO Might want to make this optional - could clear other containers
# This function is called when you bring a network down
function clearContainers() {
  CONTAINER_IDS=$(docker ps -a | awk '($2 ~ /dev-peer.*/) {print $1}')
  if [ -z "$CONTAINER_IDS" -o "$CONTAINER_IDS" == " " ]; then
    infoln "No containers available for deletion"
  else
    docker rm -f $CONTAINER_IDS
  fi
}

# Delete any images that were generated as a part of this setup
# specifically the following images are often left behind:
# This function is called when you bring the network down
function removeUnwantedImages() {
  DOCKER_IMAGE_IDS=$(docker images | awk '($1 ~ /dev-peer.*/) {print $3}')
  if [ -z "$DOCKER_IMAGE_IDS" -o "$DOCKER_IMAGE_IDS" == " " ]; then
    infoln "No images available for deletion"
  else
    docker rmi -f $DOCKER_IMAGE_IDS
  fi
}

# Versions of fabric known not to work with the test network
NONWORKING_VERSIONS="^1\.0\. ^1\.1\. ^1\.2\. ^1\.3\. ^1\.4\."

# Do some basic sanity checking to make sure that the appropriate versions of fabric
# binaries/images are available. In the future, additional checking for the presence
# of go or other items could be added.


function checkPrereqs() {
  ## Check if your have cloned the peer binaries and configuration files.
  peer version > /dev/null 2>&1

  if [[ $? -ne 0 || ! -d ${CONFIG_PATH} ]]; then
    errorln "Peer binary and configuration files not found.."
    errorln
    errorln "Follow the instructions in the Fabric docs to install the Fabric Binaries:"
    errorln "https://hyperledger-fabric.readthedocs.io/en/latest/install.html"
    exit 1
  fi
  # use the fabric tools container to see if the samples and binaries match your
  # docker images
  LOCAL_VERSION=$(peer version | sed -ne 's/^ Version: //p')
  DOCKER_IMAGE_VERSION=$(docker run --rm hyperledger/fabric-tools:$IMAGETAG peer version | sed -ne 's/^ Version: //p')

  infoln "LOCAL_VERSION=$LOCAL_VERSION"
  infoln "DOCKER_IMAGE_VERSION=$DOCKER_IMAGE_VERSION"

  if [ "$LOCAL_VERSION" != "$DOCKER_IMAGE_VERSION" ]; then
    warnln "Local fabric binaries and docker images are out of  sync. This may cause problems."
  fi

  for UNSUPPORTED_VERSION in $NONWORKING_VERSIONS; do
    infoln "$LOCAL_VERSION" | grep -q $UNSUPPORTED_VERSION
    if [ $? -eq 0 ]; then
      fatalln "Local Fabric binary version of $LOCAL_VERSION does not match the versions supported by the test network."
    fi

    infoln "$DOCKER_IMAGE_VERSION" | grep -q $UNSUPPORTED_VERSION
    if [ $? -eq 0 ]; then
      fatalln "Fabric Docker image version of $DOCKER_IMAGE_VERSION does not match the versions supported by the test network."
    fi
  done

  

    fabric-ca-client version > /dev/null 2>&1
    if [[ $? -ne 0 ]]; then
      errorln "fabric-ca-client binary not found.."
      errorln
      errorln "Follow the instructions in the Fabric docs to install the Fabric Binaries:"
      errorln "https://hyperledger-fabric.readthedocs.io/en/latest/install.html"
      exit 1
    fi
    CA_LOCAL_VERSION=$(fabric-ca-client version | sed -ne 's/ Version: //p')
    CA_DOCKER_IMAGE_VERSION=$(docker run --rm hyperledger/fabric-ca:$CA_IMAGETAG fabric-ca-client version | sed -ne 's/ Version: //p' | head -1)
    infoln "CA_LOCAL_VERSION=$CA_LOCAL_VERSION"
    infoln "CA_DOCKER_IMAGE_VERSION=$CA_DOCKER_IMAGE_VERSION"

    if [ "$CA_LOCAL_VERSION" != "$CA_DOCKER_IMAGE_VERSION" ]; then
      warnln "Local fabric-ca binaries and docker images are out of sync. This may cause problems."
    fi
  
}

# Before you can bring up a network, each organization needs to generate the crypto
# material that will define that organization on the network. Because Hyperledger
# Fabric is a permissioned blockchain, each node and user on the network needs to
# use certificates and keys to sign and verify its actions. In addition, each user
# needs to belong to an organization that is recognized as a member of the network.
# You can use the Cryptogen tool or Fabric CAs to generate the organization crypto
# material.

# By default, the sample network uses cryptogen. Cryptogen is a tool that is
# meant for development and testing that can quickly create the certificates and keys
# that can be consumed by a Fabric network. The cryptogen tool consumes a series
# of configuration files for each organization in the "organizations/cryptogen"
# directory. Cryptogen uses the files to generate the crypto  material for each
# org in the "organizations" directory.

# You can also Fabric CAs to generate the crypto material. CAs sign the certificates
# and keys that they generate to create a valid root of trust for each organization.
# The script uses Docker Compose to bring up three CAs, one for each peer organization
# and the ordering organization. The configuration file for creating the Fabric CA
# servers are in the "organizations/fabric-ca" directory. Within the same directory,
# the "registerEnroll.sh" script uses the Fabric CA client to create the identities,
# certificates, and MSP folders that are needed to create the test network in the
# "organizations/ordererOrganizations" directory.

function createOrgs() {
  if [ -d "${ORGANIZATIONS_PATH}/peerOrganizations" ]; then
    rm -Rf ${ORGANIZATIONS_PATH}/peerOrganizations && rm -Rf ${ORGANIZATIONS_PATH}/ordererOrganizations
  fi

  # Create crypto material using Fabric CA
  infoln "Generating certificates using Fabric CA"

  #generate docker compose file for CAs and start the containers 
  startCAContainers

  . ${ORGANIZATIONS_PATH}/fabric-ca/registerEnroll.sh

  for ((n=1;n<=$ORGANIZATION_NUMBER;++n))
  do
    infoln "Creating Org$n Identities"

    while [ ! -f "${ORGANIZATIONS_PATH}/fabric-ca/org$n/tls-cert.pem" ];
      do
        sleep 1
      done

    createOrgN $n

  done

  infoln "Creating Orderer Org Identities"

  createOrderer  

  infoln "Generating docker-compose file for the network"
  generatePeersAndOrdererComposeFile

  infoln "Generating CCP files for Org$ORGANIZATION_NUMBER"
  ${ORGANIZATIONS_PATH}/ccp-generate.sh $ORGANIZATION_NUMBER $PEER_PER_ORGANIZATION_NUMBER
}

function startCAContainers(){
  echo "version: '2.1'

networks:
  bank_network:

services:

  ca_orderer:
    image: hyperledger/fabric-ca:1.5.7
    environment:
      - FABRIC_CA_HOME=/etc/hyperledger/fabric-ca-server
      - FABRIC_CA_SERVER_CA_NAME=ca-orderer
      - FABRIC_CA_SERVER_TLS_ENABLED=true
      - FABRIC_CA_SERVER_PORT=5000
      - FABRIC_CA_SERVER_OPERATIONS_LISTENADDRESS=0.0.0.0:15000
    ports:
      - 5000:5000
      - 15000:15000
    command: sh -c 'fabric-ca-server start -b admin:adminpw -d'
    volumes:
      - ${ORGANIZATIONS_PATH}/fabric-ca/ordererOrg:/etc/hyperledger/fabric-ca-server
    container_name: ca_orderer
    networks:
      - bank_network
" > $DOCKER_COMPOSE_CA_FILE_PATH

    for (( i = 1; i<= $ORGANIZATION_NUMBER; i++))
    do
      echo "  ca_org$i:
    image: hyperledger/fabric-ca:1.5.7
    environment:
      - FABRIC_CA_HOME=/etc/hyperledger/fabric-ca-server
      - FABRIC_CA_SERVER_CA_NAME=ca-org$i
      - FABRIC_CA_SERVER_TLS_ENABLED=true
      - FABRIC_CA_SERVER_PORT=$((6 + $i))150
      - FABRIC_CA_SERVER_OPERATIONS_LISTENADDRESS=0.0.0.0:$((6 + $i))199
    ports:
      - $((6 + $i))150:$((6 + $i))150
      - $((6 + $i))199:$((6 + $i))199
    command: sh -c 'fabric-ca-server start -b admin:adminpw -d'
    volumes:
      - ${ORGANIZATIONS_PATH}/fabric-ca/org$i:/etc/hyperledger/fabric-ca-server
    container_name: ca_org$i
    networks:
      - bank_network
" >> $DOCKER_COMPOSE_CA_FILE_PATH
    done
  IMAGE_TAG=${CA_IMAGETAG} docker-compose -f $DOCKER_COMPOSE_CA_FILE_PATH up -d 2>&1
}

function generatePeersAndOrdererComposeFile(){
  echo 'version: "2.1"

volumes:
  orderer.example.com:' > $DOCKER_COMPOSE_FILE_PATH
  for ((i = 1 ; i <= $ORGANIZATION_NUMBER; ++i)) do
    for ((j = 0; j <= $PEER_PER_ORGANIZATION_NUMBER; j++)) do
      echo "  peer$j.org$i.example.com:" >> $DOCKER_COMPOSE_FILE_PATH
    done
  done
  echo "
networks:
  bank_network:

services:

  orderer.example.com:
    container_name: orderer.example.com
    image: hyperledger/fabric-orderer:2.2.6
    environment:
      - FABRIC_LOGGING_SPEC=INFO
      - ORDERER_GENERAL_LISTENADDRESS=0.0.0.0
      - ORDERER_GENERAL_LISTENPORT=7000
      - ORDERER_GENERAL_GENESISMETHOD=file
      - ORDERER_GENERAL_GENESISFILE=/var/hyperledger/orderer/orderer.genesis.block
      - ORDERER_GENERAL_LOCALMSPID=OrdererMSP
      - ORDERER_GENERAL_LOCALMSPDIR=/var/hyperledger/orderer/msp
      - ORDERER_OPERATIONS_LISTENADDRESS=orderer.example.com:9900
      # enabled TLS
      - ORDERER_GENERAL_TLS_ENABLED=true
      - ORDERER_GENERAL_TLS_PRIVATEKEY=/var/hyperledger/orderer/tls/server.key
      - ORDERER_GENERAL_TLS_CERTIFICATE=/var/hyperledger/orderer/tls/server.crt
      - ORDERER_GENERAL_TLS_ROOTCAS=[/var/hyperledger/orderer/tls/ca.crt]
      - ORDERER_KAFKA_TOPIC_REPLICATIONFACTOR=1
      - ORDERER_KAFKA_VERBOSE=true
      - ORDERER_GENERAL_CLUSTER_CLIENTCERTIFICATE=/var/hyperledger/orderer/tls/server.crt
      - ORDERER_GENERAL_CLUSTER_CLIENTPRIVATEKEY=/var/hyperledger/orderer/tls/server.key
      - ORDERER_GENERAL_CLUSTER_ROOTCAS=[/var/hyperledger/orderer/tls/ca.crt]
    working_dir: /opt/gopath/src/github.com/hyperledger/fabric
    command: orderer
    volumes:
        - ${SYSTEM_GENESIS_BLOCK_PATH}/genesis.block:/var/hyperledger/orderer/orderer.genesis.block
        - ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/orderers/orderer.example.com/msp:/var/hyperledger/orderer/msp
        - ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/orderers/orderer.example.com/tls/:/var/hyperledger/orderer/tls
        - orderer.example.com:/var/hyperledger/production/orderer
    ports:
      - 7000:7000
      - 9900:9900
    networks:
      - bank_network
" >>$DOCKER_COMPOSE_FILE_PATH

      for ((i = 1; i <= $ORGANIZATION_NUMBER ; ++i)) do
        for ((j = 0; j < $PEER_PER_ORGANIZATION_NUMBER; ++j)) do
          echo "  peer$j.org$i.example.com:
    container_name: peer$j.org$i.example.com
    image: hyperledger/fabric-peer:2.2.6
    environment:
      #Generic peer variables
      - CORE_VM_ENDPOINT=unix:///host/var/run/docker.sock
      - CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=docker_bank_network
      - FABRIC_LOGGING_SPEC=INFO
      #- FABRIC_LOGGING_SPEC=DEBUG
      - CORE_PEER_TLS_ENABLED=true
      - CORE_PEER_PROFILE_ENABLED=true
      - CORE_PEER_TLS_CERT_FILE=/etc/hyperledger/fabric/tls/server.crt
      - CORE_PEER_TLS_KEY_FILE=/etc/hyperledger/fabric/tls/server.key
      - CORE_PEER_TLS_ROOTCERT_FILE=/etc/hyperledger/fabric/tls/ca.crt
      - CORE_LEDGER_STATE_STATEDATABASE=CouchDB
      - CORE_LEDGER_STATE_COUCHDBCONFIG_COUCHDBADDRESS=couchdb$i$j:5984
      - CORE_LEDGER_STATE_COUCHDBCONFIG_USERNAME=admin
      - CORE_LEDGER_STATE_COUCHDBCONFIG_PASSWORD=adminpw
      - CORE_PEER_ID=peer$j.org$i.example.com
      - CORE_PEER_ADDRESS=peer$j.org$i.example.com:$((6 + $i))05$j
      - CORE_PEER_LISTENADDRESS=0.0.0.0:$((6 + $i))05$j
      - CORE_PEER_CHAINCODEADDRESS=peer$j.org$i.example.com:$((6 + $i))25$j
      - CORE_PEER_CHAINCODELISTENADDRESS=0.0.0.0:$((6 + $i))25$j
      - CORE_PEER_GOSSIP_BOOTSTRAP=peer$j.org$i.example.com:$((6 + $i))05$j
      - CORE_PEER_GOSSIP_EXTERNALENDPOINT=peer$j.org$i.example.com:$((6 + $i))05$j
      - CORE_PEER_LOCALMSPID=Org${i}MSP
      - CORE_OPERATIONS_LISTENADDRESS=peer$j.org$i.example.com:99$i$j
    volumes:
        - /var/run/docker.sock:/host/var/run/docker.sock
        - ${ORGANIZATIONS_PATH}/peerOrganizations/org$i.example.com/peers/peer$j.org$i.example.com/msp:/etc/hyperledger/fabric/msp
        - ${ORGANIZATIONS_PATH}/peerOrganizations/org$i.example.com/peers/peer$j.org$i.example.com/tls:/etc/hyperledger/fabric/tls
        - peer$j.org$i.example.com:/var/hyperledger/production
    working_dir: /opt/gopath/src/github.com/hyperledger/fabric/peer
    command: peer node start
    ports:
      - $((6 + $i))05$j:$((6 + $i))05$j
      - 99$i$j:99$i$j
    networks:
      - bank_network
    depends_on:
      - couchdb$i$j

  couchdb$i$j:
    container_name: couchdb$i$j
    image: couchdb:latest
    environment:
      - COUCHDB_USER=admin
      - COUCHDB_PASSWORD=adminpw
    networks:
      - bank_network
" >>$DOCKER_COMPOSE_FILE_PATH
      done
    done

      echo "  cli:
    container_name: cli
    image: hyperledger/fabric-tools:2.2.6
    tty: true
    stdin_open: true
    environment:
      - GOPATH=/opt/gopath
      - CORE_VM_ENDPOINT=unix:///host/var/run/docker.sock
      - FABRIC_LOGGING_SPEC=INFO
      #- FABRIC_LOGGING_SPEC=DEBUG
    working_dir: /opt/gopath/src/github.com/hyperledger/fabric/peer
    command: /bin/bash
    volumes:
        - /var/run/:/host/var/run/
        - ${ORGANIZATIONS_PATH}:/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations
        - ${SCRIPTS_PATH}:/opt/gopath/src/github.com/hyperledger/fabric/peer/scripts/
    networks:
      - bank_network
    depends_on:
      - orderer.example.com" >>$DOCKER_COMPOSE_FILE_PATH
    for (( i = 1; i <= $ORGANIZATION_NUMBER; ++i)) do
      for (( j = 0; j < $PEER_PER_ORGANIZATION_NUMBER; ++j)) do
        echo "      - peer$j.org$i.example.com">>$DOCKER_COMPOSE_FILE_PATH
      done
    done
}

# Once you create the organization crypto material, you need to create the
# genesis block of the orderer system channel. This block is required to bring
# up any orderer nodes and create any application channels.

# The configtxgen tool is used to create the genesis block. Configtxgen consumes a
# "configtx.yaml" file that contains the definitions for the sample network. The
# genesis block is defined using the "FourOrgsOrdererGenesis" profile at the bottom
# of the file. This profile defines a sample consortium, "SampleConsortium",
# consisting of our two Peer Orgs. This consortium defines which organizations are
# recognized as members of the network. The peer and ordering organizations are defined
# in the "Profiles" section at the top of the file. As part of each organization
# profile, the file points to a the location of the MSP directory for each member.
# This MSP is used to create the channel MSP that defines the root of trust for
# each organization. In essence, the channel MSP allows the nodes and users to be
# recognized as network members. The file also specifies the anchor peers for each
# peer org. In future steps, this same file is used to create the channel creation
# transaction and the anchor peer updates.
#
#
# If you receive the following warning, it can be safely ignored:
#
# [bccsp] GetDefault -> WARN 001 Before using BCCSP, please call InitFactories(). Falling back to bootBCCSP.
#
# You can ignore the logs regarding intermediate certs, we are not using them in
# this crypto implementation.

# Generate orderer system channel genesis block.
function createConsortium() {
  which configtxgen
  if [ "$?" -ne 0 ]; then
    fatalln "configtxgen tool not found."
  fi

  infoln "Generating Orderer Genesis block"

  # Note: For some unknown reason (at least for now) the block file can't be
  # named orderer.genesis.block or the orderer will fail to launch!
  set -x
  configtxgen -profile FourOrgsOrdererGenesis -channelID system-channel -outputBlock ${SYSTEM_GENESIS_BLOCK_PATH}/genesis.block
  res=$?
  { set +x; } 2>/dev/null
  if [ $res -ne 0 ]; then
    fatalln "Failed to generate orderer genesis block..."
  fi
}

# After we create the org crypto material and the system channel genesis block,
# we can now bring up the peers and ordering service. By default, the base
# file for creating the network is "docker-compose-test-net.yaml" in the ``docker``
# folder. This file defines the environment variables and file mounts that
# point the crypto material and genesis block that were created in earlier.

# Bring up the peer and orderer nodes using docker compose.
function networkUp() {
  checkPrereqs

  #infoln "Generating artifacts"
  createOrgs
  createConsortium

  IMAGE_TAG=$IMAGETAG docker-compose -f ${DOCKER_COMPOSE_FILE_PATH} up -d 2>&1

  docker ps -a
  if [ $? -ne 0 ]; then
    fatalln "Unable to start network"
  fi
}


# call the script to create the channel, join the peers of org1 and org2,
# and then update the anchor peers for each organization
function createChannel() {
  # Bring up the network if it is not already up.

  if [ ! -d "${ORGANIZATIONS_PATH}/peerOrganizations" ]; then
    infoln "Bringing up network"
    networkUp
  fi

  # now run the script that creates a channel. This script uses configtxgen once
  # more to create the channel creation transaction and the anchor peer updates.
  # configtx.yaml is mounted in the cli container, which allows us to use it to
  # create the channel artifacts
  #scripts/createChannel.sh $CHANNEL_NAME $CLI_DELAY $MAX_RETRY $VERBOSE

  scripts/createChannel.sh channel1 $CLI_DELAY $MAX_RETRY $VERBOSE
  scripts/createChannel.sh channel2  $CLI_DELAY $MAX_RETRY $VERBOSE
}


## Call the script to deploy a chaincode to the channel
function deployCC() {
  scripts/deployCC.sh $CHANNEL_NAME $CC_NAME $CC_SRC_PATH $CC_SRC_LANGUAGE $CC_VERSION $CC_SEQUENCE $CC_INIT_FCN $CC_END_POLICY $CC_COLL_CONFIG $CLI_DELAY $MAX_RETRY $VERBOSE

  if [ $? -ne 0 ]; then
    fatalln "Deploying chaincode failed"
  fi
}


# Tear down running network
function networkDown() {
  # stop containers
  docker-compose -f $DOCKER_COMPOSE_FILE_PATH -f $DOCKER_COMPOSE_CA_FILE_PATH down --volumes --remove-orphans
  # Don't remove the generated artifacts -- note, the ledgers are always removed
  clearContainers
  #Cleanup images
  removeUnwantedImages
  # remove orderer block and other channel configuration transactions and certs
  docker run --rm -v $(pwd):/data busybox sh -c 'cd /data && rm -rf ${SYSTEM_GENESIS_BLOCK_PATH}/*.block ${ORGANIZATIONS_PATH}/peerOrganizations ${ORGANIZATIONS_PATH}/ordererOrganizations'
  ## remove fabric ca artifacts
  for ((i=1 ; i <= ${ORGANIZATION_NUMBER}; ++i))
  do
    docker run --rm -v $(pwd):/data busybox sh -c 'cd /data && rm -rf ${ORGANIZATIONS_PATH}/fabric-ca/org$i/msp organizations/fabric-ca/org$i/tls-cert.pem ${ORGANIZATIONS_PATH}/fabric-ca/org$i/ca-cert.pem ${ORGANIZATIONS_PATH}/fabric-ca/org$i/IssuerPublicKey ${ORGANIZATIONS_PATH}/fabric-ca/org$i/IssuerRevocationPublicKey ${ORGANIZATIONS_PATH}/fabric-ca/org$i/fabric-ca-server.db'
  done
  # remove channel and script artifacts
  docker run --rm -v $(pwd):/data busybox sh -c 'cd /data && rm -rf ${CHANNEL_ARTIFACTS_PATH} log.txt *.tar.gz'

  sudo rm -r \
  ${ORGANIZATIONS_PATH}/peerOrganizations/* \
  ${ORGANIZATIONS_PATH}/ordererOrganizations/* \
  ${ORGANIZATIONS_PATH}/fabric-ca/*rg* \
  ${PWD}/channel-artifacts/* \
  ${SYSTEM_GENESIS_BLOCK_PATH}/*
}

# Obtain the OS and Architecture string that will be used to select the correct
# native binaries for your platform, e.g., darwin-amd64 or linux-amd64
OS_ARCH=$(echo "$(uname -s | tr '[:upper:]' '[:lower:]' | sed 's/mingw64_nt.*/windows/')-$(uname -m | sed 's/x86_64/amd64/g')" | awk '{print tolower($0)}')
# Using crpto vs CA. default is cryptogen
CRYPTO="Certificate Authorities"
# timeout duration - the duration the CLI should wait for a response from
# another container before giving up
MAX_RETRY=5
# default for delay between commands
CLI_DELAY=3
# channel name defaults to "mychannel"
CHANNEL_NAME="channel1"
# chaincode name defaults to "NA"
CC_NAME="NA"
# chaincode path defaults to "NA"
CC_SRC_PATH="NA"
# endorsement policy defaults to "NA". This would allow chaincodes to use the majority default policy.
CC_END_POLICY="NA"
# collection configuration defaults to "NA"
CC_COLL_CONFIG="NA"
# chaincode init function defaults to "NA"
CC_INIT_FCN="NA"
# use this as the default docker-compose yaml definition
#COMPOSE_FILE_BASE=${DOCKER_PATH}/docker-compose-test-net.yaml
# docker-compose.yaml file if you are using couchdb
#COMPOSE_FILE_COUCH=${DOCKER_PATH}/docker-compose-couch.yaml
# certificate authorities compose file
#COMPOSE_FILE_CA=${DOCKER_COMPOSE_CA_FILE_PATH}

# chaincode language defaults to "NA"
CC_SRC_LANGUAGE="NA"
# Chaincode version
CC_VERSION="1.0"
# Chaincode definition sequence
CC_SEQUENCE=1
# default image tag
IMAGETAG="2.2.6"
# default ca image tag
CA_IMAGETAG="1.5.7"
# default database
DATABASE="couchdb"

# Parse commandline args

## Parse mode
if [[ $# -lt 1 ]] ; then
  printHelp
  exit 0
else
  MODE=$1
  shift
fi

# parse a createChannel subcommand if used
if [[ $# -ge 1 ]] ; then
  key="$1"
  if [[ "$key" == "createChannel" ]]; then
      export MODE="createChannel"
      shift
  fi
fi

# parse flags

while [[ $# -ge 1 ]] ; do
  key="$1"
  case $key in
  -h )
    printHelp $MODE
    exit 0
    ;;
  -c )
    CHANNEL_NAME="$2"
    shift
    ;;
  -ca )
    CRYPTO="Certificate Authorities"
    ;;
  -r )
    MAX_RETRY="$2"
    shift
    ;;
  -d )
    CLI_DELAY="$2"
    shift
    ;;
  -s )
    DATABASE="$2"
    shift
    ;;
  -ccl )
    CC_SRC_LANGUAGE="$2"
    shift
    ;;
  -ccn )
    CC_NAME="$2"
    shift
    ;;
  -ccv )
    CC_VERSION="$2"
    shift
    ;;
  -ccs )
    CC_SEQUENCE="$2"
    shift
    ;;
  -ccp )
    CC_SRC_PATH="$2"
    shift
    ;;
  -ccep )
    CC_END_POLICY="$2"
    shift
    ;;
  -cccg )
    CC_COLL_CONFIG="$2"
    shift
    ;;
  -cci )
    CC_INIT_FCN="$2"
    shift
    ;;
  -i )
    IMAGETAG="$2"
    shift
    ;;
  -cai )
    CA_IMAGETAG="$2"
    shift
    ;;
  -verbose )
    VERBOSE=true
    shift
    ;;
  * )
    errorln "Unknown flag: $key"
    printHelp
    exit 1
    ;;
  esac
  shift
done

# Are we generating crypto material with this command?
if [ ! -d "organizations/peerOrganizations" ]; then
  CRYPTO_MODE="with crypto from '${CRYPTO}'"
else
  CRYPTO_MODE=""
fi

# Determine mode of operation and printing out what we asked for
if [ "$MODE" == "up" ]; then
  infoln "Starting nodes with CLI timeout of '${MAX_RETRY}' tries and CLI delay of '${CLI_DELAY}' seconds and using database '${DATABASE}' ${CRYPTO_MODE}"
elif [ "$MODE" == "createChannel" ]; then
  infoln "Creating channel '${CHANNEL_NAME}'."
  infoln "If network is not up, starting nodes with CLI timeout of '${MAX_RETRY}' tries and CLI delay of '${CLI_DELAY}' seconds and using database '${DATABASE} ${CRYPTO_MODE}"
elif [ "$MODE" == "down" ]; then
  infoln "Stopping network"
elif [ "$MODE" == "restart" ]; then
  infoln "Restarting network"
elif [ "$MODE" == "deployCC" ]; then
  infoln "deploying chaincode on channel '${CHANNEL_NAME}'"
else
  printHelp
  exit 1
fi

if [ "${MODE}" == "up" ]; then
  networkUp
elif [ "${MODE}" == "createChannel" ]; then
  createChannel
elif [ "${MODE}" == "deployCC" ]; then
  deployCC
elif [ "${MODE}" == "down" ]; then
  networkDown
else
  printHelp
  exit 1
fi
