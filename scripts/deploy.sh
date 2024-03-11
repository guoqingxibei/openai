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

VM=us-jumpbox
HK=47.56.184.46

ssh root@$HK "cd /root/${FULL_SERVICE_NAME}/ && git fetch && git reset --hard origin/${BRANCH}"
ssh root@$HK "rm -f /root/${FULL_SERVICE_NAME}/${FULL_SERVICE_NAME}"

ssh guoqingj@$VM "mkdir -p ~/openai/"
scp scripts/build.sh guoqingj@$VM:~/openai/
ssh guoqingj@$VM "chmod +x openai/build.sh && ./openai/build.sh ${FULL_SERVICE_NAME} ${BRANCH}"
scp guoqingj@$VM:~/go/src/${FULL_SERVICE_NAME}/${FULL_SERVICE_NAME} root@$HK:/root/${FULL_SERVICE_NAME}/

ssh root@$HK "chown -R ${FULL_SERVICE_NAME}:${FULL_SERVICE_NAME} /root/${FULL_SERVICE_NAME}/"
ssh root@$HK "systemctl restart ${FULL_SERVICE_NAME}"
