#!/bin/bash

function createOrgN() {
  infoln "Enrolling the CA admin for org $1"
  mkdir -p ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/

  export FABRIC_CA_CLIENT_HOME=${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/
  caPort="$((6 + $1))150"

  set -x
  fabric-ca-client enroll -u https://admin:adminpw@localhost:${caPort} --caname ca-org$1 --tls.certfiles ${ORGANIZATIONS_PATH}/fabric-ca/org$1/tls-cert.pem
  { set +x; } 2>/dev/null

  echo "NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/localhost-${caPort}-ca-org$1.pem
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/localhost-${caPort}-ca-org$1.pem
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/localhost-${caPort}-ca-org$1.pem
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/localhost-${caPort}-ca-org$1.pem
    OrganizationalUnitIdentifier: orderer" >${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/msp/config.yaml

  for ((i = 0; i < $PEER_PER_ORGANIZATION_NUMBER; i++)) do
    infoln "Registering peer$i"
    set -x
    fabric-ca-client register --caname ca-org$1 --id.name peer$i --id.secret peer${i}pw --id.type peer --tls.certfiles ${ORGANIZATIONS_PATH}/fabric-ca/org$1/tls-cert.pem
    { set +x; } 2>/dev/null
  done

  infoln "Registering user"
  set -x
  fabric-ca-client register --caname ca-org$1 --id.name user1 --id.secret user1pw --id.type client --tls.certfiles ${ORGANIZATIONS_PATH}/fabric-ca/org$1/tls-cert.pem
  { set +x; } 2>/dev/null

  infoln "Registering the org admin"
  set -x
  fabric-ca-client register --caname ca-org$1 --id.name org${1}admin --id.secret org${1}adminpw --id.type admin --tls.certfiles ${ORGANIZATIONS_PATH}/fabric-ca/org$1/tls-cert.pem
  { set +x; } 2>/dev/null

  for ((i = 0; i < $PEER_PER_ORGANIZATION_NUMBER; i++)) do
    infoln "Generating the peer$i msp"
    set -x
    fabric-ca-client enroll -u https://peer$i:peer${i}pw@localhost:${caPort} --caname ca-org$1 -M ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/peers/peer$i.org$1.example.com/msp --csr.hosts peer$i.org$1.example.com --tls.certfiles ${ORGANIZATIONS_PATH}/fabric-ca/org$1/tls-cert.pem
    { set +x; } 2>/dev/null

    cp ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/msp/config.yaml ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/peers/peer$i.org$1.example.com/msp/config.yaml

    infoln "Generating the peer$i-tls certificates"
    set -x
    fabric-ca-client enroll -u https://peer$i:peer${i}pw@localhost:${caPort} --caname ca-org$1 -M ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/peers/peer$i.org$1.example.com/tls --enrollment.profile tls --csr.hosts peer$i.org$1.example.com --csr.hosts localhost --tls.certfiles ${ORGANIZATIONS_PATH}/fabric-ca/org$1/tls-cert.pem
    { set +x; } 2>/dev/null

    cp ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/peers/peer${i}.org$1.example.com/tls/tlscacerts/* ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/peers/peer${i}.org$1.example.com/tls/ca.crt
    cp ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/peers/peer${i}.org$1.example.com/tls/signcerts/* ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/peers/peer${i}.org$1.example.com/tls/server.crt
    cp ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/peers/peer${i}.org$1.example.com/tls/keystore/* ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/peers/peer${i}.org$1.example.com/tls/server.key

    mkdir -p ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/msp/tlscacerts
    cp ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/peers/peer${i}.org$1.example.com/tls/tlscacerts/* ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/msp/tlscacerts/ca.crt

    mkdir -p ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/tlsca
    cp ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/peers/peer${i}.org$1.example.com/tls/tlscacerts/* ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/tlsca/tlsca.org$1.example.com-cert.pem

    mkdir -p ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/ca
    cp ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/peers/peer${i}.org$1.example.com/msp/cacerts/* ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/ca/ca.org$1.example.com-cert.pem

  done

  infoln "Generating the user msp"
  set -x
  fabric-ca-client enroll -u https://user1:user1pw@localhost:${caPort} --caname ca-org$1 -M ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/users/User1@org$1.example.com/msp --tls.certfiles ${ORGANIZATIONS_PATH}/fabric-ca/org$1/tls-cert.pem
  { set +x; } 2>/dev/null

  cp ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/msp/config.yaml ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/users/User1@org$1.example.com/msp/config.yaml

  infoln "Generating the org admin msp"
  set -x
  fabric-ca-client enroll -u https://org${1}admin:org${1}adminpw@localhost:${caPort} --caname ca-org$1 -M ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/users/Admin@org$1.example.com/msp --tls.certfiles ${ORGANIZATIONS_PATH}/fabric-ca/org$1/tls-cert.pem
  { set +x; } 2>/dev/null

  cp ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/msp/config.yaml ${ORGANIZATIONS_PATH}/peerOrganizations/org$1.example.com/users/Admin@org$1.example.com/msp/config.yaml
}

function createOrderer() {
  infoln "Enrolling the CA admin"
  mkdir -p ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com

  export FABRIC_CA_CLIENT_HOME=${ORGANIZATIONS_PATH}/ordererOrganizations/example.com

  set -x
  fabric-ca-client enroll -u https://admin:adminpw@localhost:5000 --caname ca-orderer --tls.certfiles ${ORGANIZATIONS_PATH}/fabric-ca/ordererOrg/tls-cert.pem
  { set +x; } 2>/dev/null
  
  echo 'NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/localhost-5000-ca-orderer.pem
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/localhost-5000-ca-orderer.pem
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/localhost-5000-ca-orderer.pem
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/localhost-5000-ca-orderer.pem
    OrganizationalUnitIdentifier: orderer' >${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/msp/config.yaml

  infoln "Registering orderer"
  set -x
  fabric-ca-client register --caname ca-orderer --id.name orderer --id.secret ordererpw --id.type orderer --tls.certfiles ${ORGANIZATIONS_PATH}/fabric-ca/ordererOrg/tls-cert.pem
  { set +x; } 2>/dev/null

  infoln "Registering the orderer admin"
  set -x
  fabric-ca-client register --caname ca-orderer --id.name ordererAdmin --id.secret ordererAdminpw --id.type admin --tls.certfiles ${ORGANIZATIONS_PATH}/fabric-ca/ordererOrg/tls-cert.pem
  { set +x; } 2>/dev/null

  infoln "Generating the orderer msp"
  set -x
  fabric-ca-client enroll -u https://orderer:ordererpw@localhost:5000 --caname ca-orderer -M ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/orderers/orderer.example.com/msp --csr.hosts orderer.example.com --csr.hosts localhost --tls.certfiles ${ORGANIZATIONS_PATH}/fabric-ca/ordererOrg/tls-cert.pem
  { set +x; } 2>/dev/null

  cp ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/msp/config.yaml ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/orderers/orderer.example.com/msp/config.yaml

  infoln "Generating the orderer-tls certificates"
  set -x
  fabric-ca-client enroll -u https://orderer:ordererpw@localhost:5000 --caname ca-orderer -M ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/orderers/orderer.example.com/tls --enrollment.profile tls --csr.hosts orderer.example.com --csr.hosts localhost --tls.certfiles ${ORGANIZATIONS_PATH}/fabric-ca/ordererOrg/tls-cert.pem
  { set +x; } 2>/dev/null

  cp ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/orderers/orderer.example.com/tls/tlscacerts/* ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt
  cp ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/orderers/orderer.example.com/tls/signcerts/* ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt
  cp ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/orderers/orderer.example.com/tls/keystore/* ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.key

  mkdir -p ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts
  cp ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/orderers/orderer.example.com/tls/tlscacerts/* ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

  mkdir -p ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/msp/tlscacerts
  cp ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/orderers/orderer.example.com/tls/tlscacerts/* ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/msp/tlscacerts/tlsca.example.com-cert.pem

  infoln "Generating the admin msp"
  set -x
  fabric-ca-client enroll -u https://ordererAdmin:ordererAdminpw@localhost:5000 --caname ca-orderer -M ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/users/Admin@example.com/msp --tls.certfiles ${ORGANIZATIONS_PATH}/fabric-ca/ordererOrg/tls-cert.pem
  { set +x; } 2>/dev/null

  cp ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/msp/config.yaml ${ORGANIZATIONS_PATH}/ordererOrganizations/example.com/users/Admin@example.com/msp/config.yaml
}
