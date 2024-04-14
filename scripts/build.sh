#!/bin/bash -xe

FULL_SERVICE_NAME=$1
BRANCH=$2

REPO_DIR=~/go/src/${FULL_SERVICE_NAME}
if [ ! -d ${REPO_DIR} ]; then
  git clone https://github.com/guoqingXibei/openai.git ${REPO_DIR}
fi
cd ${REPO_DIR}
git fetch && git reset --hard origin/${BRANCH}
/usr/local/go/bin/go build -o ${FULL_SERVICE_NAME}
