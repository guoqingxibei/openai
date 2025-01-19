#!/bin/bash -e

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

echo "[${FULL_SERVICE_NAME}] Building..."
if ! docker info >/dev/null 2>&1; then
  echo "Docker is not running, launching it..."
  open --background -a Docker
  while ! docker info >/dev/null 2>&1; do
    echo "Waiting util docker is ready..."
    sleep 2
  done
  echo "Docker is running now"
fi

IMAGE=golang:1.22
WORKDIR=/app
PROXY_SERVER=http://host.docker.internal:1087 # v2ray proxy
OPTIONS="--rm -v .:${WORKDIR} -v ./../openai-temp/go-pkg-mod:/go/pkg/mod -v ./../openai-temp/bins:/bins -w ${WORKDIR}"
INTERNAL_BIN_PATH=/bins/${FULL_SERVICE_NAME}
docker run ${OPTIONS} ${IMAGE} go build -o ${INTERNAL_BIN_PATH}

echo "[${FULL_SERVICE_NAME}] Syncing..."
HK=root@47.56.184.46
EXTERNAL_BIN_PATH=../openai-temp/bins/${FULL_SERVICE_NAME}
rsync -azq --progress ${EXTERNAL_BIN_PATH} ./resource $HK:/root/${FULL_SERVICE_NAME}/

echo "[${FULL_SERVICE_NAME}] Restarting..."
ssh ${HK} "chown -R ${FULL_SERVICE_NAME}:${FULL_SERVICE_NAME} /root/${FULL_SERVICE_NAME}/ \
&& systemctl restart ${FULL_SERVICE_NAME}"

echo "[${FULL_SERVICE_NAME}] Complete"
