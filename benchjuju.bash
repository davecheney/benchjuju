#!/bin/bash 

set -e

go version

export GOPATH=$(pwd)

for i in $(seq 1 10); do
	time go build github.com/juju/juju/cmd/jujud
done
