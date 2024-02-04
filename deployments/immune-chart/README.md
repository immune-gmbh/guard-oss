# Immune

Immune Helm Chart for easy deployment of Immune Guard on Prem

## Deployment Guide

Add the docker registry secret:

```sh
kubectl create secret docker-registry ghcr-secret --docker-server=https://ghcr.io --docker-username=<USERNAME> --docker-password=<GITHUB TOKEN>
```

Fill out all the necessary fields in `onprem-values.yaml`.
To see what needs to be filled out, take a look at `values.yaml`.

To make helm use the container registry secret above, add the following lines:

```yml
imagePullSecrets:
  - name: ghcr-secret
```

Then deploy using `helm`:

```sh
helm install -f onprem-values.yaml <deployment-name> .
```
