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
	contextutils.LoggerFrom(ctx).Infow("Getting operator config",
		zap.String("namespace", configMapNamespace))
	configMap, err := configMapClient.GetConfigMap(ctx, configMapNamespace, OperatorConfigMapName)
	if err != nil {
		return nil, err
	}
	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}
	refreshRate, err := getRefreshRate(ctx, configMap.Data)
	if err != nil {
		return nil, err
	}
	return &OperatorConfig{
		RefreshRate: refreshRate,
	}, nil
}

func getRefreshRate(ctx context.Context, data map[string]string) (time.Duration, error) {
	contextutils.LoggerFrom(ctx).Infow("Getting refresh rate",
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
