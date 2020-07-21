HELM_ROOTDIR ?= .
CHART_DIR := $(HELM_ROOTDIR)/service-mesh-hub
HELM_OUTPUT_DIR := $(HELM_ROOTDIR)/_output
PACKAGED_CHARTS_DIR := $(HELM_OUTPUT_DIR)/charts

# Make sure output directories exist
_ := $(shell mkdir -p $(PACKAGED_CHARTS_DIR))

.PHONY: clean-helm
clean-helm:
	rm -rf $(HELM_OUTPUT_DIR)
	rm $(CHART_DIR)/Chart.yaml
	rm $(CHART_DIR)/values.yamls

.PHONY: update-chart-dependencies
update-chart-dependencies:
	helm repo add ext-auth-service https://storage.googleapis.com/ext-auth-service-helm
	helm repo add rate-limiter https://storage.googleapis.com/rate-limiter-helm
	helm dependency update $(CHART_DIR)

.PHONY: package-helm
package-helm: update-chart-dependencies
	helm package --destination $(PACKAGED_CHARTS_DIR) $(CHART_DIR)

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