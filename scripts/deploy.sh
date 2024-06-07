#!/bin/bash -xe

# usage:
#  scripts/deploy.sh openai staging
#  scripts/deploy.sh openai
#  scripts/deploy.sh brother

SERVICE_NAME=$1 # openai or brother
ENV=$2 # staging or empty

if [ "$ENV" = "staging" ]; then
  BRANCH="staging"
  FULL_SERVICE_NAME="${SERVICE_NAME}-staging"
else
  BRANCH="main"
  FULL_SERVICE_NAME="${SERVICE_NAME}"
fi

# build
IMAGE=golang:1.22
WORKDIR=/app
BIN_PATH=temp/bins/${FULL_SERVICE_NAME}
PROXY_SERVER=http://10.221.14.35:3128
OPTIONS="--rm -v .:${WORKDIR} -v ./temp/go-pkg-mod:/go/pkg/mod -w ${WORKDIR} -e https_proxy=${PROXY_SERVER}"
docker run ${OPTIONS} ${IMAGE} go build -o ${BIN_PATH}

HK=47.56.184.46
ssh root@$HK "cd /root/${FULL_SERVICE_NAME}/ \
&& git fetch && git reset --hard origin/${BRANCH} \
&& rm -f /root/${FULL_SERVICE_NAME}/${FULL_SERVICE_NAME}"

scp ${BIN_PATH} root@$HK:/root/${FULL_SERVICE_NAME}/

ssh root@$HK "chown -R ${FULL_SERVICE_NAME}:${FULL_SERVICE_NAME} /root/${FULL_SERVICE_NAME}/ \
&& systemctl restart ${FULL_SERVICE_NAME}"
