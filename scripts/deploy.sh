#!/bin/bash

KUSTOMIZE_DIR="deployments/overlays/prod"
echo "Deploying from: $KUSTOMIZE_DIR"

cd $KUSTOMIZE_DIR
kustomize build --enable-helm . | kubectl apply -f -

cd -

KUSTOMIZE_DIR="deployments/components"
echo "Deploying from: $KUSTOMIZE_DIR"

cd $KUSTOMIZE_DIR
kustomize build --enable-helm . | kubectl apply -f -
