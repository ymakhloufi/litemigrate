package store

import (
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/go-playground/validator/v10"
)

type Config struct {
	Driver string `env:"DRIVER,required"`
	Host   string `env:"HOST,required"   validate:"required"`
	Port   uint16 `env:"PORT,required"   validate:"required,min=1,max=65535"`
	User   string `env:"USER,required"   validate:"required,ascii"`
	Pass   string `env:"PASS,required"   validate:"required"`
	DB     string `env:"DB,required"     validate:"required,ascii"`
	SSL    bool   `env:"SSL"             validate:"boolean"`
}

func (c Config) ToConnectionString() string {
	sslMode := "disable"
	if c.SSL {
		sslMode = "enable"
	}
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
		c.Driver, c.User, c.Pass, c.Host, c.Port, c.DB, sslMode)
}

// parseEnvIntoStruct parses environment variables into a given struct
func parseEnvIntoStruct(config interface{}) error {
	if err := env.Parse(config); err != nil {
		return err
	}
	return validator.New().Struct(config)
}

// UserAuthConfigFromEnv loads the database configuration from the environment
func UserAuthConfigFromEnv() Config {
	cfg := Config{}
	if err := parseEnvIntoStruct(&cfg); err != nil {
		panic(fmt.Sprintf("could not load db configuration: %s", err))
	}
	return cfg
}
