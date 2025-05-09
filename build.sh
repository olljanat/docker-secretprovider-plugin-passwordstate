#!/bin/bash

USAGE="Usage: ./build.sh <Docker Hub Organization> <version> [Passwordstate API URL] [Passwordstate API key] [Passwordstate list ID]"

if [ "$1" == "--help" ] || [ "$#" -lt "2" ] || [ "$#" -gt "5" ]; then
	echo $USAGE
	exit 0
fi

ORG=$1
VERSION=$2
PASSWORDSTATE_BASE_URL=$3
PASSWORDSTATE_API_KEY=$4
PASSWORDSTATE_LIST_ID=$5

rm -rf rootfs
docker plugin disable $ORG/docker-secretprovider-plugin-passwordstate:v$VERSION
docker plugin rm $ORG/docker-secretprovider-plugin-passwordstate:v$VERSION

docker plugin disable pwdstate:latest
docker plugin rm pwdstate:latest

mkdir -p rootfs
mkdir -p rootfs/etc/ssl/certs/
cp /etc/ssl/certs/ca-certificates.crt rootfs/etc/ssl/certs/
CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w -extldflags "-static"'
cp docker-secretprovider-plugin-passwordstate rootfs/

docker plugin create $ORG/docker-secretprovider-plugin-passwordstate:v$VERSION .
docker plugin push $ORG/docker-secretprovider-plugin-passwordstate:v$VERSION

docker plugin rm $ORG/docker-secretprovider-plugin-passwordstate:v$VERSION

if [ "$#" == "5" ]; then
  docker plugin install \
    --alias pwdstate \
    --grant-all-permissions \
    $ORG/docker-secretprovider-plugin-passwordstate:v$VERSION \
    PASSWORDSTATE_BASE_URL=$PASSWORDSTATE_BASE_URL \
    PASSWORDSTATE_API_KEY=$PASSWORDSTATE_API_KEY \
    PASSWORDSTATE_LIST_ID=$PASSWORDSTATE_LIST_ID
fi
