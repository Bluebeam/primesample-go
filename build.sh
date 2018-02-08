#!/usr/bin/env bash

#Fail immediately if any of the commands below Fail
set -x

# Display current version of Go
go version
echo GOROOT=$GOROOT
echo GOPATH=$GOPATH

# Install new Go
GOGZ=go1.9.2.linux-amd64.tar.gz
curl -O "https://storage.googleapis.com/golang/$GOGZ"
tar -zxf $GOGZ
rm $GOGZ

export GOROOT=/usr/local/go
export GOPATH=/usr/local/repos
rm -rf $GOROOT

mv go /usr/local
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# Display new version
go version
echo GOROOT=$GOROOT
echo GOPATH=$GOPATH

go get github.com/boltdb/bolt/...
go get -u github.com/jteeuwen/go-bindata/...
go get -u golang.org/x/oauth2

DEST=$GOPATH/src/bluebeam
rm -rf $DEST
mkdir -p $DEST
cp -r bluebeam/* $DEST

cd $DEST/gosessionroundtripper

go-bindata assets/
go build -o application

chmod +x application

mv application /usr/bin

