#!/bin/bash

usage() {
    echo "Usage $0 (build-linux|sync)"
}

if [[ "$1" == "" ]]; then
    usage
    exit 1
fi


while (( "$#" )); do
    case "$1" in
        build-linux)
            GIT_VERSION=$(git describe --always --abbrev=8  --dirty --broken)
            env GOOS=linux GOARCH=amd64 go build -ldflags "-X main.buildVersion=${GIT_VERSION}" cmd/sesam.go
            ;;
        test-sync)
            rsync -n -avzi --delete sesam webUI root@spacegate:/home/sesam/sesam-app/
            ;;
        sync)
            rsync -avzi --delete sesam webUI root@spacegate:/home/sesam/sesam-app/
            ;;
        *)
            usage
            exit 1
    esac
    shift
done