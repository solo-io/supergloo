#!/bin/bash -ex

cluster=$0

K="kubectl --context kind-${cluster}"
${K} -n gloo-mesh rollout status deployment networking
${K} -n gloo-mesh rollout status deployment discovery

# sleep to allow CRDs to register
sleep 4

