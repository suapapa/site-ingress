#!/bin/bash
set -e
CR=icn.vultrcr.com/homincr1
IMAGE_TAG=$CR/ingress:$1 
docker buildx build --platform linux/amd64 --build-arg=PROGRAM_VER=$1 -t $IMAGE_TAG .
docker push $IMAGE_TAG
