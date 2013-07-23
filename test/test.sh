#!/bin/sh
#
#

go build
./filedist -s $HOSTNAME --df=dest.txt -p ${1:-"/home/work/c"} | tee result.txt
