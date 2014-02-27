#!/bin/sh

export GOPATH=`pwd`
export GOROOT=/scr/dakkak/usr/go
export CGO_CFLAGS="-I /scr/dakkak/usr/nvml/include -L /scr/dakkak/usr/nvml/lib64 $CFLAGS"
export CGO_LDFLAGS="-L /scr/dakkak/usr/nvml/lib64 -lnvidia-ml $LDFLAGS "
export LD_LIBRARY_PATH=/scr/dakkak/usr/cuda/lib64:$LD_LIBRARY_PATH
export PATH=$GOPATH/bin:$GOROOT/bin:/scr/dakkak/usr/cuda/bin:$PATH
export TMPDIR=/scr/dakkak/tmp

rm -fr $TMPDIR/*
mkdir -p $TMPDIR

PGI=/opt/pgi
PATH=/opt/pgi/linux86-64/14.2/bin:$PATH
export PGI
export PATH

