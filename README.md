# ingress site for Homin.dev

Live > [HERE](https://homin.dev) <

## Create ConfigMap

Links (where to go from this site):

```bash
kubectl create cm links --from-file=conf/links.yaml # create
kubectl edit cm links -o yaml # edit
```

## CICD

> Currently CICD is manual :(

Integration:

```bash 
export IMAGE_TAG=gcr.io/homin-dev/ingress:latest 
docker buildx build --platform linux/amd64 -t $IMAGE_TAG .
docker push $IMAGE_TAG
```

Deployment:

```bash
kubectl apply -f k8s/deploy-ingress_proxy.yaml
```