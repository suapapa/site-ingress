#!/bin/bash
IMAGE_TAG=gcr.io/homin-dev/ingress:latest 
docker buildx build --platform linux/amd64 -t $IMAGE_TAG .
docker push $IMAGE_TAG
sleep 1
kubectl delete -f ingress-deploy.yaml
sleep 1
kubectl apply -f ingress-deploy.yaml