#!/bin/bash -xe

ENV=$1
if [ "$1" = "staging" ]; then
  BRANCH="staging"
  SERVICE_NAME="openai-staging"
else
  BRANCH="main"
  SERVICE_NAME="openai"
fi

ssh guoqingj@52.29.29.173 "./build_openai.sh ${ENV}"
ssh root@jiaguoqing.ml "rm -f /root/${SERVICE_NAME}/${SERVICE_NAME}"
cd /tmp
/Users/guoqingj/src/cloudmonitoring/tools/assistant.sh cp exp 192.168.9.99 /home/guoqingj/go/src/${SERVICE_NAME}/${SERVICE_NAME} && scp ${SERVICE_NAME} root@jiaguoqing.ml:/root/${SERVICE_NAME}/
ssh root@jiaguoqing.ml "chown -R ${SERVICE_NAME}:${SERVICE_NAME} /root/${SERVICE_NAME}/"
ssh root@jiaguoqing.ml "systemctl restart ${SERVICE_NAME}"
