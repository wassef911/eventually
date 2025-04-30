#!/bin/bash

set -euo pipefail

# k directory (default to current if not provided)
KUSTOMIZE_DIR="${1:-.}"

echo "Tearing down resources from: $KUSTOMIZE_DIR"

cd $KUSTOMIZE_DIR
kustomize build --enable-helm . | kubectl delete -f -
