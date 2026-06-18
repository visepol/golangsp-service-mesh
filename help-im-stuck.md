eval "$(make k8s-env)"

kubectl -n no-mesh rollout restart deploy
kubectl -n mesh    rollout restart deploy

#no mesh
kubectl -n no-mesh get endpoints go-api
kubectl -n no-mesh get services -o wide
kubectl -n no-mesh get pods -o wide

#mesh
kubectl -n mesh get endpoints go-api
kubectl -n mesh get services -o wide
kubectl -n mesh get pods -o wide
kubectl describe namespace mesh