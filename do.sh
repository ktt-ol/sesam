#!/bin/bash

usage() {
    echo "Usage $0 (build-linux|sync)"
}

if [ "$1" == "" ]; then
    usage
    exit 1
fi


while (( "$#" )); do
    case "$1" in
        build-linux)
            env GOOS=linux GOARCH=amd64 go build cmd/sesam.go
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