#!/usr/bin/env bash

set -euo pipefail
source ./tests/testing.sh

./szero -h

kubectl apply -f ./tests/deployment.yaml
kubectl apply -f ./tests/statefulset.yaml
kubectl apply -f ./tests/daemonset.yaml

echo "Waiting for resources to be ready..."

kubectl wait --for=condition=available deployment/testdeployment001 --timeout=60s
kubectl wait --for=condition=ready pod -l app=teststatefulset001 --timeout=60s
kubectl wait --for=condition=ready pod -l app=testdaemonset001 --timeout=60s

assert "kubectl get pods --no-headers | wc -l" "6"
assert "kubectl get deployment testdeployment001 -o jsonpath='{.spec.replicas}'" "3"
assert "kubectl get sts teststatefulset001 -o jsonpath='{.spec.replicas}'" "2"
assert "kubectl get ds testdaemonset001 -o jsonpath='{.status.desiredNumberScheduled}'" "1"

echo "Test dry-run mode (should not change anything)"

./szero down --dry-run

assert "kubectl get deployment testdeployment001 -o jsonpath='{.spec.replicas}'" "3"
assert "kubectl get sts teststatefulset001 -o jsonpath='{.spec.replicas}'" "2"
assert "kubectl get ds testdaemonset001 -o jsonpath='{.status.desiredNumberScheduled}'" "1"
assert "kubectl get pods --no-headers | wc -l" "6"

echo "Test actual downscale"
./szero down --wait

assert "kubectl get deployment testdeployment001 -o jsonpath='{.spec.replicas}'" "0"
assert "kubectl get sts teststatefulset001 -o jsonpath='{.spec.replicas}'" "0"
assert "kubectl get ds testdaemonset001 -o jsonpath='{.status.desiredNumberScheduled}'" "0"
assert_eventually "kubectl get pods --no-headers | wc -l" "0"

echo "Test dry-run mode for upscale (should not change anything)"
./szero up --dry-run

assert "kubectl get deployment testdeployment001 -o jsonpath='{.spec.replicas}'" "0"
assert "kubectl get sts teststatefulset001 -o jsonpath='{.spec.replicas}'" "0"
assert "kubectl get ds testdaemonset001 -o jsonpath='{.status.desiredNumberScheduled}'" "0"
assert "kubectl get pods --no-headers | wc -l" "0"

echo "Test actual upscale"
./szero up --wait

assert "kubectl get deployment testdeployment001 -o jsonpath='{.spec.replicas}'" "3"
assert "kubectl get sts teststatefulset001 -o jsonpath='{.spec.replicas}'" "2"
assert_eventually "kubectl get pods --no-headers | wc -l" "6"
assert "kubectl get ds testdaemonset001 -o jsonpath='{.status.desiredNumberScheduled}'" "1"
