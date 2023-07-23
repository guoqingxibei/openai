#!/bin/bash -xe

ENV=$1
BRANCH=$ENV
SERVICE_NAME=$ENV

HK=47.56.184.46
JUMP=52.29.29.173

ssh root@$HK "cd /root/${SERVICE_NAME}/ && git fetch && git reset --hard origin/${BRANCH}"
ssh guoqingj@$JUMP "./build_brother.sh ${ENV}"
ssh root@$HK "rm -f /root/${SERVICE_NAME}/${SERVICE_NAME}"
cd /tmp
/Users/guoqingj/src/cloudmonitoring/tools/assistant.sh cp exp 192.168.9.99 /home/guoqingj/go/src/${SERVICE_NAME}/${SERVICE_NAME} && scp ${SERVICE_NAME} root@$HK:/root/${SERVICE_NAME}/
ssh root@$HK "chown -R ${SERVICE_NAME}:${SERVICE_NAME} /root/${SERVICE_NAME}/"
ssh root@$HK "systemctl restart ${SERVICE_NAME}"
