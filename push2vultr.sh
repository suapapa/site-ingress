#!/bin/bash
set -e

CR=icn.vultrcr.com/homincr1
IMAGE_VER=latest

DO_BACKEND=false
DO_FRONTEND=false
RESTART=false
ARGS=()

for arg in "$@"; do
    case $arg in
        -r)
            RESTART=true
            ;;
        -b)
            DO_BACKEND=true
            ;;
        -f)
            DO_FRONTEND=true
            ;;
        *)
            ARGS+=("$arg")
            ;;
    esac
done

# If neither flag is set, do both (default)
if [ "$DO_BACKEND" = false ] && [ "$DO_FRONTEND" = false ]; then
    DO_BACKEND=true
    DO_FRONTEND=true
fi

set -- "${ARGS[@]}"
PROGRAM_VER=$1

# Backend
if [ "$DO_BACKEND" = true ]; then
    IMAGE_TAG_BACKEND=$CR/ingress-backend:$IMAGE_VER
    echo "Building Backend: $IMAGE_TAG_BACKEND"
    docker buildx build --platform linux/amd64 --build-arg=PROGRAM_VER=$PROGRAM_VER -t $IMAGE_TAG_BACKEND .
    docker push $IMAGE_TAG_BACKEND
fi

# Frontend
if [ "$DO_FRONTEND" = true ]; then
    IMAGE_TAG_FRONTEND=$CR/ingress-frontend:$IMAGE_VER
    echo "Building Frontend: $IMAGE_TAG_FRONTEND"
    docker buildx build --platform linux/amd64 -f frontend/Dockerfile -t $IMAGE_TAG_FRONTEND .
    docker push $IMAGE_TAG_FRONTEND
fi

if [ "$RESTART" = true ]; then
    echo "Restarting deployment..."
    kubectl rollout restart deployment ingress
fi
