#!/bin/bash

set -e

IMAGE_TAG=gcr.io/homin-dev/ingress:$1 
docker buildx build --platform linux/amd64 --build-arg=PROGRAM_VER=$1 -t $IMAGE_TAG .
docker push $IMAGE_TAG

if [ "$1" = "$dev" ]; then
    echo "skip push to GH"
else
    IMAGE_TAG_LATEST=gcr.io/homin-dev/ingress:latest 
    docker tag $IMAGE_TAG $IMAGE_TAG_LATEST
    docker push $IMAGE_TAG_LATEST

    git tag -a $1 -m "add tag for $1"
    git push --tags
fi


