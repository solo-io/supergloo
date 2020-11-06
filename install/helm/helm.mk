HELM_ROOTDIR ?= .
GLOOMESH_CHART_DIR := $(HELM_ROOTDIR)/gloo-mesh
CA_CHART_DIR := $(HELM_ROOTDIR)/cert-agent
HELM_OUTPUT_DIR := $(HELM_ROOTDIR)/_output
CHART_OUTPUT_DIR := $(HELM_OUTPUT_DIR)/charts

.PHONY: clean-helm
clean-helm:
	rm -rf $(HELM_OUTPUT_DIR)
	rm -f $(GLOOMESH_CHART_DIR)/Chart.yaml
	rm -f $(GLOOMESH_CHART_DIR)/values.yaml
	rm -f $(CA_CHART_DIR)/Chart.yaml
	rm -f $(CA_CHART_DIR)/values.yaml

.PHONY: package-helm
package-helm: chart-gen
	helm package --destination $(CHART_OUTPUT_DIR)/gloo-mesh $(GLOOMESH_CHART_DIR)
	helm package --destination $(CHART_OUTPUT_DIR)/cert-agent $(CA_CHART_DIR)

# TODO rename GCP project from service-mesh-hub to gloo-mesh
.PHONY: fetch-helm
fetch-helm:
	gsutil -m rsync -r gs://service-mesh-hub/gloo-mesh $(CHART_OUTPUT_DIR)/gloo-mesh
	gsutil -m rsync -r gs://service-mesh-hub/cert-agent $(CHART_OUTPUT_DIR)/cert-agent

.PHONY: index-helm
index-helm: package-helm fetch-helm
	helm repo index $(CHART_OUTPUT_DIR)/gloo-mesh
	helm repo index $(CHART_OUTPUT_DIR)/cert-agent

.PHONY: publish-chart
publish-chart: index-helm
ifeq ($(RELEASE),"true")
	gsutil -m rsync -r $(CHART_OUTPUT_DIR)/gloo-mesh gs://service-mesh-hub/gloo-mesh
	gsutil -m rsync -r $(CHART_OUTPUT_DIR)/cert-agent gs://service-mesh-hub/cert-agent
else
	@echo "Not a release, skipping chart upload to GCS"
endif
