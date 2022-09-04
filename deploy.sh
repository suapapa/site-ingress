#!/bin/bash
IMAGE_TAG=gcr.io/homin-dev/ingress:latest 
docker build -t $IMAGE_TAG .
docker push $IMAGE_TAG
kubectl delete -f ingress-deploy.yaml
sleep 1
kubectl apply -f ingress-deploy.yaml