#!/bin/bash 

set -e

GO=go

export GOPATH=$(pwd)

${GO} version

for i in $(seq 1 10); do
	time ${GO} build github.com/juju/juju/cmd/jujud
done
