package config

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/configutils"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
)

const (
	OperatorConfigMapName = "operator-config"

	RefreshRateKey     = "refreshRate"
	DefaultRefreshRate = time.Second
)

type OperatorConfig struct {
	RefreshRate time.Duration
}

func GetOperatorConfig(ctx context.Context, configMapClient configutils.ConfigMapClient, configMapNamespace string) (*OperatorConfig, error) {
	contextutils.LoggerFrom(ctx).Debugw("Getting operator config",
		zap.String("namespace", configMapNamespace))

	data := make(map[string]string)
	configMap, err := configMapClient.GetConfigMap(ctx, configMapNamespace, OperatorConfigMapName)
	if err != nil {
		contextutils.LoggerFrom(ctx).Debugw("Operator config map could not be found, using defaults",
			zap.String("namespace", configMapNamespace))

	} else {
		if configMap.Data != nil {
			data = configMap.Data
		}
	}

	refreshRate, err := getRefreshRate(ctx, data)
	if err != nil {
		return nil, err
	}
	return &OperatorConfig{
		RefreshRate: refreshRate,
	}, nil
}

func getRefreshRate(ctx context.Context, data map[string]string) (time.Duration, error) {
	contextutils.LoggerFrom(ctx).Debugw("Getting refresh rate",
		zap.Any("data", data),
		zap.String("key", RefreshRateKey),
		zap.Int64("default", DefaultRefreshRate.Nanoseconds()))
	refreshRate := data[RefreshRateKey]
	if refreshRate == "" {
		return DefaultRefreshRate, nil
	}
	duration, err := time.ParseDuration(refreshRate)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Could not parse refresh rate",
			zap.String("refreshRate", refreshRate),
			zap.Error(err))
		return DefaultRefreshRate, err
	}
	return duration, nil
}
