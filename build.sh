#!/bin/bash

set -e

DIRNAME="$(dirname ${BASH_SOURCE[0]})"
ROOT="$(cd $DIRNAME; pwd)"

main() {
  build_app
  generate
  build_bin
}

build_app() {
  pushd "${ROOT}/app"
  yarn build
  popd
}

generate() {
   pushd "${ROOT}"
   GOPATH= go generate ./...
   popd
}

build_bin() {
  pushd "${ROOT}"
  GOPATH= go build ./
  popd
}

main
