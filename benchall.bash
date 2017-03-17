#!/bin/bash 

set -e

export GOPATH=$(pwd)

buildjuju() {
	$@ version
	/usr/bin/time -f '%e' $@ build -o /tmp/jujud.$@ github.com/juju/juju/cmd/jujud \
		&& ls -s /tmp/jujud.$@ \
		&& rm /tmp/jujud.$@
}

for i in $(seq 1 5); do
	buildjuju go1.4
	buildjuju go1.7
	buildjuju go1.8
	buildjuju go.tip
done
