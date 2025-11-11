#!/bin/bash

VALIDATORS=4

mkdir -p qbft-network
rm -rf qbft-network/*

cd qbft-network

ADDRESSES=()
PUBKEYS=()

for i in $(seq 0 $((VALIDATORS - 1))); do

    echo "Generating keys and addresses for node $i..."

    mkdir -p node-$i

    docker run --rm -v $(pwd)/node-$i:/data hyperledger/besu:latest --data-path=/data public-key export-address --to=/data/address > /dev/null
    docker run --rm -v $(pwd)/node-$i:/data hyperledger/besu:latest --data-path=/data public-key export --to=/data/pubkey > /dev/null

    ADDRESS=$(cat node-$i/address | sed 's/0x//')
    PUBKEY=$(cat node-$i/pubkey | sed 's/0x//')
    ADDRESSES+=("$ADDRESS")
    PUBKEYS+=("$PUBKEY")

done

echo "[" > toEncode.json
for i in $(seq 0 $((VALIDATORS - 1))); do
    if [ $i -eq $((VALIDATORS - 1)) ]; then
        echo "  \"${ADDRESSES[$i]}\"" >> toEncode.json
    else
        echo "  \"${ADDRESSES[$i]}\"," >> toEncode.json
    fi
done
echo "]" >> toEncode.json

EXTRA_DATA=$(docker run --rm -v $(pwd):/data hyperledger/besu:latest rlp encode --from=/data/toEncode.json --type=QBFT_EXTRA_DATA)

cat ../helm/besu/genesis-base.json | jq --arg extraData "$EXTRA_DATA" '.extraData = $extraData' > ../helm/besu/genesis.json

for i in $(seq 0 $((VALIDATORS - 1))); do
    ENODE="${PUBKEYS[$i]}:besu-$i"
    echo "$ENODE" >> enodes.txt
done

cp enodes.txt ../helm/besu/enodes.txt

echo -e "\nCreating Kubernetes secrets for validator keys..."

kubectl create namespace sidechain

NAMESPACE=${NAMESPACE:-sidechain}

for i in $(seq 0 $((VALIDATORS - 1))); do
    SECRET_NAME="besu-$i-key"
    KEY_FILE="node-$i/key"

    echo "Creating secret $SECRET_NAME in namespace $NAMESPACE..."

    kubectl delete secret "$SECRET_NAME" -n "$NAMESPACE" 2>/dev/null || true

    kubectl create secret generic "$SECRET_NAME" \
        --from-file=key="$KEY_FILE" \
        -n "$NAMESPACE"
done

echo -e "\nDone."