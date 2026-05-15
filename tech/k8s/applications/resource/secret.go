package resource

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type SecretLoader interface {
	Getenv(key string) any
}

type viperSecret struct {
	instance *viper.Viper
}

func (s *viperSecret) Getenv(key string) any {
	return s.instance.Get(key)
}

func NewViperSecretLoader(opts SecretOpts) (SecretLoader, error) {
	vi := viper.New()

	if opts.EnvPrefix != "" {
		vi.SetEnvPrefix(opts.EnvPrefix)
	}
	vi.AutomaticEnv()

	if opts.ConfigType != "" {
		vi.SetConfigType(opts.ConfigType)
	}

	if opts.Path != "" {
		vi.SetConfigFile(opts.Path)
	} else {
		if opts.Name != "" {
			vi.SetConfigName(opts.Name)
		}
		for _, p := range opts.SearchPaths {
			vi.AddConfigPath(p)
		}
	}

	if err := vi.ReadInConfig(); err != nil && errors.Is(err, viper.ConfigFileNotFoundError{}) {
		return nil, fmt.Errorf("ViperSecretLoader: read config: %w", err)
	}

	return &viperSecret{
		instance: vi,
	}, nil
}
