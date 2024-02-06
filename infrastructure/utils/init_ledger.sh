peer chaincode invoke \
  -o localhost:7000 --ordererTLSHostnameOverride orderer.example.com \
  --tls --cafile "${PWD}/../organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" \
  -C channel1 \
  -n bankchaincode1  \
  --peerAddresses localhost:7050 --tlsRootCertFiles "${PWD}/../organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" \
  --peerAddresses localhost:8050 --tlsRootCertFiles "${PWD}/../organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" \
  --peerAddresses localhost:9050 --tlsRootCertFiles "${PWD}/../organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt" \
  --peerAddresses localhost:10050 --tlsRootCertFiles "${PWD}/../organizations/peerOrganizations/org4.example.com/peers/peer0.org4.example.com/tls/ca.crt" \
  -c '{"function":"InitLedger","Args":[]}'
