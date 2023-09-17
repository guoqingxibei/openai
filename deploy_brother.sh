#!/bin/bash -xe

BRANCH=main
SERVICE_NAME=brother

HK=47.56.184.46
VM=10.221.14.56

ssh root@$HK "cd /root/${SERVICE_NAME}/ && git fetch && git reset --hard origin/${BRANCH}"
ssh guoqingj@$VM "./build_brother.sh ${ENV}"
ssh root@$HK "rm -f /root/${SERVICE_NAME}/${SERVICE_NAME}"
cd /tmp
scp guoqingj@$VM:~/go/src/${SERVICE_NAME}/${SERVICE_NAME} .
scp ${SERVICE_NAME} root@$HK:/root/${SERVICE_NAME}/
ssh root@$HK "chown -R ${SERVICE_NAME}:${SERVICE_NAME} /root/${SERVICE_NAME}/"
ssh root@$HK "systemctl restart ${SERVICE_NAME}"
