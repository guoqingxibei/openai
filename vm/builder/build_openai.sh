#!/bin/bash -xe

ENV=$1
if [ "$1" = "staging" ]; then
  BRANCH="staging"
  SERVICE_NAME="openai-staging"
else
  BRANCH="main"
  SERVICE_NAME="openai"
fi

cd ~/go/src/${SERVICE_NAME}/
git fetch && git reset --hard origin/${BRANCH}
/usr/bin/go build -o ${SERVICE_NAME}
