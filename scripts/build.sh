#!/bin/bash

set -e

APP_NAME="racecourse"
IMAGE_VERSION="0.0.1"
HELM_CHART_DIR="./helm"

echo "Checking Colima..."
if ! command -v colima &> /dev/null; then
    echo "Colima is not installed. Please install it first."
    exit 1
fi

if ! colima status --profile besu &> /dev/null; then
    echo "Starting Colima..."
    if ! colima start --profile besu \
      --runtime docker \
      --cpu 4 --memory 6 \
      --disk 40 \
      --vm-type vz \
      --arch aarch64 \
      --vz-rosetta; then
        echo "Failed to start Colima."
        exit 1
    fi
    echo "Colima started successfully."
else
    echo "Colima is already running."
fi

echo "Initializing cluster..."
if ! command -v kind &> /dev/null; then
    echo "kind is not installed. Please install kind first."
    exit 1
fi

if [[ $(kind get clusters 2>&1 ) == "No kind clusters found." ]]; then
    echo "No cluster found. Creating a new cluster..."
    if ! kind create cluster --name besu --config scripts/kind-config.yaml; then
        echo "Failed to create kind cluster."
        exit 1
    fi
    echo "Kind cluster created successfully."
else
    echo "Kind cluster is already running."
fi

echo -e "\nBuilding the Racecourse image..."
if ! docker build -t $APP_NAME:$IMAGE_VERSION .; then
    echo "Failed to build the Docker image."
    exit 1
fi

echo -e "\nUploading image to the node..."
if ! kind load docker-image --name besu $APP_NAME:$IMAGE_VERSION; then
    echo "Failed to upload the Docker image. Ensure the cluster exists."
    exit 1
fi

echo -e "\nGenerating keys..."
if ! ./scripts/generate-keys.sh; then
    echo "Failed to generate keys."
    exit 1
fi

echo -e "\nInstalling Besu..."
if ! helm upgrade --install besu ${HELM_CHART_DIR}/besu \
    --namespace sidechain \
    --create-namespace \
    --wait \
    --timeout 5m; then \
    echo "Failed to install Besu."
    exit 1
fi

echo -e "\nInstalling Nginx..."
if ! helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx; then
    echo "Failed to add Nginx ingress controller repository."
    exit 1
fi

if ! helm repo update; then
    echo "Failed to update Nginx ingress controller repository."
    exit 1
fi
if ! helm upgrade -n nginx --create-namespace \
    --install ingress-nginx ingress-nginx/ingress-nginx \
    -f ${HELM_CHART_DIR}/nginx/values.yaml > /dev/null; then
    echo "Failed to install Nginx ingress controller."
    exit 1
fi
    echo "Nginx installed successfully."

echo -e "\nInstalling Firefly Signer..."
if ! helm upgrade --install firefly-signer ${HELM_CHART_DIR}/firefly-signer \
    --namespace sidechain \
    --create-namespace \
    --wait \
    --timeout 5m; then
    echo "Failed to install Firefly Signer."
    exit 1
fi

echo -e "\nDeploying racecourse operator..."

cd operator/

echo " - Regenerating manifests..."
if ! make manifests generate; then
    echo "Failed to regenerate manifests."
    exit 1
fi

echo " - Building operator image..."
if ! make docker-build; then
    echo "Failed to build operator image."
    exit 1
fi

echo " - Loading operator image to cluster..."
if ! kind load docker-image $APP_NAME-operator:latest --name besu; then
    echo "Failed to load operator image."
    exit 1
fi

echo " - Deploying operator..."
if ! make deploy; then
    echo "Failed to deploy operator."
    exit 1
fi

read -e -p ">> Deploy sample Racecourse? (y/n): " DEPLOY
if [[ "$DEPLOY" == "y" ]]; then
    echo "Deploying an instance of Racecourse..."
    if ! kubectl apply -f config/samples/minimal_racecourse.yaml; then
        echo "Failed to deploy sample racecourse."
        exit 1
    fi
    echo "Sample racecourse deployed successfully."
fi