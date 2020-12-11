#!/bin/bash -ex

if istioctl version | grep -E -- '1.7|1.8'
then
  exit 0
else
  printf "Encountered unsupported version of Istio: \n$(istioctl version)"
  exit 1
fi
