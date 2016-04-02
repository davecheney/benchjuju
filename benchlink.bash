#!/bin/bash

# Benchmark linking an executable in a format benchstat can parse

# Arguments are: how many times to run the link, the package path
# containing the executable and a benchmark name.  For example:
#
# (master)mwhudson@aeglos:benchjuju$ ./benchlink.bash 10 cmd/go CmdGo
# BenchmarkLinkCmdGo 1 0.41 secs 116956 MaxRSS
# BenchmarkLinkCmdGo 1 0.41 secs 121888 MaxRSS
# BenchmarkLinkCmdGo 1 0.41 secs 116884 MaxRSS
# BenchmarkLinkCmdGo 1 0.43 secs 117980 MaxRSS
# BenchmarkLinkCmdGo 1 0.42 secs 116292 MaxRSS
# BenchmarkLinkCmdGo 1 0.41 secs 117340 MaxRSS
# BenchmarkLinkCmdGo 1 0.42 secs 116592 MaxRSS
# BenchmarkLinkCmdGo 1 0.42 secs 115624 MaxRSS
# BenchmarkLinkCmdGo 1 0.43 secs 115744 MaxRSS
# BenchmarkLinkCmdGo 1 0.45 secs 115888 MaxRSS

buildlog=$(go build -o /tmp/blah -x -work $2 2>&1)
eval $(echo "$buildlog" | grep WORK=)
linkcmd=$(echo "$buildlog" | grep -F "$(go tool -n link)")
for i in `seq $1`; do
    eval /usr/bin/time -f '"BenchmarkLink$3 1 %e secs %M MaxRSS"' $linkcmd
done

