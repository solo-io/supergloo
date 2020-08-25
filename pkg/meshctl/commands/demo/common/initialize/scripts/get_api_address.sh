#!/bin/bash -ex

cluster=$0

case $(uname) in
  "Darwin")
  {
      apiServerAddress=host.docker.internal
  } ;;
  "Linux")
  {
      apiServerAddress=$(docker exec "${cluster}-control-plane" ip addr show dev eth0 | sed -nE 's|\s*inet\s+([0-9.]+).*|\1|p'):6443
  } ;;
  *)
  {
      echo "Unsupported OS"
      exit 1
  } ;;
esac
printf ${apiServerAddress}