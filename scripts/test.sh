#!/bin/sh

set -x

echo "testing: players-tt-api"

BUILD_DIR=./build
cd ${BUILD_DIR}

pwd
ls -al 

./players-tt-api --version