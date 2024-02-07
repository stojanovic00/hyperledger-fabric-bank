#!/bin/bash

function one_line_pem {
    echo "`awk 'NF {sub(/\\n/, ""); printf "%s\\\\\\\n",$0;}' $1`"
}

function json_ccp {
    local PP=$(one_line_pem $4)
    local CP=$(one_line_pem $5)
    sed -e "s/\${ORG}/$1/" \
        -e "s/\${P0PORT}/$2/" \
        -e "s/\${CAPORT}/$3/" \
        -e "s#\${PEERPEM}#$PP#" \
        -e "s#\${CAPEM}#$CP#" \
        ${ORGANIZATIONS_PATH}/ccp-template.json
}

function yaml_ccp {
    local PP=$(one_line_pem $4)
    local CP=$(one_line_pem $5)
    sed -e "s/\${ORG}/$1/" \
        -e "s/\${P0PORT}/$2/" \
        -e "s/\${CAPORT}/$3/" \
        -e "s#\${PEERPEM}#$PP#" \
        -e "s#\${CAPEM}#$CP#" \
        ${ORGANIZATIONS_PATH}/ccp-template.yaml | sed -e $'s/\\\\n/\\\n          /g'
}

for ((i=1 ; i <= $1 ; ++i)) do
    P0PORT="$((6 + $i))050"
    CAPORT="$((6 + $i))150"
    PEERPEM=${ORGANIZATIONS_PATH}/peerOrganizations/org$i.example.com/tlsca/tlsca.org$i.example.com-cert.pem
    CAPEM=${ORGANIZATIONS_PATH}/peerOrganizations/org$i.example.com/ca/ca.org$i.example.com-cert.pem

    echo "$(json_ccp $i $P0PORT $CAPORT $PEERPEM $CAPEM)" > ${ORGANIZATIONS_PATH}/peerOrganizations/org$i.example.com/connection-org$i.json
    echo "$(yaml_ccp $i $P0PORT $CAPORT $PEERPEM $CAPEM)" > ${ORGANIZATIONS_PATH}/peerOrganizations/org$i.example.com/connection-org$i.yaml
done