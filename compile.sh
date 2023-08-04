#!/bin/sh
set -e # If a command fails, stop the script

# Set the GOPATH to the current directory
export GOPATH=$(pwd)

# Check if Go is installed
if ! command -v go &> /dev/null
then
    echo "Go could not be found. Please install it."
    exit
fi

cd app/cmd/charger
GOOS=linux GOARCH=amd64 go build -o main

cd ../listener
GOOS=linux GOARCH=amd64 go build -o main

echo "Build successful"
