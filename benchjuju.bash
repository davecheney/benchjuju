#!/bin/bash 

set -e

GO=go

export GOPATH=$(pwd)

${GO} version

for i in $(seq 1 10); do
	/usr/bin/time -f '%e' ${GO} build -o /tmp/jujud github.com/juju/juju/cmd/jujud \
		&& ls -al /tmp/jujud \
		&& rm /tmp/jujud
done
