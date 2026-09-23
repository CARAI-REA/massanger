# Kubernetes manifests (callService)

Apply order:

```bash
kubectl apply -f deploy/k8s/00-namespace-secrets.yaml
kubectl apply -f deploy/k8s/01-rooms.yaml
kubectl apply -f deploy/k8s/02-signaling.yaml
kubectl apply -f deploy/k8s/03-sfu.yaml
kubectl apply -f deploy/k8s/04-coturn.yaml
```

Replace all `CHANGE_ME` secrets before apply. TLS for WSS is on external nginx — see [deploy/nginx/callService.conf.example](../nginx/callService.conf.example).

SFU uses `hostNetwork` so ICE UDP binds on the node. Set public TURN external IP on coturn.
