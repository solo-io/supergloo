#!/bin/bash

meshctlBinaryName=%s
clusterName=%s
contextName=%s
useDevCsrChart=%s

if [ -n "$useDevCsrChart" ]; then
  csrChartArg='--dev-csr-agent-chart --local-cluster-domain-override=host.docker.internal'
fi

$meshctlBinaryName cluster register --remote-context $contextName --remote-cluster-name $clusterName $csrChartArg
