#!/bin/bash

set -euo pipefail

# k directory (default to current if not provided)
KUSTOMIZE_DIR="${1:-.}"

echo "Deploying using kustomize from: $KUSTOMIZE_DIR"

kubectl apply -k "$KUSTOMIZE_DIR"
