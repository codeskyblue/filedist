#!/bin/sh
#
#

cd `dirname $0`

cd ../
go build
cd $OLDPWD

path=${1:-"/home/work/c"}
../filedist -s $HOSTNAME \
        --dfile=dest.txt -p ${path} | tee result.txt
