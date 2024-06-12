#!/bin/sh

set -x

BUILD_DIR=$(pwd)/build
DIST_DIR=$(pwd)/dist
INFO_DIR=${BUILD_DIR}/info

. ${INFO_DIR}/maven.sh

ZIPFILE=${DIST_DIR}/${NAME}.zip

mvn --batch-mode deploy:deploy-file \
	-DgroupId=${GROUPID} \
	-DartifactId=${ARTIFACTID} \
	-Dversion=${VERSION} \
	-Dpackaging=${PACKAGING} \
	-Dfile=${ZIPFILE} \
	-DrepositoryId=${REPOSITORYID} \
	-Durl=${URL}