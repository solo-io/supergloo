#!/bin/bash -ex

cluster=$0

K="kubectl --context kind-${cluster}"
${K} -n service-mesh-hub rollout status deployment networking
${K} -n service-mesh-hub rollout status deployment discovery

# sleep to allow CRDs to register
sleep 4

