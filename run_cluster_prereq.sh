#!/usr/bin/env bash

kubectl apply -f cluster_rbac.yaml

kubectl create secret generic aerospike-secret --from-file=config/secrets -n aerospike
sleep 2

kubectl create secret generic auth-secret --from-literal=password='admin123' -n aerospike
sleep 2


