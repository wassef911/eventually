#!/bin/bash

KUSTOMIZE_DIR="deployments/overlays/prod"
echo "Teardown: $KUSTOMIZE_DIR"

cd $KUSTOMIZE_DIR
kustomize build --enable-helm . | kubectl delete -f -

cd -

KUSTOMIZE_DIR="deployments/components"
echo "Teardown: $KUSTOMIZE_DIR"

cd $KUSTOMIZE_DIR
kustomize build --enable-helm . | kubectl delete -f -
