$ErrorActionPreference = "Stop"
minikube docker-env | Invoke-Expression
docker build -t rate-limiter:latest .
kubectl apply -f k8s/redis.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/hpa.yaml
kubectl rollout status deployment/rate-limiter
$ip = minikube ip
Write-Host "Service URL: http://${ip}:30080"
