#!/bin/sh

# Benchmark a linker change

# Arguments are: "old" go branch (often master), "new" go branch,
# number of iterations to pass to benchlink.bash. For example:

# rsc.io/benchstat should be on $PATH, although the old.txt and
# new.txt files are left in this directory if you forget.

# root@scw-d4325d:~# benchlinkbranch.sh master slice-lengths 10
# Building go master
# Already on 'master'
# Your branch is up-to-date with 'origin/master'.
# installing juju with master
# Benchmarking go old
# Benchmarking juju old
# Building go slice-lengths
# installing juju with slice-lengths
# Benchmarking go new
# Benchmarking juju new
# name       old secs    new secs    delta
# LinkCmdGo   1.17 ± 1%   1.19 ± 4%  +1.71%  (p=0.018 n=10+10)
# LinkJuju    10.7 ± 1%   10.8 ± 2%    ~      (p=0.645 n=9+10)
# 
# name       old MaxRSS  new MaxRSS  delta
# LinkCmdGo   122k ± 0%   121k ± 3%    ~      (p=0.968 n=9+10)
# LinkJuju    829k ± 5%   836k ± 4%    ~     (p=0.579 n=10+10)


set -ex

export GOPATH="$(pwd)"

# Set REBUILD_ALL to 1 if everything needs to be rebuilt between the
# runs (e.g. an object file format change) or anything else if "go
# install cmd" is enough of an update (and save yourself a bunch of
# time).
REBUILD_ALL=1

cd "$(go env GOROOT)/src"

echo "Building go $1"

git checkout $1

if [ $REBUILD_ALL = 1 ]; then
    ./make.bash > /dev/null 2>&1
else
    go install -v cmd
fi

echo "installing juju with $1"

go install github.com/juju/juju/cmd/juju

echo "Benchmarking go old"

"${GOPATH}/benchlink.bash" $3 cmd/go CmdGo 2>&1 > "${GOPATH}/old.txt" 2>&1

echo "Benchmarking juju old"

"${GOPATH}/benchlink.bash" $3 github.com/juju/juju/cmd/juju Juju >> "${GOPATH}/old.txt" 2>&1


echo "Building go $2"

if [ $REBUILD_ALL = 1 ]; then
    ./make.bash > /dev/null 2>&1
else
    go install -v cmd
fi

echo "installing juju with $2"

go install github.com/juju/juju/cmd/juju

echo "Benchmarking go new"

"${GOPATH}/benchlink.bash" $3 cmd/go CmdGo 2>&1 > "${GOPATH}/new.txt" 2>&1

echo "Benchmarking juju new"

"${GOPATH}/benchlink.bash" $3 github.com/juju/juju/cmd/juju Juju >> "${GOPATH}/new.txt" 2>&1

benchstat "${GOPATH}/old.txt" "${GOPATH}/new.txt"
