# ingress site for Homin.dev

Live > [HERE](https://homin.dev/ingress) <

## CICD

> Currently CICD is manual :(

Integration:

```bash 
export IMAGE_TAG=gcr.io/homin-dev/ingress:latest 
docker buildx build --platform linux/amd64 -t $IMAGE_TAG .
docker push $IMAGE_TAG
```

Deployment:

> k8s configs are move to [suapapa/k8s-homin.dev](https://github.com/suapapa/k8s-homin.dev)

```bash
k apply -f cm/ingress-links.yaml deploy/deploy-ingress_proxy.yaml
```