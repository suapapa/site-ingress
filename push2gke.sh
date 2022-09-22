#!/bin/bash
git tag -a $1 -m "add tag for $1"
git push --tags

IMAGE_TAG=gcr.io/homin-dev/ingress:$1 
docker buildx build --platform linux/amd64 -t $IMAGE_TAG .
docker push $IMAGE_TAG

# IMAGE_TAG_LATEST=gcr.io/homin-dev/ingress:latest 
# docker tag $IMAGE_TAG $IMAGE_TAG_LATEST
# docker push $IMAGE_TAG_LATEST