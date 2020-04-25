#!/bin/bash
## Build Operator Image with operator-sdk
## See details: https://github.com/operator-framework/operator-sdk

usage() {
    echo "$0: need at least one argument."
    echo "$0 <Image Tag>"
}

if ! type operator-sdk >/dev/null 2>&1; then
  echo "ERROR: No Found ooprator-sdk on the host."
  echo "Please install operator-sdk first."
  echo "See details: https://github.com/operator-framework/operator-sdk"
  exit 1
fi

IMAGETAG=${1:-latest}

operator-sdk build \
      quay.io/hkaneko/gitbucket-operator:${IMAGETAG} \
      --go-build-args "-o ./build/_output/bin/gitbucket-operator"