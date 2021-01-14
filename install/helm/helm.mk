HELM_ROOTDIR ?= .
GLOOMESH_CHART_DIR := $(HELM_ROOTDIR)/gloo-mesh
CA_CHART_DIR := $(HELM_ROOTDIR)/cert-agent
GLOOMESH_CRDS_CHART_DIR := $(HELM_ROOTDIR)/gloo-mesh-crds
AGENT_CRDS_CHART_DIR := $(HELM_ROOTDIR)/agent-crds
HELM_OUTPUT_DIR := $(HELM_ROOTDIR)/_output
CHART_OUTPUT_DIR := $(HELM_OUTPUT_DIR)/charts

.PHONY: clean-helm
clean-helm:
	rm -rf $(HELM_OUTPUT_DIR)
	rm -f $(GLOOMESH_CHART_DIR)/Chart.yaml
	rm -f $(GLOOMESH_CHART_DIR)/values.yaml
	rm -f $(CA_CHART_DIR)/Chart.yaml
	rm -f $(CA_CHART_DIR)/values.yaml
	rm -f $(GLOOMESH_CRDS_CHART_DIR)/Chart.yaml
	rm -f $(AGENT_CRDS_CHART_DIR)/Chart.yaml

.PHONY: package-helm
package-helm: chart-gen fmt
	helm package --destination $(CHART_OUTPUT_DIR)/gloo-mesh-crds $(GLOOMESH_CRDS_CHART_DIR)
	helm package --destination $(CHART_OUTPUT_DIR)/agent-crds $(AGENT_CRDS_CHART_DIR)
	helm package --destination $(CHART_OUTPUT_DIR)/gloo-mesh $(GLOOMESH_CHART_DIR)
	helm package --destination $(CHART_OUTPUT_DIR)/cert-agent $(CA_CHART_DIR)

.PHONY: fetch-helm
fetch-helm:
	gsutil -m rsync -r gs://gloo-mesh/gloo-mesh $(CHART_OUTPUT_DIR)/gloo-mesh
	gsutil -m rsync -r gs://gloo-mesh/gloo-mesh-crds $(CHART_OUTPUT_DIR)/gloo-mesh-crds
	gsutil -m rsync -r gs://gloo-mesh/cert-agent $(CHART_OUTPUT_DIR)/cert-agent
	gsutil -m rsync -r gs://gloo-mesh/agent-crds $(CHART_OUTPUT_DIR)/agent-crds

.PHONY: index-helm
index-helm: package-helm fetch-helm
	helm repo index $(CHART_OUTPUT_DIR)/gloo-mesh
	helm repo index $(CHART_OUTPUT_DIR)/gloo-mesh-crds
	helm repo index $(CHART_OUTPUT_DIR)/cert-agent
	helm repo index $(CHART_OUTPUT_DIR)/agent-crds

.PHONY: publish-chart
publish-chart: index-helm
ifeq ($(RELEASE),"true")
	gsutil -h "Cache-Control:no-cache,max-age=0" -m rsync -r $(CHART_OUTPUT_DIR)/gloo-mesh gs://gloo-mesh/gloo-mesh
	gsutil -h "Cache-Control:no-cache,max-age=0" -m rsync -r $(CHART_OUTPUT_DIR)/gloo-mesh-crds gs://gloo-mesh/gloo-mesh-crds
	gsutil -h "Cache-Control:no-cache,max-age=0" -m rsync -r $(CHART_OUTPUT_DIR)/cert-agent gs://gloo-mesh/cert-agent
	gsutil -h "Cache-Control:no-cache,max-age=0" -m rsync -r $(CHART_OUTPUT_DIR)/agent-crds gs://gloo-mesh/agent-crds
else
	@echo "Not a release, skipping chart upload to GCS"
endif
