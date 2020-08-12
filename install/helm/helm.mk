HELM_ROOTDIR ?= .
SMH_CHART_DIR := $(HELM_ROOTDIR)/service-mesh-hub
CA_CHART_DIR := $(HELM_ROOTDIR)/cert-agent
HELM_OUTPUT_DIR := $(HELM_ROOTDIR)/_output
CHART_OUTPUT_DIR := $(HELM_OUTPUT_DIR)/charts

.PHONY: clean-helm
clean-helm:
	rm -rf $(HELM_OUTPUT_DIR)
	rm -f $(SMH_CHART_DIR)/Chart.yaml
	rm -f $(SMH_CHART_DIR)/values.yaml
	rm -f $(CA_CHART_DIR)/Chart.yaml
	rm -f $(CA_CHART_DIR)/values.yaml

.PHONY: package-helm
package-helm: chart-gen
	helm package --destination $(CHART_OUTPUT_DIR)/service-mesh-hub $(SMH_CHART_DIR)
	helm package --destination $(CHART_OUTPUT_DIR)/cert-agent $(CA_CHART_DIR)

.PHONY: fetch-helm
fetch-helm:
	gsutil -m rsync -r gs://service-mesh-hub/service-mesh-hub $(CHART_OUTPUT_DIR)/service-mesh-hub
	gsutil -m rsync -r gs://service-mesh-hub/cert-agent $(CHART_OUTPUT_DIR)/cert-agent

.PHONY: index-helm
index-helm: package-helm fetch-helm
	helm repo index $(CHART_OUTPUT_DIR)/service-mesh-hub
	helm repo index $(CHART_OUTPUT_DIR)/cert-agent

.PHONY: publish-chart
publish-chart: index-helm
ifeq ($(RELEASE),"true")
	gsutil -m rsync -r $(CHART_OUTPUT_DIR)/service-mesh-hub gs://service-mesh-hub/service-mesh-hub
	gsutil -m rsync -r $(CHART_OUTPUT_DIR)/cert-agent gs://service-mesh-hub/cert-agent
else
	@echo "Not a release, skipping chart upload to GCS"
endif
