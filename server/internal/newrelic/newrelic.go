package newrelic

import (
	"github.com/newrelic/go-agent/v3/newrelic"
	"go.uber.org/zap"
)

type NewRelicConfig struct {
	AppName    string `env:"APP_NAME,required"`
	AppLicense string `env:"LICENSE,required"`
}

type NewRelicClient struct {
	Client *newrelic.Application
}

func New(nrc NewRelicConfig) *NewRelicClient {

	zap.S().Info("Initializing New Relic Client... ")

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(nrc.AppName),
		newrelic.ConfigLicense(nrc.AppLicense),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)

	if err != nil {
		zap.S().Errorf("could not instantiate new relic client: %w", err)
		panic(err)
	}

	zap.S().Info("New Relic Client Initialized!")

	return &NewRelicClient{
		Client: app,
	}
}
