HELM_ROOTDIR ?= .
SMH_CHART_DIR := $(HELM_ROOTDIR)/service-mesh-hub
CA_CHART_DIR := $(HELM_ROOTDIR)/cert-agent
HELM_OUTPUT_DIR := $(HELM_ROOTDIR)/_output
PACKAGED_CHARTS_DIR := $(HELM_OUTPUT_DIR)/charts

# Make sure output directories exist
_ := $(shell mkdir -p $(PACKAGED_CHARTS_DIR))

.PHONY: clean-helm
clean-helm:
	rm -rf $(HELM_OUTPUT_DIR)
	rm $(SMH_CHART_DIR)/Chart.yaml
	rm $(SMH_CHART_DIR)/values.yamls
	rm $(CA_CHART_DIR)/Chart.yaml
	rm $(CA_CHART_DIR)/values.yamls

.PHONY: package-helm
package-helm:
	helm package --destination $(PACKAGED_CHARTS_DIR) $(SMH_CHART_DIR)
	helm package --destination $(PACKAGED_CHARTS_DIR) $(CA_CHART_DIR)

.PHONY: fetch-helm
fetch-helm:
	gsutil -m rsync -r gs://service-mesh-hub/ $(HELM_OUTPUT_DIR)

.PHONY: index-helm
index-helm:
	helm repo index $(HELM_OUTPUT_DIR)

.PHONY: push-helm
push-helm:
ifeq ($(RELEASE),"true")
	gsutil -m rsync -r $(HELM_OUTPUT_DIR) gs://service-mesh-hub
else
	@echo "Not a release, skipping chart upload to GCS"
endif

.PHONY: publish-chart
publish-chart: package-helm fetch-helm index-helm push-helm