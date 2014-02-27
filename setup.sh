#!/bin/sh

export GOPATH=`pwd`
export GOROOT=$HOME/usr/go
export CGO_CFLAGS="-I $HOME/usr/nvml/include -L $HOME/usr/nvml/lib64 -L/usr/lib/nvidia-current $CFLAGS"
export CGO_LDFLAGS="-L $HOME/usr/nvml/lib64 -L /usr/lib/nvidia-current -lnvidia-ml $LDFLAGS "
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

PGI=/opt/pgi
PATH=/opt/pgi/linux86-64/14.2/bin:$PATH
export PGI
export PATH


