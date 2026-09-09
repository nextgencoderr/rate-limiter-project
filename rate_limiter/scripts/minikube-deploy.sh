#!/usr/bin/env bash
set -euo pipefail

eval "$(minikube docker-env)"
docker build -t rate-limiter:latest .
kubectl apply -f k8s/redis.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/hpa.yaml
kubectl rollout status deployment/rate-limiter
echo "Service URL: http://$(minikube ip):30080"
