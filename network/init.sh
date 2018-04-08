DELAY=3

node init.js

export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/coderschool.vn/msp/tlscacerts/tlsca.coderschool.vn-cert.pem

echo "=== Setting up salmon chaincode"

export SALMON=salmon
export SALMON_VERSION=v1

# intall
docker exec cli.coderschool.vn peer chaincode install --name $SALMON --version $SALMON_VERSION --lang golang --path github.com/anh/salmon-supply-chain/chaincode/salmon
sleep $DELAY

# instantiate
docker exec cli.coderschool.vn peer chaincode instantiate -o orderer.coderschool.vn:7050 --tls true --cafile $ORDERER_CA -C transfers -n $SALMON -v $SALMON_VERSION -c '{"Args":[]}'
sleep $DELAY

echo "=== Setting up agreement chaincode on fredrick-alice channel"

export AGREEMENT=agreement
export AGREEMENT_VERSION=v1

# intall
docker exec cli.coderschool.vn peer chaincode install --name $AGREEMENT --version $AGREEMENT_VERSION --lang golang --path github.com/anh/salmon-supply-chain/chaincode/agreement
sleep $DELAY

# instantiate
docker exec cli.coderschool.vn peer chaincode instantiate -o orderer.coderschool.vn:7050 --tls true --cafile $ORDERER_CA -C fredrick-alice -n $AGREEMENT -v $AGREEMENT_VERSION -c '{"Args":[]}'
sleep $DELAY

# init data
docker exec cli.coderschool.vn peer chaincode invoke --tls true --cafile $ORDERER_CA -o orderer.coderschool.vn:7050  -C fredrick-alice -n $AGREEMENT -v $AGREEMENT_VERSION -c '{"Args":["recordAgreement", "1", "50"]}'

echo "=== Setting up agreement chaincode on fredrick-bob channel"

# instantiate
docker exec cli.coderschool.vn peer chaincode instantiate -o orderer.coderschool.vn:7050 --tls true --cafile $ORDERER_CA -C fredrick-bob -n $AGREEMENT -v $AGREEMENT_VERSION -c '{"Args":[]}'
sleep $DELAY

# init data
docker exec cli.coderschool.vn peer chaincode invoke --tls true --cafile $ORDERER_CA -o orderer.coderschool.vn:7050  -C fredrick-bob -n $AGREEMENT -v $AGREEMENT_VERSION -c '{"Args":["recordAgreement", "1", "100"]}'
