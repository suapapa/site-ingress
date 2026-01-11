#!/bin/bash
set -e

CR=icn.vultrcr.com/homincr1
IMAGE_VER=latest

RESTART=false
ARGS=()
for arg in "$@"; do
    if [[ "$arg" == "-r" ]]; then
        RESTART=true
    else
        ARGS+=("$arg")
    fi
done
set -- "${ARGS[@]}"
PROGRAM_VER=$1

# Backend
IMAGE_TAG_BACKEND=$CR/ingress-backend:$IMAGE_VER
echo "Building Backend: $IMAGE_TAG_BACKEND"
docker buildx build --platform linux/amd64 --build-arg=PROGRAM_VER=$PROGRAM_VER -t $IMAGE_TAG_BACKEND .
docker push $IMAGE_TAG_BACKEND

# Frontend
IMAGE_TAG_FRONTEND=$CR/ingress-frontend:$IMAGE_VER
echo "Building Frontend: $IMAGE_TAG_FRONTEND"
docker buildx build --platform linux/amd64 -f frontend/Dockerfile -t $IMAGE_TAG_FRONTEND .
docker push $IMAGE_TAG_FRONTEND

if [ "$RESTART" = true ]; then
    echo "Restarting deployment..."
    kubectl rollout restart deployment ingress
fi
