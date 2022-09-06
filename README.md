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

> k8s configs are move to [suapapa/k8s-homin.dev](https://github.com/suapapa/k8s-homin.dev)

```bash
kubectl apply -f cm/ingress-links.yaml deploy/deploy-ingress_proxy.yaml
```

## SSL Cert

### Create

Certbot which is a CLI tool for Let'sEncrypt needs iteractive for create first SSL cert. So;

Connect the POD:

```bash
kubectl exec -it ingress-proxy-54d459b8bd-2pqxc -- /bin/sh
```
And, Create new cert in the POD:

```bash
certbot certonly --webroot --webroot-path /tmp/letsencrypt -m "ff4500@gmail.com" -d "homin.dev"  --agree-tos
```

### Renew

The SSL cert will be renewed by cron. Check `Dockerfile` for the detail