package vault_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/MarioCarrion/todo-api-microservice-example/internal/envvar/vault"
)

type vaultClient struct {
	Token   string
	Address string
	Client  *api.Client
}

//nolint:paralleltest,tparallel
func TestProvider_Get(t *testing.T) {
	t.Parallel()

	type output struct {
		res     string
		withErr bool
	}

	tests := []struct {
		name   string
		setup  func(*vaultClient) error
		input  string
		output output
	}{
		{
			"OK",
			func(v *vaultClient) error {
				if _, err := v.Client.Logical().Write("/secret/data/ok",
					map[string]interface{}{
						"data": map[string]interface{}{
							"one": "1",
							"two": "2",
						},
					}); err != nil {
					return fmt.Errorf("couldn't write: %w", err)
				}

				return nil
			},
			"/ok:one",
			output{
				res: "1",
			},
		},
		{
			"OK: cached",
			func(_ *vaultClient) error { return nil },
			"/ok:two",
			output{
				res: "2",
			},
		},
		{
			"ERR: missing key value",
			func(_ *vaultClient) error { return nil },
			"/ok",
			output{
				withErr: true,
			},
		},
		{
			"ERR: key not found in cached data",
			func(_ *vaultClient) error { return nil },
			"/ok:three",
			output{
				withErr: true,
			},
		},
		{
			"ERR: secret not found",
			func(_ *vaultClient) error { return nil },
			"/not:found",
			output{
				withErr: true,
			},
		},
		{
			"ERR: key not found in retrieved data",
			func(v *vaultClient) error {
				if _, err := v.Client.Logical().Write("/secret/data/err",
					map[string]interface{}{
						"data": map[string]interface{}{
							"hello": "world",
						},
					}); err != nil {
					return fmt.Errorf("couldn't write: %w", err)
				}

				return nil
			},
			"/err:something",
			output{
				withErr: true,
			},
		},
	}

	// Provider is not local to the subtest because we want to test the local caching logic

	client := newVault(t)
	provider, err := vault.New(client.Token, client.Address, "/secret")

	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { //nolint: wsl
			// Not calling t.Parallel() because vault.Provider is not goroutine safe.

			if err := tt.setup(client); err != nil {
				t.Fatalf("could not set up %s", err)
			}

			actualRes, actualErr := provider.Get(tt.input)

			if actualRes != tt.output.res {
				t.Fatalf("expected %s, got %s", tt.output.res, actualRes)
			}

			if (actualErr != nil) != tt.output.withErr {
				t.Fatalf("expected error %t, got %s", tt.output.withErr, actualErr)
			}
		})
	}
}

func newVault(tb testing.TB) *vaultClient {
	tb.Helper()

	pool, err := dockertest.NewPool("")

	if err != nil {
		tb.Fatalf("Couldn't connect to docker: %s", err)
	}

	pool.MaxWait = 5 * time.Second

	token := "myroot"
	address := "0.0.0.0:8300"

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "vault",
		Tag:        "1.12.4",
		Env: []string{
			"VAULT_DEV_ROOT_TOKEN_ID=" + token,
			"VAULT_DEV_LISTEN_ADDRESS=" + address,
		},
		ExposedPorts: []string{"8300/tcp"},
		PortBindings: map[docker.Port][]docker.PortBinding{ // Because of the way Vault works internally we bind to a port on the host
			"8300/tcp": {
				{
					HostIP:   "0.0.0.0",
					HostPort: "8300/tcp",
				},
			},
		},
		CapAdd: []string{"IPC_LOCK"},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		tb.Fatalf("Couldn't start resource: %s", err)
	}

	_ = resource.Expire(60)

	tb.Cleanup(func() {
		if err := pool.Purge(resource); err != nil {
			tb.Fatalf("Couldn't purge container: %s", err)
		}
	})

	address = "http://" + address

	config := &api.Config{
		Address: address,
	}

	client, err := api.NewClient(config)
	if err != nil {
		tb.Fatalf("Couldn't open new client: %s", err)
	}

	client.SetToken(token)

	if err := pool.Retry(func() error {
		_, err = client.Logical().Write("/secret/data/example",
			map[string]interface{}{
				"data": map[string]interface{}{
					"one": "two",
				},
			})
		if err != nil {
			return err
		}

		secret, err := client.Logical().Read("/secret/data/example")
		if err != nil {
			return err
		}

		if secret == nil {
			return errors.New("no secret")
		}

		return nil
	}); err != nil {
		tb.Fatalf("Couldn't retry: %s", err)
	}

	return &vaultClient{
		Client:  client,
		Token:   token,
		Address: address,
	}
}
