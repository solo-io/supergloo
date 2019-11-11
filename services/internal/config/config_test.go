package config_test

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/configutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/solo-io/mesh-projects/services/internal/config"
)

var _ = Describe("ConfigTest", func() {

	getMockClient := func(data map[string]string, err error) configutils.ConfigMapClient {
		return &configutils.MockConfigMapClient{
			Data:     data,
			SetError: err,
			GetError: err,
		}
	}

	getActual := func(refreshRateString string) (*config.OperatorConfig, error) {
		data := map[string]string{
			config.RefreshRateKey: refreshRateString,
		}
		client := getMockClient(data, nil)
		return config.GetOperatorConfig(context.TODO(), client, "test-namespace")
	}

	getExpected := func(refreshRate time.Duration) config.OperatorConfig {
		return config.OperatorConfig{
			RefreshRate: refreshRate,
		}
	}

	It("works with valid config map", func() {
		actual, err := getActual("1s")
		expected := getExpected(time.Second)
		Expect(err).NotTo(HaveOccurred())
		Expect(*actual).To(BeEquivalentTo(expected))
	})

	It("errors with invalid refreshRate value", func() {
		_, err := getActual("invalid")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(BeEquivalentTo("time: invalid duration invalid"))
	})

	It("works with missing refreshRate value by using default", func() {
		data := map[string]string{}
		client := getMockClient(data, nil)
		actual, err := config.GetOperatorConfig(context.TODO(), client, "test-namespace")
		Expect(err).NotTo(HaveOccurred())
		expected := getExpected(time.Second)
		Expect(*actual).To(BeEquivalentTo(expected))
	})

	It("does not error from client error", func() {
		expectedErr := errors.Errorf("Dummy error")
		client := getMockClient(nil, expectedErr)
		cfg, err := config.GetOperatorConfig(context.TODO(), client, "test-namespace")
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg).To(Equal(&config.OperatorConfig{
			RefreshRate: config.DefaultRefreshRate,
		}))
	})
})
