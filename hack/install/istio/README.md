# Notes on Istio helm chart

We've forked the istio helm chart to make several modifications, listed here. To re-package the chart, 
navigate into the istio-1.0.3 directory and run: 

On Linux: `tar -zcvf istio-1.0.3.tgz istio` 

On Mac: `tar --disable-copyfile -zcvf istio-1.0.3.tgz istio`

# Change log

These are changes made to the standard istio chart. 

1. Change boolean logic so installing CRDs can be turned off (`istio/templates/crds.yaml`) 
2. Support post-upgrade as a hook for istio-security-post-install job (`istio/charts/security/templates/create-custom-resources-job.yaml`)