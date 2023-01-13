#!/bin/bash


set -e -u

go version # so we see the version tested in CI

if ! [ $(type -P "ginkgo") ]; then
 go install -mod=mod github.com/onsi/ginkgo/ginkgo@v1
fi

SCRIPT_PATH="$(cd "$(dirname "${0}")" && pwd)"
cd "${SCRIPT_PATH}/.."

pushd src/code.cloudfoundry.org/healthchecker/cmd/healthchecker/
 ginkgo -r
 result=$?
popd

return "${result}"
