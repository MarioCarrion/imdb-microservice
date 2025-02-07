package vault

import (
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"

	"github.com/MarioCarrion/todo-api-microservice-example/internal"
)

// Provider ...
type Provider struct {
	path    string
	client  *api.Logical
	results map[string]map[string]string
}

// New ...
func New(token, addr, path string) (*Provider, error) {
	config := &api.Config{
		Address: addr,
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "api.NewClient")
	}

	client.SetToken(token)

	return &Provider{
		path:    path,
		client:  client.Logical(),
		results: make(map[string]map[string]string),
	}, nil
}

// Get retrieves a value from vault using the KV engine. The actual key selected is determined by the value
// separated by the colon. For example "database:password" will retrieve the key "password" from the path
// "database".
func (p *Provider) Get(v string) (string, error) {
	// <path>/data/<path-secret>:key
	split := strings.Split(v, ":")
	if len(split) == 1 {
		return "", internal.NewErrorf(internal.ErrorCodeUnknown, "missing key value")
	}

	pathSecret := split[0]
	key := split[1]

	res, ok := p.results[pathSecret]
	if ok {
		val, ok := res[key]
		if !ok {
			return "", internal.NewErrorf(internal.ErrorCodeUnknown, "key not found in cached data")
		}

		return val, nil
	}

	secret, err := p.client.Read(fmt.Sprintf("%s/data/%s", p.path, pathSecret))
	if err != nil {
		return "", internal.WrapErrorf(err, internal.ErrorCodeUnknown, "reading")
	}

	if secret == nil {
		return "", internal.NewErrorf(internal.ErrorCodeUnknown, "secret not found")
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return "", internal.NewErrorf(internal.ErrorCodeUnknown, "invalid data in secret")
	}

	secrets := make(map[string]string)

	for k, v := range data {
		val, ok := v.(string)
		if !ok {
			return "", internal.NewErrorf(internal.ErrorCodeUnknown, "secret value in data is not string")
		}

		secrets[k] = val
	}

	val, ok := secrets[key]
	if !ok {
		return "", internal.NewErrorf(internal.ErrorCodeUnknown, "key not found in retrieved data")
	}

	p.results[pathSecret] = secrets

	return val, nil
}
