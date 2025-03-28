#!/usr/bin/env bash

set -euo pipefail
source ./tests/testing.sh

./szero -h

kubectl apply -f ./tests/deployment.yaml
kubectl apply -f ./tests/statefulset.yaml
kubectl wait --for=condition=available deployment/testdeployment001 --timeout=60s
kubectl wait --for=condition=ready pod -l app=teststatefulset001 --timeout=60s

assert "kubectl get pods --no-headers | wc -l" "5"
assert "kubectl get deployment testdeployment001 -o jsonpath='{.spec.replicas}'" "3"
assert "kubectl get sts teststatefulset001 -o jsonpath='{.spec.replicas}'" "2"
./szero down
assert "kubectl get deployment testdeployment001 -o jsonpath='{.spec.replicas}'" "0"
assert "kubectl get sts teststatefulset001 -o jsonpath='{.spec.replicas}'" "0"
assert_eventually "kubectl get pods --no-headers | wc -l" "0"
./szero up
assert "kubectl get deployment testdeployment001 -o jsonpath='{.spec.replicas}'" "3"
assert "kubectl get sts teststatefulset001 -o jsonpath='{.spec.replicas}'" "2"
assert_eventually "kubectl get pods --no-headers | wc -l" "5"
