#!/bin/bash -xe

BRANCH=main
SERVICE_NAME=brother

cd ~/go/src/${SERVICE_NAME}/
git fetch && git reset --hard origin/${BRANCH}
/usr/bin/go build -o ${SERVICE_NAME}
