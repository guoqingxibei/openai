#!/bin/bash -xe

ENV=$1
if [ "$1" = "staging" ]; then
  BRANCH="staging"
  SERVICE_NAME="openai-staging"
else
  BRANCH="main"
  SERVICE_NAME="openai"
fi

VM=10.221.14.56
ssh guoqingj@$VM "./build_openai.sh ${ENV}"
scp guoqingj@$VM:~/go/src/${SERVICE_NAME}/${SERVICE_NAME} /tmp/

HK=47.56.184.46
ssh root@$HK "cd /root/${SERVICE_NAME}/ && git fetch && git reset --hard origin/${BRANCH}"
ssh root@$HK "rm -f /root/${SERVICE_NAME}/${SERVICE_NAME}"
scp /tmp/${SERVICE_NAME} root@$HK:/root/${SERVICE_NAME}/
ssh root@$HK "chown -R ${SERVICE_NAME}:${SERVICE_NAME} /root/${SERVICE_NAME}/"
ssh root@$HK "systemctl restart ${SERVICE_NAME}"
