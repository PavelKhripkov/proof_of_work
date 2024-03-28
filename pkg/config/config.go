package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

// LoadConfig parses env variables into provided provided struct.
func LoadConfig(prefix string, spec interface{}) error {
	err := envconfig.Process(prefix, spec)
	if err != nil {
		return errors.Wrap(err, "couldn't parse config from env variables")
	}

	return nil
}
