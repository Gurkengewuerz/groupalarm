#!/bin/bash

USERNAME="mc8051"
PROJECT="groupalarm"
REGISTRY="reg.mc8051.de"

docker buildx create --use --name ${USERNAME}-${PROJECT}
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --push \
  -t ${REGISTRY}/${USERNAME}/${PROJECT}:latest \
  -f docker/Dockerfile \
  .
docker buildx stop ${USERNAME}-${PROJECT}
docker buildx rm ${USERNAME}-${PROJECT}

echo "Done!"
