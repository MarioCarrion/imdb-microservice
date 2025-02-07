package envvar

import (
	"os"

	"github.com/joho/godotenv"

	"github.com/MarioCarrion/todo-api-microservice-example/internal"
)

//go:generate counterfeiter -generate

//counterfeiter:generate -o envvartesting/provider.gen.go . Provider

// Provider ...
type Provider interface {
	Get(key string) (string, error)
}

// Configuration ...
type Configuration struct {
	provider Provider
}

// Load read the env filename and load it into ENV for this process.
func Load(filename string) error {
	if err := godotenv.Load(filename); err != nil {
		return internal.NewErrorf(internal.ErrorCodeUnknown, "loading env var file")
	}

	return nil
}

// New ...
func New(provider Provider) *Configuration {
	return &Configuration{
		provider: provider,
	}
}

// Get returns the value from environment variable `<key>`. When an environment variable `<key>_SECURE` exists
// the provider is used for getting the value.
func (c *Configuration) Get(key string) (string, error) {
	res := os.Getenv(key)
	valSecret := os.Getenv(key + "_SECURE")

	if valSecret != "" {
		valSecretRes, err := c.provider.Get(valSecret)
		if err != nil {
			return "", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "provider.Get")
		}

		res = valSecretRes
	}

	return res, nil
}
