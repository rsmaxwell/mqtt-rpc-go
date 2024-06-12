#!/bin/sh

set -x

BUILD_DIR="$(pwd)/build"
PACKAGE_DIR="$(pwd)/package"
DIST_DIR="$(pwd)/dist"
INFO_DIR=${BUILD_DIR}/info

. ${INFO_DIR}/maven.sh

ZIPFILE=${NAME}.zip

rm -rf ${PACKAGE_DIR} ${DIST_DIR}
mkdir -p ${PACKAGE_DIR} ${DIST_DIR}

cd ${PACKAGE_DIR}
cp -r ${BUILD_DIR}/* .

zip -r ${DIST_DIR}/${ZIPFILE} *