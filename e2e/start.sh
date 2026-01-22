#!/bin/bash
set -e

# Ensure we are in the project root
cd "$(dirname "$0")/.."

# Add local bin to PATH
export PATH=$(pwd)/bin:$PATH

KIND_CONFIG="e2e/kind-config.yaml"
KUBECONFIG_FILE="e2e/kubeconfig.yaml"
DB_FILE="e2e/e2e.db"

# Create kind cluster if not exists
if ! sudo $(pwd)/bin/kind get clusters | grep -q "^kind$"; then
    echo "Creating Kind cluster..."
    # Use kind-config.yaml if it exists, otherwise default
    if [ -f "$KIND_CONFIG" ]; then
        sudo $(pwd)/bin/kind create cluster --config "$KIND_CONFIG" --wait 1m
    else
        sudo $(pwd)/bin/kind create cluster --wait 1m
    fi
else
    echo "Kind cluster 'kind' already exists."
fi

# Generate kubeconfig
echo "Exporting kubeconfig..."
sudo $(pwd)/bin/kind get kubeconfig --name kind > "$KUBECONFIG_FILE"
sudo chown $(id -u):$(id -g) "$KUBECONFIG_FILE"
export KUBECONFIG=$(pwd)/"$KUBECONFIG_FILE"

# Set other env vars
export DB_DSN="$DB_FILE"
export CLOUD_SENTINEL_K8S_USERNAME=admin
export CLOUD_SENTINEL_K8S_PASSWORD=admin

echo "Starting server..."
go run main.go
