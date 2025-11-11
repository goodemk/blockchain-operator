#!/bin/bash

set -e

if [[ $(kind get clusters 2>&1 ) != "No kind clusters found." ]]; then
    echo "Found existing cluster. Cleaning up..."
    if ! kind delete cluster --name besu; then
        echo "Failed to delete kind cluster."
        exit 1
    fi
    echo "Kind cluster deleted successfully."
else
    echo "Nothing to clean up."
fi

echo "Cleanup complete!"
