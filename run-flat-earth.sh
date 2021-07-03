#!/bin/bash

if [ -z $1 ]; then
    echo "No arguments supplied"
    exit 1
fi

export TF_LOG=DEBUG

# kanishk98: make this configurable by user, maybe this should be sent over by the frontend? anyway, for now, we'll store them in the home directory.
export TF_LOG_PATH=~/.flat-earth/logs/log.txt
if [ ! -f $TF_LOG_PATH ]; then
    echo "Creating new log file..."
    mkdir -p "${TF_LOG_PATH%/*}"
    touch $TF_LOG_PATH
fi

# kanishk98: obviously, we need to make this point to the location of the last-created FE executable.
pushd $1
~/go/bin/terraform $1
