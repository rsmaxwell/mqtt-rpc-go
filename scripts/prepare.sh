#!/bin/sh

set -x 

BRANCH=${1}
if [ -z "${BRANCH}" ]; then
    echo "Error: $0[${LINENO}]"
    echo "Missing BRANCH argument"
    exit 1
fi

if [ -z "$2" ]; then
    echo "Missing PLATFORM argument"
    exit 1
fi
PLATFORM="$2"




if [ -z "${BUILD_ID}" ]; then
    VERSION="0.0-SNAPSHOT"
else
    VERSION="0.0.$((${BUILD_ID}))"
fi



TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
    
find . -name "version.go" | while read versionfile; do

    echo "Replacing tags in ${versionfile}"

    sed -i "s@<VERSION>@${VERSION}@g"            ${versionfile}
    sed -i "s@<BUILD_ID>@${BUILD_ID}@g"          ${versionfile}
    sed -i "s@<BUILD_DATE>@${TIMESTAMP}@g"       ${versionfile}
    sed -i "s@<GIT_COMMIT>@${GIT_COMMIT}@g"      ${versionfile}
    sed -i "s@<GIT_BRANCH>@${GIT_BRANCH}@g"      ${versionfile}
    sed -i "s@<GIT_URL>@${GIT_URL}@g"            ${versionfile}
done


BUILD_DIR=$(pwd)/build
INFO_DIR=${BUILD_DIR}/info



rm -rf ${BUILD_DIR}
mkdir -p ${INFO_DIR}
cd ${INFO_DIR}



cat << EOF > info.json
{
	"VERSION": "${VERSION}",
	"BUILD_ID": ${BUILD_ID},
	"TIMESTAMP": "${TIMESTAMP}",
	"pipeline": {
		"GIT_COMMIT": "${GIT_COMMIT}",
		"GIT_BRANCH": "${GIT_BRANCH}",
		"GIT_URL": "${GIT_URL}"
	},
	"project": {
		"GIT_COMMIT": "$(git rev-parse HEAD)",
		"GIT_BRANCH": "${BRANCH}",
		"GIT_URL": "$(git config --local remote.origin.url)"
	}
}
EOF


NAME=players-tt-api

GROUPID=com.rsmaxwell.players
ARTIFACTID=${NAME}-${PLATFORM}
PACKAGING=zip

REPOSITORY=releases
REPOSITORYID=releases
URL=https://pluto.rsmaxwell.co.uk/archiva/repository/${REPOSITORY}


cat << EOF > maven.sh
NAME=${NAME}
GROUPID=${GROUPID}
ARTIFACTID=${ARTIFACTID}
PACKAGING=${PACKAGING}
VERSION=${VERSION}
REPOSITORY=${REPOSITORY}
REPOSITORYID=${REPOSITORYID}
URL=${URL}
EOF
